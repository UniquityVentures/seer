use std::{
    panic,
    rc::Rc,
    sync::{
        Arc, Mutex,
        atomic::{AtomicBool, Ordering},
    },
    thread::{self, JoinHandle},
    time::{Duration, Instant},
};

use crossbeam::channel::{Receiver, SendError, Sender, select, unbounded};

use dom_smoothie::{Config, Readability, TextMode};
use dpi::PhysicalSize;
use euclid::Scale;
use servo::{
    EventLoopWaker, LoadStatus, Preferences, RenderingContext, Servo, ServoBuilder,
    SoftwareRenderingContext, WebDriverCommandMsg, WebDriverScriptCommand, WebView, WebViewBuilder,
    WebViewDelegate,
};
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
const WORKER_RESTART_DELAY: Duration = Duration::from_millis(250);
const SERVO_UNAVAILABLE: &str = "servo scraper unavailable";
const COOKIE_BANNER_TIMEOUT: Duration = Duration::from_secs(10);

/// XPath to find "Reject All" buttons or links, case-insensitive.
/// Uses translate() for case folding since XPath 1.0 lacks lower-case().
const REJECT_BUTTON_XPATH: &str = concat!(
    "//*[self::button or self::a or self::input[@type='button'] or self::input[@type='submit'] ",
    "or self::div[@role='button'] or self::span[@role='button']][contains(",
    "translate(., 'ABCDEFGHIJKLMNOPQRSTUVWXYZ', 'abcdefghijklmnopqrstuvwxyz'), 'reject all')]",
);

/// Servo initializes process-global options once; only one engine can exist per process.
static SERVO_INITIALIZED: AtomicBool = AtomicBool::new(false);

struct ScrapeDelegate;

impl WebViewDelegate for ScrapeDelegate {
    fn notify_new_frame_ready(&self, webview: WebView) {
        webview.paint();
    }
}

struct ServoEngine {
    servo: Servo,
    rendering_context: Rc<SoftwareRenderingContext>,
}

#[derive(Clone)]
struct ScraperEventLoopWaker {
    sender: Sender<()>,
}

impl EventLoopWaker for ScraperEventLoopWaker {
    fn clone_box(&self) -> Box<dyn EventLoopWaker> {
        Box::new(self.clone())
    }

    fn wake(&self) {
        self.sender
            .send(())
            .expect("To be able to send the signal to wake the servo engine");
    }
}

impl ServoEngine {
    fn new() -> Result<(Self, Receiver<()>), String> {
        let size = PhysicalSize::new(VIEWPORT_WIDTH, VIEWPORT_HEIGHT);
        let (waker_sender, waker_receiver) = unbounded::<()>();
        let event_loop_waker = Box::new(ScraperEventLoopWaker {
            sender: waker_sender,
        });
        let rendering_context = Rc::new(
            SoftwareRenderingContext::new(size)
                .map_err(|err| format!("creating SoftwareRenderingContext: {err:?}"))?,
        );
        rendering_context
            .make_current()
            .map_err(|err| format!("making rendering context current: {err:?}"))?;

        let prefs = Preferences {
            user_agent: MOBILE_USER_AGENT.to_string(),
            network_http_proxy_uri: String::new(),
            network_https_proxy_uri: String::new(),
            dom_intersection_observer_enabled: true,
            ..Default::default()
        };

        let servo = ServoBuilder::default()
            .preferences(prefs)
            .event_loop_waker(event_loop_waker)
            .build();

        // Logging is initialized by the binary (`env_logger::init`); Servo's setup_logging
        // also calls log::set_logger and panics if one is already installed.
        Ok((
            Self {
                servo,
                rendering_context,
            },
            waker_receiver,
        ))
    }

    fn scrape(
        &mut self,
        source_url: Url,
        waker: &Receiver<()>,
    ) -> Result<(String, String, String), String> {
        let delegate: Rc<dyn WebViewDelegate> = Rc::new(ScrapeDelegate);
        let webview = WebViewBuilder::new(&self.servo, self.rendering_context.clone())
            .url(
                Url::parse("about:blank")
                    .map_err(|err| format!("invalid about:blank url: {err}"))?,
            )
            .hidpi_scale_factor(Scale::new(1.0))
            .delegate(delegate)
            .build();

        wait_until(
            &self.servo,
            waker,
            || webview.url().is_some(),
            Duration::from_secs(30),
            "initial webview url not ready",
        )?;
        wait_until(
            &self.servo,
            waker,
            || webview.load_status() == LoadStatus::Complete,
            Duration::from_secs(30),
            "initial webview load not complete",
        )?;

        webview.load(source_url.clone());
        wait_for_navigation(&self.servo, waker, &webview, &source_url, PAGE_LOAD_TIMEOUT)?;

        self.settle_rendering(waker);

        // Try to dismiss cookie consent banners before capturing page source.
        self.try_dismiss_cookie_banner(&webview, waker);

        let final_url = self.resolve_final_url(&source_url, &webview, waker)?;
        let html = self.get_page_source(&webview, waker)?;
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

    fn settle_rendering(&mut self, waker: &Receiver<()>) {
        // Give Servo time to finish any remaining rendering work by pumping
        // the event loop for a fixed duration, driven by waker signals.
        let start = Instant::now();
        let deadline = start + Duration::from_secs(5);
        while Instant::now() < deadline {
            pump_event_loop(&self.servo, waker, Duration::from_millis(50));
        }
    }

    fn resolve_final_url(
        &mut self,
        source_url: &Url,
        webview: &WebView,
        waker: &Receiver<()>,
    ) -> Result<String, String> {
        if let Ok(url) = self.get_url_with_timeout(Duration::from_secs(10), webview, waker)
            && is_usable_document_url(&url)
        {
            return Ok(url);
        }

        if let Some(url) = webview.url() {
            let url = url.to_string();
            if is_usable_document_url(&url) {
                return Ok(url);
            }
        }

        Ok(source_url.to_string())
    }

    fn get_url_with_timeout(
        &mut self,
        timeout: Duration,
        webview: &WebView,
        waker: &Receiver<()>,
    ) -> Result<String, String> {
        let (sender, receiver) =
            servo_base::generic_channel::channel().ok_or("failed to create url channel")?;

        self.servo
            .execute_webdriver_command(WebDriverCommandMsg::ScriptCommand(
                webview.id().into(),
                WebDriverScriptCommand::GetUrl(sender),
            ));

        let start = Instant::now();
        let deadline = start + timeout;
        loop {
            if Instant::now() >= deadline {
                return Err("timed out waiting for get url".into());
            }
            match receiver.try_recv_timeout(Duration::from_millis(50)) {
                Ok(url) => {
                    return Ok(url);
                }
                Err(servo_base::generic_channel::TryReceiveError::Empty) => {
                    pump_event_loop(&self.servo, waker, Duration::from_millis(50));
                }
                Err(servo_base::generic_channel::TryReceiveError::ReceiveError(_)) => {
                    return Err("get url channel closed".into());
                }
            }
        }
    }

    fn get_page_source(
        &mut self,
        webview: &WebView,
        waker: &Receiver<()>,
    ) -> Result<String, String> {
        let (sender, receiver) =
            servo_base::generic_channel::channel().ok_or("failed to create page source channel")?;

        self.servo
            .execute_webdriver_command(WebDriverCommandMsg::ScriptCommand(
                webview.id().into(),
                WebDriverScriptCommand::GetPageSource(sender),
            ));

        let start = Instant::now();
        let deadline = start + PAGE_LOAD_TIMEOUT;
        loop {
            if Instant::now() >= deadline {
                return Err("timed out waiting for get page source".into());
            }
            match receiver.try_recv_timeout(Duration::from_millis(50)) {
                Ok(Ok(html)) => {
                    return Ok(html);
                }
                Ok(Err(status)) => {
                    return Err(format!("get page source failed: {status:?}"));
                }
                Err(servo_base::generic_channel::TryReceiveError::Empty) => {
                    pump_event_loop(&self.servo, waker, Duration::from_millis(50));
                }
                Err(servo_base::generic_channel::TryReceiveError::ReceiveError(_)) => {
                    return Err("get page source channel closed".into());
                }
            }
        }
    }

    /// Try to find and click a "Reject All" cookie consent button.
    ///
    /// This is best-effort: if no button is found or the click fails, the scrape
    /// continues with the page as-is. After a successful click, the event loop is
    /// pumped to let the page settle (banner animation/removal, DOM updates).
    fn try_dismiss_cookie_banner(&mut self, webview: &WebView, waker: &Receiver<()>) {
        let elements = match self.find_elements_xpath(webview, REJECT_BUTTON_XPATH, waker) {
            Ok(ids) if !ids.is_empty() => ids,
            Ok(_) => {
                return;
            }
            Err(_) => {
                return;
            }
        };

        // Click the first matching element.
        let element_id = &elements[0];
        match self.click_element(webview, element_id, waker) {
            Ok(_) => {
                self.settle_rendering(waker);
            }
            Err(_) => {}
        }
    }

    /// Find elements matching an XPath selector in the given webview.
    fn find_elements_xpath(
        &mut self,
        webview: &WebView,
        xpath: &str,
        waker: &Receiver<()>,
    ) -> Result<Vec<String>, String> {
        let (sender, receiver) = servo_base::generic_channel::channel()
            .ok_or("failed to create find elements channel")?;

        self.servo
            .execute_webdriver_command(WebDriverCommandMsg::ScriptCommand(
                webview.id().into(),
                WebDriverScriptCommand::FindElementsXpathSelector(xpath.to_string(), sender),
            ));

        let start = Instant::now();
        let deadline = start + COOKIE_BANNER_TIMEOUT;
        loop {
            if Instant::now() >= deadline {
                return Err("timed out waiting for find elements".into());
            }
            match receiver.try_recv_timeout(Duration::from_millis(50)) {
                Ok(Ok(ids)) => {
                    return Ok(ids);
                }
                Ok(Err(status)) => {
                    return Err(format!("find elements failed: {status:?}"));
                }
                Err(servo_base::generic_channel::TryReceiveError::Empty) => {
                    pump_event_loop(&self.servo, waker, Duration::from_millis(50));
                }
                Err(servo_base::generic_channel::TryReceiveError::ReceiveError(_)) => {
                    return Err("find elements channel closed".into());
                }
            }
        }
    }

    /// Click an element by its WebDriver element ID.
    fn click_element(
        &mut self,
        webview: &WebView,
        element_id: &str,
        waker: &Receiver<()>,
    ) -> Result<(), String> {
        let (sender, receiver) = servo_base::generic_channel::channel()
            .ok_or("failed to create click element channel")?;

        self.servo
            .execute_webdriver_command(WebDriverCommandMsg::ScriptCommand(
                webview.id().into(),
                WebDriverScriptCommand::ElementClick(element_id.to_string(), sender),
            ));

        let start = Instant::now();
        let deadline = start + COOKIE_BANNER_TIMEOUT;
        loop {
            if Instant::now() >= deadline {
                return Err("timed out waiting for element click".into());
            }
            match receiver.try_recv_timeout(Duration::from_millis(50)) {
                Ok(Ok(_)) => {
                    return Ok(());
                }
                Ok(Err(status)) => {
                    return Err(format!("element click failed: {status:?}"));
                }
                Err(servo_base::generic_channel::TryReceiveError::Empty) => {
                    pump_event_loop(&self.servo, waker, Duration::from_millis(50));
                }
                Err(servo_base::generic_channel::TryReceiveError::ReceiveError(_)) => {
                    return Err("element click channel closed".into());
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
    waker: &Receiver<()>,
    webview: &WebView,
    target: &Url,
    timeout: Duration,
) -> Result<(), String> {
    let start = Instant::now();
    let deadline = start + timeout;
    while Instant::now() < deadline {
        if webview.load_status() == LoadStatus::Complete
            && navigation_reached_target(webview, target)
        {
            return Ok(());
        }
        pump_event_loop(servo, waker, Duration::from_millis(100));
    }

    Err("timed out waiting for page load".into())
}

fn navigation_reached_target(webview: &WebView, _target: &Url) -> bool {
    let Some(current) = webview.url() else {
        return false;
    };
    // Ensure we have left the about:blank screen and are on a usable page
    is_usable_document_url(&current.to_string())
}

fn wait_until<F>(
    servo: &Servo,
    waker: &Receiver<()>,
    mut condition: F,
    timeout: Duration,
    error: &str,
) -> Result<(), String>
where
    F: FnMut() -> bool,
{
    let start = Instant::now();
    let deadline = start + timeout;
    while !condition() {
        if Instant::now() >= deadline {
            return Err(error.into());
        }
        pump_event_loop(servo, waker, Duration::from_millis(100));
    }
    Ok(())
}

/// Pump Servo's event loop by waiting for waker signals.
///
/// Blocks on the waker channel up to `timeout`. Each received signal triggers
/// one `spin_event_loop` call. If no signal arrives within the timeout, a
/// single spin is performed as a fallback to avoid stalling on missed wakes.
fn pump_event_loop(servo: &Servo, waker: &Receiver<()>, timeout: Duration) {
    // Block until the first waker signal or timeout.
    match waker.recv_timeout(timeout) {
        Ok(()) => {
            servo.spin_event_loop();
        }
        Err(crossbeam::channel::RecvTimeoutError::Timeout) => {
            // No wake signal within the timeout — do a single defensive spin
            // to handle any work that might have been queued without a wake.
            servo.spin_event_loop();
            return;
        }
        Err(crossbeam::channel::RecvTimeoutError::Disconnected) => {
            log::warn!("pump_event_loop: waker channel disconnected");
            return;
        }
    }

    // Drain any additional signals that accumulated while we were spinning,
    // so we process all pending work in one batch.
    while waker.try_recv().is_ok() {
        servo.spin_event_loop();
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

struct ScrapeRequest {
    command_id: u64,
    url: Url,
    reply: tokio::sync::oneshot::Sender<messages::Response>,
}

fn run_servo_worker(request_rx: Receiver<ScrapeRequest>) {
    let (mut engine, event_waker) = match ServoEngine::new() {
        Ok(engine) => {
            SERVO_INITIALIZED.store(true, Ordering::Release);
            engine
        }
        Err(err) => {
            log::error!("servo worker failed to initialize: {err}");
            drain_requests_with_error(request_rx, err);
            return;
        }
    };

    log::info!("servo worker: entering event loop");
    loop {
        select! {
            recv(event_waker) -> _ => {
                log::trace!("servo worker: waker signal, spinning event loop");
                engine.servo.spin_event_loop();
            },
            recv(request_rx) -> request => {
                let request = if let Ok(request) = request {
                    request
                } else {
                    log::info!("servo worker: request channel closed, exiting");
                    break;
                };
                let ScrapeRequest {
                    command_id,
                    url,
                    reply,
                } = request;

                log::info!("servo worker: received scrape request #{command_id} for {url}");
                let scrape_start = Instant::now();
                let response = match panic::catch_unwind(panic::AssertUnwindSafe(|| {
                    engine.scrape(url, &event_waker)
                })) {
                    Ok(Ok((source_url, content, rendered_html))) => {
                        log::info!(
                            "servo worker: scrape #{command_id} succeeded in {:.2}s",
                            scrape_start.elapsed().as_secs_f64()
                        );
                        website_ok_response(command_id, source_url, content, rendered_html)
                    }
                    Ok(Err(err)) => {
                        log::error!(
                            "servo worker: scrape #{command_id} failed in {:.2}s: {err}",
                            scrape_start.elapsed().as_secs_f64()
                        );
                        website_scrape_error_response(command_id, err)
                    }
                    Err(_) => {
                        log::error!("servo worker panicked during scrape #{command_id}");
                        website_scrape_error_response(
                            command_id,
                            "servo worker panicked during scrape".into(),
                        )
                    }
                };
                let _ = reply.send(response);
            }
        };
    }
}

fn drain_requests_with_error(request_rx: Receiver<ScrapeRequest>, message: String) {
    while let Ok(request) = request_rx.recv() {
        let _ = request.reply.send(website_scrape_error_response(
            request.command_id,
            message.clone(),
        ));
    }
}

fn run_servo_supervisor(request_tx: Arc<Mutex<Sender<ScrapeRequest>>>) {
    let mut permanently_failed = false;

    loop {
        let (tx, rx) = unbounded();
        {
            let mut sender = request_tx
                .lock()
                .unwrap_or_else(|poisoned| poisoned.into_inner());
            *sender = tx;
        }

        if permanently_failed {
            drain_requests_with_error(rx, SERVO_UNAVAILABLE.into());
            break;
        }

        let worker = match thread::Builder::new()
            .name("servo-scraper".into())
            .stack_size(32 * 1024 * 1024)
            .spawn(move || run_servo_worker(rx))
        {
            Ok(worker) => worker,
            Err(err) => {
                log::error!("failed to spawn servo worker: {err}");
                thread::sleep(WORKER_RESTART_DELAY);
                continue;
            }
        };

        match worker.join() {
            Ok(()) => {
                log::info!("servo worker stopped");
                break;
            }
            Err(_) => {
                if SERVO_INITIALIZED.load(Ordering::Acquire) {
                    log::error!(
                        "servo worker panicked after initialization; cannot restart Servo in-process"
                    );
                    permanently_failed = true;
                } else {
                    log::error!("servo worker panicked before initialization, restarting");
                    thread::sleep(WORKER_RESTART_DELAY);
                }
            }
        }
    }
}

struct ServoThread {
    request_tx: Arc<Mutex<Sender<ScrapeRequest>>>,
    _supervisor: JoinHandle<()>,
}

impl ServoThread {
    fn start() -> Result<Self, String> {
        let (request_tx, _request_rx) = unbounded();
        let request_tx = Arc::new(Mutex::new(request_tx));
        let supervisor = thread::Builder::new()
            .name("servo-scraper-supervisor".into())
            .spawn({
                let request_tx = Arc::clone(&request_tx);
                move || run_servo_supervisor(request_tx)
            })
            .map_err(|err| format!("failed to spawn servo supervisor: {err}"))?;

        log::info!("started servo scraper thread");

        Ok(Self {
            request_tx,
            _supervisor: supervisor,
        })
    }

    fn dispatch(&self, request: ScrapeRequest) -> Result<(), SendError<ScrapeRequest>> {
        let sender = self
            .request_tx
            .lock()
            .unwrap_or_else(|poisoned| poisoned.into_inner());
        sender.send(request)
    }
}

pub struct WebsiteScraper {
    servo: ServoThread,
}

impl Scraper for WebsiteScraper {
    type Error = String;
    type Args = crate::messages::WebsiteScraperArgs;

    fn init() -> Result<Self, Self::Error> {
        Ok(Self {
            servo: ServoThread::start()?,
        })
    }

    async fn scrape(&self, command_id: u64, trigger: Self::Args) -> messages::Response {
        let url = match Url::parse(&trigger.source_url) {
            Ok(url) => url,
            Err(err) => {
                log::error!("website scrape failed for {}: {err}", trigger.source_url);
                return website_scrape_error_response(command_id, err.to_string());
            }
        };

        let (reply_tx, reply_rx) = tokio::sync::oneshot::channel();
        let request = ScrapeRequest {
            command_id,
            url,
            reply: reply_tx,
        };

        if self.servo.dispatch(request).is_err() {
            let message = SERVO_UNAVAILABLE.to_string();
            log::error!(
                "website scrape failed for {}: {message}",
                trigger.source_url
            );
            return website_scrape_error_response(command_id, message);
        }

        match reply_rx.await {
            Ok(response) => response,
            Err(err) => {
                log::error!("website scrape failed for {}: {err}", trigger.source_url);
                website_scrape_error_response(command_id, err.to_string())
            }
        }
    }
}
