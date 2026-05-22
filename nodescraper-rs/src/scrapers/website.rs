use std::{sync::OnceLock, time::Duration};

use thirtyfour::common::capabilities::firefox::FirefoxPreferences;
use thirtyfour::{FirefoxCapabilities, WebDriver};
use tokio::sync::Mutex;

use crate::messages::{
    self, response::ResponseType, response_error::ResponseError, response_ok::Response,
    trigger_scraper::ScraperArgs, website_scraper_error::WebsiteError,
};

const MAX_HTML_BYTES: usize = 4 << 20;
const NAVIGATE_TIMEOUT: Duration = Duration::from_secs(120);
const PAGE_SETTLE: Duration = Duration::from_millis(2_000);

struct BrowserSession {
    driver: WebDriver,
}

static BROWSER: OnceLock<Mutex<Option<BrowserSession>>> = OnceLock::new();

fn missing_field_response(command_id: u64) -> messages::Response {
    messages::Response {
        command_id,
        response_type: Some(ResponseType::Error(messages::ResponseError {
            response_error: Some(ResponseError::MissingCommandFieldError(
                messages::MissingCommandFieldError {},
            )),
        })),
    }
}

fn website_ok_response(command_id: u64, source_url: String, content: String) -> messages::Response {
    messages::Response {
        command_id,
        response_type: Some(ResponseType::Ok(messages::ResponseOk {
            response: Some(Response::WebsiteScraper(messages::WebsiteScraperResponse {
                source_url,
                content,
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

async fn firefox_capabilities() -> Result<FirefoxCapabilities, String> {
    let mut prefs = FirefoxPreferences::new();
    prefs
        .set_user_agent(
            "Mozilla/5.0 (iPhone; CPU iPhone OS 16_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.0 Mobile/15E148 Safari/604.1".to_string(),
        )
        .map_err(|err| err.to_string())?;

    let mut caps = FirefoxCapabilities::new();
    caps.set_headless().map_err(|err| err.to_string())?;
    caps.set_preferences(prefs).map_err(|err| err.to_string())?;
    Ok(caps)
}

async fn shared_driver() -> Result<(), String> {
    let lock = BROWSER.get_or_init(|| Mutex::new(None));
    let mut guard = lock.lock().await;
    if guard.is_some() {
        return Ok(());
    }

    let caps = firefox_capabilities().await?;
    let driver = WebDriver::managed(caps)
        .await
        .map_err(|err| format!("start firefox: {err}"))?;

    *guard = Some(BrowserSession { driver });
    Ok(())
}

async fn fetch_rendered_html(source_url: &str) -> Result<(String, String), String> {
    let source_url = source_url.trim();
    if source_url.is_empty() {
        return Err("source url is empty".into());
    }

    shared_driver().await?;

    let lock = BROWSER.get().expect("browser lock initialized");
    let guard = lock.lock().await;
    let driver = &guard
        .as_ref()
        .ok_or_else(|| "browser session missing".to_string())?
        .driver;

    let scrape = async {
        driver
            .goto(source_url)
            .await
            .map_err(|err| err.to_string())?;
        tokio::time::sleep(PAGE_SETTLE).await;
        let final_url = driver.current_url().await.map_err(|err| err.to_string())?;
        let mut html = driver.source().await.map_err(|err| err.to_string())?;
        if html.trim().is_empty() {
            return Err("browser returned empty page source".into());
        }
        if html.len() > MAX_HTML_BYTES {
            html.truncate(MAX_HTML_BYTES);
        }
        Ok((final_url.to_string(), html))
    };

    match tokio::time::timeout(NAVIGATE_TIMEOUT, scrape).await {
        Ok(result) => result,
        Err(_) => Err(format!("navigation timed out after {NAVIGATE_TIMEOUT:?}")),
    }
}

pub async fn handle_website(
    command_id: u64,
    trigger: messages::TriggerScraper,
) -> messages::Response {
    let Some(ScraperArgs::WebsiteScraper(args)) = trigger.scraper_args else {
        return missing_field_response(command_id);
    };

    match fetch_rendered_html(&args.source_url).await {
        Ok((source_url, content)) => website_ok_response(command_id, source_url, content),
        Err(err) => {
            log::error!("website scrape failed for {}: {err}", args.source_url);
            website_scrape_error_response(command_id, err)
        }
    }
}
