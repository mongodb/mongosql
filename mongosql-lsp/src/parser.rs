//! Recursive-descent parser for Rust `{:#?}` pretty-debug output.
//!
//! Produces a concrete syntax tree (CST) with byte-offset ranges so that the
//! LSP server can map cursor positions back to node names for hover and can
//! emit semantic token spans for syntax highlighting.

use std::ops::Range;

// ── CST node types ────────────────────────────────────────────────────────────

/// A node in the `{:#?}` concrete syntax tree.
#[derive(Debug)]
pub enum Node {
    /// `Name(payload)` — enum variant with a tuple payload.
    EnumVariant {
        /// Byte range of the variant name.
        name: Range<usize>,
        payload: Option<Box<Node>>,
    },
    /// `Name { field: value, … }` — struct with named fields.
    Struct {
        /// `Some` when the struct is prefixed by an identifier (e.g. `Filter {`).
        name: Option<Range<usize>>,
        fields: Vec<Field>,
    },
    /// `{ Key { … }: value, … }` — Rust `HashMap`/`BTreeMap` debug output.
    ///
    /// Each entry is `(key_node, value_node)` where the key is itself a struct.
    Map { entries: Vec<(Node, Node)> },
    /// `[ item, … ]` — sequence / array.
    Sequence { items: Vec<Node> },
    /// Any atomic leaf: string literal, number, keyword (`true`, `false`, `None`, `Some(…)`).
    Leaf { range: Range<usize> },
}

/// A `key: value` pair inside a struct.
#[derive(Debug)]
pub struct Field {
    /// Byte range of the field name (before the `:`).
    pub key: Range<usize>,
    pub value: Node,
}

/// The fully parsed document, keeping the original source text for range → `&str` look-ups.
#[derive(Debug)]
pub struct ParsedDoc {
    pub root: Node,
    /// The original source text (kept so ranges can be resolved later).
    pub text: String,
}

impl ParsedDoc {
    /// Returns the name of the innermost named node that contains `offset`, if any.
    pub fn node_name_at(&self, offset: usize) -> Option<&str> {
        find_name_at(&self.root, offset, &self.text)
    }
}

// ── Entry point ───────────────────────────────────────────────────────────────

/// Parse a full `{:#?}` document.  Never fails — unknown syntax is stored as `Leaf`.
pub fn parse(text: &str) -> ParsedDoc {
    let mut p = Parser { src: text, pos: 0 };
    let root = p.parse_value();
    ParsedDoc {
        root,
        text: text.to_owned(),
    }
}

// ── Parser internals ──────────────────────────────────────────────────────────

struct Parser<'a> {
    src: &'a str,
    pos: usize,
}

impl Parser<'_> {
    // ── helpers ──────────────────────────────────────────────────────────────

    fn remaining(&self) -> &str {
        &self.src[self.pos..]
    }

    fn peek(&self) -> Option<char> {
        self.remaining().chars().next()
    }

    fn advance(&mut self, bytes: usize) {
        self.pos += bytes;
    }

    fn skip_whitespace(&mut self) {
        while let Some(c) = self.peek() {
            if c.is_whitespace() {
                self.advance(c.len_utf8());
            } else {
                break;
            }
        }
    }

    fn eat(&mut self, ch: char) -> bool {
        if self.remaining().starts_with(ch) {
            self.advance(ch.len_utf8());
            true
        } else {
            false
        }
    }

    // ── identifier (ASCII word chars) ─────────────────────────────────────────

    fn parse_ident(&mut self) -> Option<Range<usize>> {
        let start = self.pos;
        while let Some(c) = self.peek() {
            if c.is_alphanumeric() || c == '_' {
                self.advance(c.len_utf8());
            } else {
                break;
            }
        }
        if self.pos > start {
            Some(start..self.pos)
        } else {
            None
        }
    }

    // ── string literal `"…"` ─────────────────────────────────────────────────

    fn parse_string(&mut self) -> Node {
        let start = self.pos;
        // consume opening `"`
        self.advance(1);
        let mut escaped = false;
        while let Some(c) = self.peek() {
            self.advance(c.len_utf8());
            if escaped {
                escaped = false;
            } else if c == '\\' {
                escaped = true;
            } else if c == '"' {
                break;
            }
        }
        Node::Leaf {
            range: start..self.pos,
        }
    }

    // ── number literal ────────────────────────────────────────────────────────

    fn parse_number(&mut self) -> Node {
        let start = self.pos;
        // optional leading `-`
        if self.remaining().starts_with('-') {
            self.advance(1);
        }
        while let Some(c) = self.peek() {
            if c.is_ascii_digit()
                || c == '.'
                || c == '_'
                || c == 'e'
                || c == 'E'
                || c == '+'
                || c == '-'
            {
                self.advance(c.len_utf8());
            } else {
                break;
            }
        }
        Node::Leaf {
            range: start..self.pos,
        }
    }

    // ── sequence `[ … ]` ─────────────────────────────────────────────────────

    fn parse_sequence(&mut self) -> Node {
        // consume `[`
        self.advance(1);
        let mut items = Vec::new();
        loop {
            self.skip_whitespace();
            if self.remaining().starts_with(']') {
                self.advance(1);
                break;
            }
            if self.remaining().is_empty() {
                break;
            }
            let item = self.parse_value();
            items.push(item);
            self.skip_whitespace();
            if self.remaining().starts_with(',') {
                self.advance(1);
            }
        }
        Node::Sequence { items }
    }

    // ── struct body `{ field: value, … }` or map `{ Key {…}: val, … }` ─────────

    /// Parse the body of a `{…}` block, returning either a `Node::Struct` (for
    /// normal named-field structs) or a `Node::Map` (for Rust `HashMap`/`BTreeMap`
    /// debug output where each key is itself a struct expression).
    ///
    /// The opening `{` must not yet have been consumed.
    fn parse_struct_body(&mut self) -> Node {
        self.advance(1); // consume `{`
        let mut fields: Vec<Field> = Vec::new();
        let mut entries: Vec<(Node, Node)> = Vec::new();

        loop {
            self.skip_whitespace();
            if self.remaining().starts_with('}') {
                self.advance(1);
                break;
            }
            if self.remaining().is_empty() {
                break;
            }

            if let Some(key_range) = self.parse_ident() {
                self.skip_whitespace();
                if self.eat(':') {
                    // Normal struct field: `field_name: value`
                    self.skip_whitespace();
                    let value = self.parse_value();
                    fields.push(Field {
                        key: key_range,
                        value,
                    });
                } else if self.remaining().starts_with('{') {
                    // Map-key pattern: `StructName { inner_fields }: map_value`.
                    // The `ident { … }` is the HashMap key; what follows `:` is the value.
                    let inner = self.parse_struct_body(); // consumes inner `{…}`
                    let key_node = match inner {
                        Node::Struct {
                            fields: inner_fields,
                            ..
                        } => Node::Struct {
                            name: Some(key_range),
                            fields: inner_fields,
                        },
                        other => other,
                    };
                    self.skip_whitespace();
                    self.eat(':'); // consume the map-entry separator
                    self.skip_whitespace();
                    let map_value = self.parse_value();
                    entries.push((key_node, map_value));
                }
                // else: bare ident not followed by `:` or `{` — skip silently
            } else if self.remaining().starts_with('<') {
                // Skip opaque tokens such as `<SchemaCache…>`.
                while let Some(c) = self.peek() {
                    self.advance(c.len_utf8());
                    if c == '>' {
                        break;
                    }
                }
            } else if let Some(c) = self.peek() {
                self.advance(c.len_utf8());
            }

            self.skip_whitespace();
            if self.remaining().starts_with(',') {
                self.advance(1);
            }
        }

        if entries.is_empty() {
            Node::Struct { name: None, fields }
        } else {
            Node::Map { entries }
        }
    }

    // ── tuple payload `( … )` ─────────────────────────────────────────────────

    fn parse_paren_payload(&mut self) -> Node {
        // consume `(`
        self.advance(1);
        self.skip_whitespace();
        if self.remaining().starts_with(')') {
            self.advance(1);
            return Node::Sequence { items: vec![] };
        }
        let inner = self.parse_value();
        self.skip_whitespace();
        if self.remaining().starts_with(',') {
            self.advance(1);
        }
        self.skip_whitespace();
        if self.remaining().starts_with(')') {
            self.advance(1);
        }
        inner
    }

    // ── top-level value dispatcher ────────────────────────────────────────────

    fn parse_value(&mut self) -> Node {
        self.skip_whitespace();
        match self.peek() {
            Some('"') => self.parse_string(),
            Some('[') => self.parse_sequence(),
            Some('{') => {
                // Anonymous `{…}` — may be a struct body or a map body.
                self.parse_struct_body()
            }
            Some('(') => self.parse_paren_payload(),
            Some('<') => {
                // Opaque token such as `<SchemaCache…>` — consume through `>`.
                let start = self.pos;
                while let Some(c) = self.peek() {
                    self.advance(c.len_utf8());
                    if c == '>' {
                        break;
                    }
                }
                Node::Leaf {
                    range: start..self.pos,
                }
            }
            Some(c) if c == '-' || c.is_ascii_digit() => self.parse_number(),
            Some(_) => {
                // identifier: could be keyword, enum variant, or struct name
                if let Some(ident_range) = self.parse_ident() {
                    self.skip_whitespace();
                    match self.peek() {
                        Some('(') => {
                            // enum variant with tuple payload — or `Some(…)` / `None`
                            let payload = self.parse_paren_payload();
                            Node::EnumVariant {
                                name: ident_range,
                                payload: Some(Box::new(payload)),
                            }
                        }
                        Some('{') => {
                            // Named struct: `StructName { field: value, … }`.
                            // `parse_struct_body` returns `Node::Struct { name: None, … }`;
                            // we attach the name here.
                            let body = self.parse_struct_body();
                            match body {
                                Node::Struct { fields, .. } => Node::Struct {
                                    name: Some(ident_range),
                                    fields,
                                },
                                other => other,
                            }
                        }
                        _ => {
                            // bare keyword: `None`, `true`, `false`, etc.
                            Node::Leaf { range: ident_range }
                        }
                    }
                } else {
                    // skip one unrecognised char and return an empty leaf
                    let start = self.pos;
                    if let Some(c) = self.peek() {
                        self.advance(c.len_utf8());
                    }
                    Node::Leaf {
                        range: start..self.pos,
                    }
                }
            }
            None => Node::Leaf {
                range: self.pos..self.pos,
            },
        }
    }
}

// ── Visitor helpers ───────────────────────────────────────────────────────────

/// Walk the CST returning all semantic token spans as `(range, token_type_index)` pairs.
///
/// Token type indices match the legend declared in `server::TOKEN_TYPES`:
/// - 0 `type`     — variant / struct names
/// - 1 `property` — struct field keys
/// - 2 `string`   — string literals
/// - 3 `number`   — numeric literals
/// - 4 `keyword`  — `true`, `false`, `None`, `Some`
pub fn collect_tokens(node: &Node, src: &str) -> Vec<(Range<usize>, u32)> {
    let mut out = Vec::new();
    collect_tokens_inner(node, src, &mut out);
    out
}

fn collect_tokens_inner(node: &Node, src: &str, out: &mut Vec<(Range<usize>, u32)>) {
    match node {
        Node::EnumVariant { name, payload } => {
            out.push((name.clone(), 0)); // `type`
            if let Some(p) = payload {
                collect_tokens_inner(p, src, out);
            }
        }
        Node::Struct { name, fields } => {
            if let Some(n) = name {
                out.push((n.clone(), 0)); // `type`
            }
            for f in fields {
                out.push((f.key.clone(), 1)); // `property`
                collect_tokens_inner(&f.value, src, out);
            }
        }
        Node::Map { entries } => {
            for (key, value) in entries {
                // The key is typically a named struct; walk it to emit its name
                // as a `type` token and its inner fields as `property` tokens.
                collect_tokens_inner(key, src, out);
                collect_tokens_inner(value, src, out);
            }
        }
        Node::Sequence { items } => {
            for item in items {
                collect_tokens_inner(item, src, out);
            }
        }
        Node::Leaf { range } => {
            if range.is_empty() {
                return;
            }
            let text = &src[range.clone()];
            let ty = match text {
                "true" | "false" | "None" | "Some" => 4, // `keyword`
                s if s.starts_with('"') => 2,            // `string`
                s if s.starts_with(|c: char| c.is_ascii_digit() || c == '-') => 3, // `number`
                _ => 4,                                  // treat unknown bare idents as keywords
            };
            out.push((range.clone(), ty));
        }
    }
}

/// Walk the CST and collect byte-offset ranges of every `{`, `(`, or `[` … matching closer.
/// Used to produce folding ranges.
pub fn collect_fold_ranges(node: &Node, src: &str) -> Vec<(usize, usize)> {
    let mut out = Vec::new();
    collect_fold_inner(node, src, &mut out);
    out
}

#[expect(
    clippy::only_used_in_recursion,
    reason = "src is forwarded to recursive calls for API symmetry with collect_tokens_inner"
)]
fn collect_fold_inner(node: &Node, src: &str, out: &mut Vec<(usize, usize)>) {
    // We don't track brace positions directly in the CST; instead we record a range
    // that spans a multi-token node by inspecting the text around its children.
    // For simplicity we emit one fold per struct/sequence/variant-with-payload that
    // contains at least one child.
    match node {
        Node::EnumVariant {
            payload: Some(p), ..
        } => {
            collect_fold_inner(p, src, out);
        }
        Node::Struct { fields, .. } if !fields.is_empty() => {
            // Estimate the byte range of this struct by first/last field
            let first = fields.first().map(|f| f.key.start);
            let last = fields.last().map(|f| node_end(&f.value));
            if let (Some(start), Some(end)) = (first, last) {
                out.push((start, end));
            }
            for f in fields {
                collect_fold_inner(&f.value, src, out);
            }
        }
        Node::Sequence { items } if items.len() > 1 => {
            let first = items.first().map(node_start);
            let last = items.last().map(node_end);
            if let (Some(start), Some(end)) = (first, last) {
                out.push((start, end));
            }
            for item in items {
                collect_fold_inner(item, src, out);
            }
        }
        Node::Map { entries } if !entries.is_empty() => {
            let first = entries.first().map(|(k, _)| node_start(k));
            let last = entries.last().map(|(_, v)| node_end(v));
            if let (Some(start), Some(end)) = (first, last) {
                out.push((start, end));
            }
            for (key, value) in entries {
                collect_fold_inner(key, src, out);
                collect_fold_inner(value, src, out);
            }
        }
        _ => {}
    }
}

fn node_start(n: &Node) -> usize {
    match n {
        Node::EnumVariant { name, .. } => name.start,
        Node::Struct { name: Some(r), .. } => r.start,
        Node::Struct { fields, .. } => fields.first().map_or(0, |f| f.key.start),
        Node::Map { entries } => entries.first().map_or(0, |(k, _)| node_start(k)),
        Node::Sequence { items } => items.first().map_or(0, node_start),
        Node::Leaf { range } => range.start,
    }
}

fn node_end(n: &Node) -> usize {
    match n {
        Node::EnumVariant {
            payload: Some(p), ..
        } => node_end(p),
        Node::EnumVariant { name, .. } => name.end,
        Node::Struct { fields, .. } => fields.last().map_or(0, |f| node_end(&f.value)),
        Node::Map { entries } => entries.last().map_or(0, |(_, v)| node_end(v)),
        Node::Sequence { items } => items.last().map_or(0, node_end),
        Node::Leaf { range } => range.end,
    }
}

fn find_name_at<'s>(node: &Node, offset: usize, src: &'s str) -> Option<&'s str> {
    match node {
        Node::EnumVariant { name, payload } => {
            if name.contains(&offset) {
                return Some(&src[name.clone()]);
            }
            if let Some(p) = payload {
                return find_name_at(p, offset, src);
            }
            None
        }
        Node::Struct { name, fields } => {
            if let Some(n) = name {
                if n.contains(&offset) {
                    return Some(&src[n.clone()]);
                }
            }
            for f in fields {
                if let Some(r) = find_name_at(&f.value, offset, src) {
                    return Some(r);
                }
            }
            None
        }
        Node::Map { entries } => entries.iter().find_map(|(k, v)| {
            find_name_at(k, offset, src).or_else(|| find_name_at(v, offset, src))
        }),
        Node::Sequence { items } => items.iter().find_map(|i| find_name_at(i, offset, src)),
        Node::Leaf { .. } => None,
    }
}
