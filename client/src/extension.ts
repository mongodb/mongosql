import * as path from "path";
import * as vscode from "vscode";
import {
  LanguageClient,
  LanguageClientOptions,
  ServerOptions,
  TransportKind,
} from "vscode-languageclient/node";

let client: LanguageClient;

export function activate(context: vscode.ExtensionContext): void {
  const command =
    process.env["SERVER_PATH"] ??
    context.asAbsolutePath(path.join("..", "target", "debug", "mongosql-lsp"));

  const run: ServerOptions = {
    command,
    transport: TransportKind.stdio,
    options: { env: { ...process.env, RUST_LOG: "debug" } },
  };

  const clientOptions: LanguageClientOptions = {
    documentSelector: [
      { scheme: "file", language: "mongosql-mir" },
      { scheme: "file", language: "mongosql-air" },
    ],
  };

  client = new LanguageClient(
    "mongosql-lsp",
    "MongoSQL LSP",
    run,
    clientOptions
  );

  context.subscriptions.push(client);
  client.start();
}

export function deactivate(): Thenable<void> | undefined {
  return client?.stop();
}
