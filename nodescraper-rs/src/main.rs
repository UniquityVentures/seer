use futures_util::{SinkExt, StreamExt};
use prost::Message;
use std::{sync::OnceLock, time::Duration};
use tokio::pin;
pub mod scrapers;

pub mod messages {
    include!(concat!(env!("OUT_DIR"), "/nodescraper.rs"));
}

static DEVICE_ID: OnceLock<u64> = OnceLock::new();

const VERSION: messages::VersionResponse = messages::VersionResponse {
    major: 0,
    minor: 0,
    patch: 1,
};

fn get_remote_url() -> Result<url::Url, url::ParseError> {
    let remote_url = match option_env!("REMOTE_URL") {
        Some(s) => s,
        None => "ws://localhost:42069/fleet/websocket",
    };
    let remote_url = url::Url::parse(remote_url)?;
    Ok(remote_url)
}

#[tokio::main]
async fn main() {
    env_logger::init();
    let remote_url = get_remote_url().expect("Error while getting the remote url");
    let scheme = remote_url.scheme();
    if scheme != "ws" && scheme != "wss" {
        panic!("Scheme for remote url should be ws or wss")
    }
    loop {
        tokio::time::sleep(Duration::new(1, 0)).await;
        let (ws_stream, _) = match tokio_tungstenite::connect_async(remote_url.clone()).await {
            Ok(conn) => conn,
            Err(e) => {
                log::error!("Error while connecting to remote {remote_url} {e}");
                continue;
            }
        };
        let (mut sink, stream) = ws_stream.split();

        let command_stream =
            stream.then(async |item| item.map(|item| messages::Command::decode(item.into_data())));
        pin!(command_stream);
        while let Some(command) = command_stream.next().await {
            let command = match command {
                Ok(c) => c,
                Err(e) => {
                    log::error!("Error while getting the data from websocket: {e}");
                    continue;
                }
            };
            let command = match command {
                Ok(c) => c,
                Err(e) => {
                    log::error!("Error while parsing data to Command format: {e}");
                    continue;
                }
            };
            let response = handle_command(command).await.encode_to_vec();
            match sink
                .send(tokio_tungstenite::tungstenite::Message::binary(response))
                .await
            {
                Ok(_) => {}
                Err(e) => {
                    log::error!("Error while sending back message to the server: {e}");
                    continue;
                }
            };
        }
    }
}

async fn handle_command(command: messages::Command) -> messages::Response {
    match command.command_type {
        Some(command_type) => match command_type {
            messages::command::CommandType::GetVersion(_) => messages::Response {
                command_id: command.id,
                response_type: Some(messages::response::ResponseType::Ok(messages::ResponseOk {
                    response: Some(messages::response_ok::Response::Version(VERSION)),
                })),
            },
            messages::command::CommandType::TriggerScraper(t) => {
                scrapers::website::handle_website(command.id, t).await
            }
            messages::command::CommandType::GetId(_) => messages::Response {
                command_id: command.id,
                response_type: Some(messages::response::ResponseType::Ok(messages::ResponseOk {
                    response: Some(messages::response_ok::Response::Id(messages::VersionId {
                        id: *DEVICE_ID.get_or_init(|| rand::random()),
                    })),
                })),
            },
        },
        None => messages::Response {
            command_id: command.id,
            response_type: Some(messages::response::ResponseType::Error(
                messages::ResponseError {
                    response_error: Some(
                        messages::response_error::ResponseError::MissingCommandFieldError(
                            messages::MissingCommandFieldError {},
                        ),
                    ),
                },
            )),
        },
    }
}
