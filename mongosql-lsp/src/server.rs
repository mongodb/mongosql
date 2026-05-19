//! LSP backend for `.mir` and `.air` debug-tree files.

use crop::Rope;
use dashmap::DashMap;
use tower_lsp::jsonrpc::Result;
use tower_lsp::lsp_types::{
    DidChangeTextDocumentParams, DidCloseTextDocumentParams, DidOpenTextDocumentParams,
    DidSaveTextDocumentParams, FoldingRange, FoldingRangeParams, FoldingRangeProviderCapability,
    Hover, HoverContents, HoverParams, HoverProviderCapability, InitializeParams, InitializeResult,
    InitializedParams, MarkupContent, MarkupKind, MessageType, Position, SaveOptions,
    SemanticToken, SemanticTokenType, SemanticTokens, SemanticTokensFullOptions,
    SemanticTokensLegend, SemanticTokensOptions, SemanticTokensParams, SemanticTokensResult,
    SemanticTokensServerCapabilities, ServerCapabilities, TextDocumentSyncCapability,
    TextDocumentSyncKind, TextDocumentSyncOptions, TextDocumentSyncSaveOptions, Url,
};
use tower_lsp::{Client, LanguageServer};

use crate::hover::HOVER_MAP;
use crate::parser::{self, ParsedDoc};

// ── Semantic token legend ─────────────────────────────────────────────────────

/// Token type names registered in `initialize`.
///
/// Index mapping:
/// - 0 `type`     — enum-variant / struct names
/// - 1 `property` — struct field keys
/// - 2 `string`   — string literals
/// - 3 `number`   — numeric literals
/// - 4 `keyword`  — `true`, `false`, `None`, `Some`
pub const TOKEN_TYPES: &[&str] = &["type", "property", "string", "number", "keyword"];

// ── Backend ───────────────────────────────────────────────────────────────────

/// LSP backend state.
pub struct Backend {
    client: Client,
    /// Stores rope-encoded document text, keyed by URI string.
    document_map: DashMap<String, Rope>,
    /// Stores the parsed CST for each open document.
    ast_map: DashMap<String, ParsedDoc>,
}

impl Backend {
    /// Creates a new backend bound to `client`.
    pub fn new(client: Client) -> Self {
        Self {
            client,
            document_map: DashMap::new(),
            ast_map: DashMap::new(),
        }
    }

    /// Re-parse the document whenever it is opened or changed.
    fn on_change(&self, uri: &Url, text: &str) {
        let rope = Rope::from(text);
        let doc = parser::parse(text);
        self.document_map.insert(uri.to_string(), rope);
        self.ast_map.insert(uri.to_string(), doc);
    }
}

// ── LanguageServer impl ───────────────────────────────────────────────────────

#[tower_lsp::async_trait]
impl LanguageServer for Backend {
    async fn initialize(&self, _: InitializeParams) -> Result<InitializeResult> {
        Ok(InitializeResult {
            capabilities: ServerCapabilities {
                text_document_sync: Some(TextDocumentSyncCapability::Options(
                    TextDocumentSyncOptions {
                        open_close: Some(true),
                        change: Some(TextDocumentSyncKind::FULL),
                        save: Some(TextDocumentSyncSaveOptions::SaveOptions(SaveOptions {
                            include_text: Some(true),
                        })),
                        ..Default::default()
                    },
                )),
                semantic_tokens_provider: Some(
                    SemanticTokensServerCapabilities::SemanticTokensOptions(
                        SemanticTokensOptions {
                            legend: SemanticTokensLegend {
                                token_types: TOKEN_TYPES
                                    .iter()
                                    .map(|s| SemanticTokenType::new(s))
                                    .collect(),
                                token_modifiers: vec![],
                            },
                            full: Some(SemanticTokensFullOptions::Bool(true)),
                            range: None,
                            ..Default::default()
                        },
                    ),
                ),
                folding_range_provider: Some(FoldingRangeProviderCapability::Simple(true)),
                hover_provider: Some(HoverProviderCapability::Simple(true)),
                ..Default::default()
            },
            ..Default::default()
        })
    }

    async fn initialized(&self, _: InitializedParams) {
        self.client
            .log_message(MessageType::INFO, "mongosql-lsp initialized")
            .await;
    }

    async fn shutdown(&self) -> Result<()> {
        Ok(())
    }

    async fn did_open(&self, params: DidOpenTextDocumentParams) {
        self.on_change(&params.text_document.uri, &params.text_document.text);
    }

    async fn did_change(&self, params: DidChangeTextDocumentParams) {
        // Full-sync mode: there is always exactly one content change.
        if let Some(change) = params.content_changes.into_iter().next() {
            self.on_change(&params.text_document.uri, &change.text);
        }
    }

    async fn did_save(&self, params: DidSaveTextDocumentParams) {
        if let Some(text) = params.text {
            self.on_change(&params.text_document.uri, &text);
        }
    }

    async fn did_close(&self, params: DidCloseTextDocumentParams) {
        let uri = params.text_document.uri.to_string();
        self.document_map.remove(&uri);
        self.ast_map.remove(&uri);
    }

    async fn semantic_tokens_full(
        &self,
        params: SemanticTokensParams,
    ) -> Result<Option<SemanticTokensResult>> {
        let uri = params.text_document.uri.to_string();
        let tokens = (|| -> Option<Vec<SemanticToken>> {
            let rope = self.document_map.get(&uri)?;
            let doc = self.ast_map.get(&uri)?;
            let spans = parser::collect_tokens(&doc.root, &doc.text);
            Some(encode_tokens(&spans, &rope, &doc.text))
        })();

        Ok(tokens.map(|data| {
            SemanticTokensResult::Tokens(SemanticTokens {
                result_id: None,
                data,
            })
        }))
    }

    async fn folding_range(&self, params: FoldingRangeParams) -> Result<Option<Vec<FoldingRange>>> {
        let uri = params.text_document.uri.to_string();
        let ranges = (|| -> Option<Vec<FoldingRange>> {
            let rope = self.document_map.get(&uri)?;
            let doc = self.ast_map.get(&uri)?;
            let spans = parser::collect_fold_ranges(&doc.root, &doc.text);
            let mut result = Vec::new();
            for (start_byte, end_byte) in spans {
                let start_pos = byte_to_position(start_byte, &rope, &doc.text);
                let end_pos = byte_to_position(end_byte, &rope, &doc.text);
                // Only emit folding ranges that span at least two lines.
                if end_pos.line > start_pos.line {
                    result.push(FoldingRange {
                        start_line: start_pos.line,
                        start_character: Some(start_pos.character),
                        end_line: end_pos.line,
                        end_character: Some(end_pos.character),
                        kind: None,
                        collapsed_text: None,
                    });
                }
            }
            Some(result)
        })();
        Ok(ranges)
    }

    async fn hover(&self, params: HoverParams) -> Result<Option<Hover>> {
        let uri = params
            .text_document_position_params
            .text_document
            .uri
            .to_string();
        let pos = params.text_document_position_params.position;

        // Inline IIFE so `?` works cleanly inside.
        let hover = (|| -> Option<Hover> {
            let rope = self.document_map.get(&uri)?;
            let doc = self.ast_map.get(&uri)?;
            let offset = position_to_byte_offset(pos, &rope, &doc.text);
            let name = doc.node_name_at(offset)?;
            let desc = HOVER_MAP.get(name)?;
            Some(Hover {
                contents: HoverContents::Markup(MarkupContent {
                    kind: MarkupKind::Markdown,
                    value: (*desc).to_owned(),
                }),
                range: None,
            })
        })();
        Ok(hover)
    }
}

// ── Helpers ───────────────────────────────────────────────────────────────────

/// Convert a byte offset into a `(line, character)` LSP `Position`.
#[expect(
    clippy::cast_possible_truncation,
    reason = "LSP positions are u32; documents in practice never approach 4 GiB"
)]
fn byte_to_position(offset: usize, rope: &Rope, src: &str) -> Position {
    let offset = offset.min(src.len());
    let text_before = &src[..offset];
    let line = text_before.bytes().filter(|&b| b == b'\n').count() as u32;
    let last_nl = text_before.rfind('\n').map_or(0, |i| i + 1);
    let col_str = &text_before[last_nl..];
    // LSP character offset is in UTF-16 code units.
    let character = col_str.encode_utf16().count() as u32;
    // suppress unused warning for rope param (kept for API symmetry)
    let _ = rope;
    Position { line, character }
}

/// Convert an LSP `Position` into a byte offset in the source text.
#[expect(
    clippy::cast_possible_truncation,
    reason = "LSP positions are u32; documents in practice never approach 4 GiB"
)]
fn position_to_byte_offset(pos: Position, _rope: &Rope, src: &str) -> usize {
    let mut line = 0u32;
    let mut byte = 0usize;
    for ch in src.chars() {
        if line == pos.line {
            break;
        }
        if ch == '\n' {
            line += 1;
        }
        byte += ch.len_utf8();
    }
    // Now advance `character` UTF-16 code units within the line.
    let mut col = 0u32;
    for ch in src[byte..].chars() {
        if col >= pos.character || ch == '\n' {
            break;
        }
        col += ch.len_utf16() as u32;
        byte += ch.len_utf8();
    }
    byte.min(src.len())
}

/// Encode semantic token spans as LSP delta-encoded `SemanticToken` entries.
#[expect(
    clippy::cast_possible_truncation,
    reason = "LSP positions are u32; documents in practice never approach 4 GiB"
)]
fn encode_tokens(
    spans: &[(std::ops::Range<usize>, u32)],
    rope: &Rope,
    src: &str,
) -> Vec<SemanticToken> {
    let mut out = Vec::with_capacity(spans.len());
    let mut prev_line = 0u32;
    let mut prev_start = 0u32;

    for (range, token_type) in spans {
        if range.is_empty() {
            continue;
        }
        let start_pos = byte_to_position(range.start, rope, src);
        let end_offset = range.end.min(src.len());
        let length = src[range.start..end_offset].encode_utf16().count() as u32;

        let delta_line = start_pos.line - prev_line;
        let delta_start = if delta_line == 0 {
            start_pos.character - prev_start
        } else {
            start_pos.character
        };

        out.push(SemanticToken {
            delta_line,
            delta_start,
            length,
            token_type: *token_type,
            token_modifiers_bitset: 0,
        });

        prev_line = start_pos.line;
        prev_start = start_pos.character;
    }

    out
}
