//! mongosql-lsp — Language Server Protocol server for `.mir` and `.air` debug files.
//!
//! Reads from stdin and writes to stdout (JSON-RPC over stdio), making it compatible
//! with any LSP-capable editor.  Start it with the VS Code extension or configure it
//! directly in Neovim, Helix, etc.
//!
//! # Usage
//!
//! ```text
//! cargo build -p mongosql-lsp
//! # The VS Code extension launches this automatically via SERVER_PATH.
//! ```

mod hover;
mod parser;
mod server;

use server::Backend;
use tower_lsp::{LspService, Server};

#[tokio::main]
async fn main() {
    env_logger::init();

    let stdin = tokio::io::stdin();
    let stdout = tokio::io::stdout();

    let (service, socket) = LspService::build(Backend::new).finish();

    Server::new(stdin, stdout, socket).serve(service).await;
}
