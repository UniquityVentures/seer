use std::{
    rc::Rc,
    sync::mpsc,
    thread::{self, JoinHandle},
    time::{Duration, Instant},
};

use dom_smoothie::{Config, Readability, TextMode};
use dpi::PhysicalSize;
use euclid::Scale;
use servo::{
    LoadStatus, Preferences, RenderingContext, Servo, ServoBuilder, SoftwareRenderingContext,
    WebDriverCommandMsg, WebDriverScriptCommand, WebView, WebViewBuilder, WebViewDelegate,
};
use servo_base::id::BrowsingContextId;
use url::Url;

use crate::messages::{
    self, response::ResponseType, response_error::ResponseError, response_ok::Response,
    website_scraper_error::WebsiteError,
};

use super::Scraper;

const MAX_HTML_BYTES: usize = 4 << 20;
const PAGE_LOAD_TIMEOUT: Duration = Duration::from_secs(120);
const MOBILE_USER_AGENT: &str = "Mozilla/5.0 (iPhone; CPU iPhone OS 16_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.0 Mobile/15E148 Safari/604.1";
const VIEWPORT_WIDTH: u32 = 390;
const VIEWPORT_HEIGHT: u32 = 844;

struct ScrapeRequest {
    source_url: String,
    reply: tokio::sync::oneshot::Sender<Result<(String, String, String), String>>,
}

struct ScrapeDelegate;

impl WebViewDelegate for ScrapeDelegate {
    fn notify_new_frame_ready(&self, webview: WebView) {
        webview.paint();
    }
}

struct ServoEngine {
    servo: Servo,
    webview: WebView,
    browsing_context_id: BrowsingContextId,
}

impl ServoEngine {
    fn new() -> Result<Self, String> {
        let size = PhysicalSize::new(VIEWPORT_WIDTH, VIEWPORT_HEIGHT);
        let rendering_context = Rc::new(
            SoftwareRenderingContext::new(size)
                .map_err(|err| format!("creating SoftwareRenderingContext: {err:?}"))?,
        );
        rendering_context
            .make_current()
            .map_err(|err| format!("making rendering context current: {err:?}"))?;

        let mut prefs = Preferences::default();
        prefs.user_agent = MOBILE_USER_AGENT.to_string();
        prefs.network_http_proxy_uri = String::new();
        prefs.network_https_proxy_uri = String::new();
        prefs.dom_intersection_observer_enabled = true;

        let servo = ServoBuilder::default().preferences(prefs).build();
        // Logging is initialized by the binary (`env_logger::init`); Servo's setup_logging
        // also calls log::set_logger and panics if one is already installed.

        let delegate: Rc<dyn WebViewDelegate> = Rc::new(ScrapeDelegate);
        let webview = WebViewBuilder::new(&servo, rendering_context)
            .url(
                Url::parse("about:blank")
                    .map_err(|err| format!("invalid about:blank url: {err}"))?,
            )
            .hidpi_scale_factor(Scale::new(1.0))
            .delegate(delegate)
            .build();

        let browsing_context_id = BrowsingContextId::from(webview.id());
        wait_until(
            &servo,
            || webview.url().is_some(),
            Duration::from_secs(30),
            "initial webview url not ready",
        )?;
        wait_until(
            &servo,
            || webview.load_status() == LoadStatus::Complete,
            Duration::from_secs(30),
            "initial webview load not complete",
        )?;

        Ok(Self {
            servo,
            webview,
            browsing_context_id,
        })
    }

    fn scrape(&mut self, source_url: &str) -> Result<(String, String, String), String> {
        let source_url = source_url.trim();
        if source_url.is_empty() {
            return Err("source url is empty".into());
        }

        let url = Url::parse(source_url).map_err(|err| format!("invalid url: {err}"))?;

        self.webview.load(url.clone());
        wait_for_navigation(&self.servo, &self.webview, &url, PAGE_LOAD_TIMEOUT)?;
        self.settle_rendering();

        let final_url = self.resolve_final_url(source_url)?;
        let html = self.get_page_source()?;
        if html.trim().is_empty() {
            return Err("rendered page source is empty".into());
        }

        let rendered_html = truncate_utf8(&html, MAX_HTML_BYTES);
        if rendered_html.len() < 128 {
            return Err("rendered page source is too small; navigation may have failed".into());
        }
        let content = extract_readable_content(&rendered_html, &final_url)?;
        Ok((final_url, content, rendered_html))
    }

    fn settle_rendering(&mut self) {
        for _ in 0..50 {
            pump_event_loop(&self.servo, Duration::from_millis(20));
            thread::sleep(Duration::from_millis(10));
        }
    }

    fn resolve_final_url(&mut self, source_url: &str) -> Result<String, String> {
        if let Ok(url) = self.get_url_with_timeout(Duration::from_secs(10)) {
            if is_usable_document_url(&url) {
                return Ok(url);
            }
        }

        if let Some(url) = self.webview.url() {
            let url = url.to_string();
            if is_usable_document_url(&url) {
                return Ok(url);
            }
        }

        Ok(source_url.to_string())
    }

    fn get_url_with_timeout(&mut self, timeout: Duration) -> Result<String, String> {
        let (sender, receiver) =
            servo_base::generic_channel::channel().ok_or("failed to create url channel")?;
        self.servo
            .execute_webdriver_command(WebDriverCommandMsg::ScriptCommand(
                self.browsing_context_id,
                WebDriverScriptCommand::GetUrl(sender),
            ));

        let deadline = Instant::now() + timeout;
        loop {
            if Instant::now() >= deadline {
                return Err("timed out waiting for get url".into());
            }
            match receiver.try_recv_timeout(Duration::from_millis(50)) {
                Ok(url) => return Ok(url),
                Err(servo_base::generic_channel::TryReceiveError::Empty) => {
                    pump_event_loop(&self.servo, Duration::from_millis(10));
                }
                Err(servo_base::generic_channel::TryReceiveError::ReceiveError(_)) => {
                    return Err("get url channel closed".into());
                }
            }
        }
    }

    fn get_page_source(&mut self) -> Result<String, String> {
        let (sender, receiver) =
            servo_base::generic_channel::channel().ok_or("failed to create page source channel")?;
        self.servo
            .execute_webdriver_command(WebDriverCommandMsg::ScriptCommand(
                self.browsing_context_id,
                WebDriverScriptCommand::GetPageSource(sender),
            ));

        let deadline = Instant::now() + PAGE_LOAD_TIMEOUT;
        loop {
            if Instant::now() >= deadline {
                return Err("timed out waiting for get page source".into());
            }
            match receiver.try_recv_timeout(Duration::from_millis(50)) {
                Ok(Ok(html)) => return Ok(html),
                Ok(Err(status)) => {
                    return Err(format!("get page source failed: {status:?}"));
                }
                Err(servo_base::generic_channel::TryReceiveError::Empty) => {
                    pump_event_loop(&self.servo, Duration::from_millis(10));
                }
                Err(servo_base::generic_channel::TryReceiveError::ReceiveError(_)) => {
                    return Err("get page source channel closed".into());
                }
            }
        }
    }
}

fn is_usable_document_url(url: &str) -> bool {
    match Url::parse(url) {
        Ok(parsed) => {
            matches!(parsed.scheme(), "http" | "https" | "file")
                && parsed != Url::parse("about:blank").expect("about:blank parses")
        }
        Err(_) => false,
    }
}

fn wait_for_navigation(
    servo: &Servo,
    webview: &WebView,
    target: &Url,
    timeout: Duration,
) -> Result<(), String> {
    let deadline = Instant::now() + timeout;
    let mut saw_started = false;
    while Instant::now() < deadline {
        match webview.load_status() {
            LoadStatus::Started | LoadStatus::HeadParsed => saw_started = true,
            LoadStatus::Complete if saw_started => {
                if navigation_reached_target(webview, target) {
                    return Ok(());
                }
            }
            LoadStatus::Complete => {}
        }

        if saw_started
            && webview.load_status() == LoadStatus::Complete
            && navigation_reached_target(webview, target)
        {
            return Ok(());
        }

        pump_event_loop(servo, Duration::from_millis(10));
        thread::sleep(Duration::from_millis(1));
    }

    Err("timed out waiting for page load".into())
}

fn navigation_reached_target(webview: &WebView, target: &Url) -> bool {
    let Some(current) = webview.url() else {
        return false;
    };
    if current == *target {
        return true;
    }
    match (current.host_str(), target.host_str()) {
        (Some(current_host), Some(target_host)) => current_host == target_host,
        _ => false,
    }
}

fn wait_until<F>(
    servo: &Servo,
    mut condition: F,
    timeout: Duration,
    error: &str,
) -> Result<(), String>
where
    F: FnMut() -> bool,
{
    let deadline = Instant::now() + timeout;
    while !condition() {
        if Instant::now() >= deadline {
            return Err(error.into());
        }
        pump_event_loop(servo, Duration::from_millis(10));
        thread::sleep(Duration::from_millis(1));
    }
    Ok(())
}

fn pump_event_loop(servo: &Servo, duration: Duration) {
    let deadline = Instant::now() + duration;
    while Instant::now() < deadline {
        servo.spin_event_loop();
        thread::sleep(Duration::from_millis(1));
    }
}

fn truncate_utf8(input: &str, max_bytes: usize) -> String {
    if input.len() <= max_bytes {
        return input.to_string();
    }
    let mut end = max_bytes;
    while end > 0 && !input.is_char_boundary(end) {
        end -= 1;
    }
    input[..end].to_string()
}

fn website_ok_response(
    command_id: u64,
    source_url: String,
    content: String,
    rendered_html: String,
) -> messages::Response {
    messages::Response {
        command_id,
        response_type: Some(ResponseType::Ok(messages::ResponseOk {
            response: Some(Response::WebsiteScraper(messages::WebsiteScraperResponse {
                source_url,
                content,
                rendered_html,
            })),
        })),
    }
}

fn website_scrape_error_response(command_id: u64, message: String) -> messages::Response {
    messages::Response {
        command_id,
        response_type: Some(ResponseType::Error(messages::ResponseError {
            response_error: Some(ResponseError::WebsiteScraperError(
                messages::WebsiteScraperError {
                    website_error: Some(WebsiteError::ScrapeFailed(messages::ScrapeFailed {
                        message,
                    })),
                },
            )),
        })),
    }
}

fn extract_readable_content(html: &str, document_url: &str) -> Result<String, String> {
    let cfg = Config {
        text_mode: TextMode::Markdown,
        max_elements_to_parse: 9000,
        ..Default::default()
    };
    let mut readability =
        Readability::new(html, Some(document_url), Some(cfg)).map_err(|err| err.to_string())?;
    let article = readability.parse().map_err(|err| err.to_string())?;
    let content = article.text_content.trim().to_string();
    if content.is_empty() {
        return Err("no extractable readable content".into());
    }
    Ok(content)
}

fn run_servo_thread(request_rx: mpsc::Receiver<ScrapeRequest>) {
    let mut engine = match ServoEngine::new() {
        Ok(engine) => engine,
        Err(err) => {
            log::error!("failed to initialize servo engine: {err}");
            while let Ok(request) = request_rx.recv() {
                let _ = request.reply.send(Err(err.clone()));
            }
            return;
        }
    };

    while let Ok(request) = request_rx.recv() {
        let result = engine.scrape(&request.source_url);
        let _ = request.reply.send(result);
    }
}

pub struct WebsiteScraper {
    request_tx: mpsc::Sender<ScrapeRequest>,
    _servo_thread: JoinHandle<()>,
}

impl Scraper for WebsiteScraper {
    type Error = String;
    type Args = crate::messages::WebsiteScraperArgs;

    fn init() -> Result<Self, Self::Error> {
        let (request_tx, request_rx) = mpsc::channel();
        let servo_thread = thread::Builder::new()
            .name("servo-scraper".into())
            .spawn(move || run_servo_thread(request_rx))
            .map_err(|err| format!("failed to spawn servo thread: {err}"))?;

        Ok(Self {
            request_tx,
            _servo_thread: servo_thread,
        })
    }

    async fn scrape(&self, command_id: u64, trigger: Self::Args) -> messages::Response {
        let (reply_tx, reply_rx) = tokio::sync::oneshot::channel();
        let request = ScrapeRequest {
            source_url: trigger.source_url.clone(),
            reply: reply_tx,
        };

        if self.request_tx.send(request).is_err() {
            let message = "servo scraper thread is unavailable".into();
            log::error!(
                "website scrape failed for {}: {message}",
                trigger.source_url
            );
            return website_scrape_error_response(command_id, message);
        }

        match reply_rx.await {
            Ok(result) => match result {
                Ok((source_url, content, rendered_html)) => {
                    website_ok_response(command_id, source_url, content, rendered_html)
                }
                Err(err) => {
                    log::error!("website scrape failed for {}: {err}", trigger.source_url);
                    website_scrape_error_response(command_id, err)
                }
            },
            Err(_) => {
                let message = "servo scraper thread dropped the response".into();
                log::error!(
                    "website scrape failed for {}: {message}",
                    trigger.source_url
                );
                website_scrape_error_response(command_id, message)
            }
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn scrape_example_com() {
        // Servo opts/script init are process-global; run scrape on this thread like Servo's tests.
        let mut engine = ServoEngine::new().expect("init servo engine");
        let (source_url, content, rendered_html) = engine
            .scrape("https://example.com")
            .expect("scrape https://example.com should succeed");
        assert!(
            source_url.contains("example.com"),
            "unexpected source url: {source_url}"
        );
        assert!(!content.is_empty(), "content should not be empty");
        assert!(!rendered_html.is_empty(), "html should not be empty");
    }
}
