---
status: proposed
date: 2026-02-13
---

# ADR-0007: ログライブラリの選定

## コンテキスト

[TUI Logging設計](../design/tui-logging.md)に基づき、TTY環境ではスタイル付きログ、CI/non-TTY環境ではプレーンテキストログを出力する。
ログAPIは環境に依存せず統一的に使用でき、出力形式のみを環境に応じて切り替えたい。

## 検討した選択肢

### 選択肢1: charmbracelet/log + log/slog

charm logは`slog.Handler`インターフェースを実装しているため、TTY環境ではcharm logのスタイル付き出力、CI/non-TTY環境では`slog.TextHandler`によるプレーンテキスト出力に切り替えられる。
`log/slog`を統一APIとして使用することが可能。

#### 良い点

- slogを統一インターフェースとして使えるため、呼び出し側はHandler実装に依存しない
- ADR-0006で採用するBubble Tea/Lip Glossと同一エコシステムでスタイルが統一される
- TTY/non-TTYの切り替えがHandler差し替えのみで実現できる

#### 悪い点

- charm系への依存がさらに増える

### 選択肢2: log/slog単体

Go標準ライブラリのslogのみを使用する。TextHandlerまたはJSONHandlerで出力する。

#### 良い点

- 外部依存なし
- シンプル

#### 悪い点

- TTY環境での視覚的な装飾（色付け、アイコン等）がない
- Bubble TeaのTUIとログ出力のスタイルに一貫性がなくなる

## 決定

`charmbracelet/log` + `log/slog`を採用する。

slogを統一APIとしHandler差し替えでTTY/non-TTYの出力形式を切り替える。
charm系エコシステムに統一することでTUIとログのスタイルに一貫性を持たせる。

## 結果

### 良い影響

- slogを統一APIとすることで、将来のHandler差し替えが容易
- TUI・ログのスタイルが一貫する

### 悪い影響

- `charmbracelet/log`への依存が増える
