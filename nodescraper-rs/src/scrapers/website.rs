use std::time::Duration;

use dom_smoothie::{Config, Readability, TextMode};
use thirtyfour::{
    FirefoxCapabilities, WebDriver,
    common::capabilities::firefox::FirefoxPreferences,
};

use crate::messages::{
    self, response::ResponseType, response_error::ResponseError, response_ok::Response,
    trigger_scraper::ScraperArgs, website_scraper_error::WebsiteError,
};

const MAX_HTML_BYTES: usize = 4 << 20;
const PAGE_LOAD_TIMEOUT: Duration = Duration::from_secs(120);
const MOBILE_USER_AGENT: &str = "Mozilla/5.0 (iPhone; CPU iPhone OS 16_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.0 Mobile/15E148 Safari/604.1";

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

fn firefox_capabilities() -> Result<FirefoxCapabilities, String> {
    let mut prefs = FirefoxPreferences::new();
    prefs
        .set_user_agent(MOBILE_USER_AGENT.to_string())
        .map_err(|err| err.to_string())?;

    let mut caps = FirefoxCapabilities::new();
    caps.set_headless().map_err(|err| err.to_string())?;
    caps.set_preferences(prefs)
        .map_err(|err| err.to_string())?;
    Ok(caps)
}

async fn wait_for_document_ready(driver: &WebDriver) -> Result<(), String> {
    let deadline = tokio::time::Instant::now() + PAGE_LOAD_TIMEOUT;
    loop {
        let ready = driver
            .execute("return document.readyState", Vec::new())
            .await
            .map_err(|err| err.to_string())?;
        if ready.json().as_str() == Some("complete") {
            return Ok(());
        }
        if tokio::time::Instant::now() >= deadline {
            return Err("timed out waiting for page load".into());
        }
        tokio::time::sleep(Duration::from_millis(250)).await;
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

async fn scrape_website(source_url: &str) -> Result<(String, String, String), String> {
    let source_url = source_url.trim();
    if source_url.is_empty() {
        return Err("source url is empty".into());
    }

    let caps = firefox_capabilities()?;
    let driver = WebDriver::managed(caps)
        .await
        .map_err(|err| err.to_string())?;

    let scrape_result = async {
        driver.goto(source_url).await.map_err(|err| err.to_string())?;
        wait_for_document_ready(&driver).await?;
        let final_url = driver.current_url().await.map_err(|err| err.to_string())?;
        let html = driver.source().await.map_err(|err| err.to_string())?;
        if html.trim().is_empty() {
            return Err("rendered page source is empty".into());
        }
        let rendered_html = truncate_utf8(&html, MAX_HTML_BYTES);
        let content = extract_readable_content(&rendered_html, final_url.as_ref())?;
        Ok((final_url.to_string(), content, rendered_html))
    }
    .await;

    if let Err(err) = driver.quit().await {
        log::warn!("failed to quit firefox session: {err}");
    }

    scrape_result
}

pub async fn handle_website(
    command_id: u64,
    trigger: messages::TriggerScraper,
) -> messages::Response {
    let Some(ScraperArgs::WebsiteScraper(args)) = trigger.scraper_args else {
        return missing_field_response(command_id);
    };

    match scrape_website(&args.source_url).await {
        Ok((source_url, content, rendered_html)) => {
            website_ok_response(command_id, source_url, content, rendered_html)
        }
        Err(err) => {
            log::error!("website scrape failed for {}: {err}", args.source_url);
            website_scrape_error_response(command_id, err)
        }
    }
}
