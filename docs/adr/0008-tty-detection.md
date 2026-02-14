---
status: proposed
date: 2026-02-13
---

# ADR-0008: TTY判定ライブラリの選定

## コンテキスト

[TUI Logging設計](../design/tui-logging.md)に基づき、stdoutがTTYかどうかで出力形式（TUI/テキストログ）を自動切り替えする。
この判定を行なうライブラリの選定が必要である。

## 検討した選択肢

### 選択肢1: `mattn/go-isatty`

ファイルディスクリプタがターミナルかどうかを判定する軽量ライブラリ。

#### 良い点

- Cygwin/MSYS2のpty判定にも対応
- 広く使われている（charmbracelet系でも内部依存）
- 軽量で単一責務

#### 悪い点

- 外部依存が増える

### 選択肢2: `golang.org/x/term`

Go準標準ライブラリのターミナル操作パッケージ。
`term.IsTerminal()`を提供する。

#### 良い点

- Go準標準ライブラリであり信頼性が高い

#### 悪い点

- Cygwin/MSYS2のpty判定に非対応
- ターミナルサイズ取得等の不要な機能も含む

## 決定

`mattn/go-isatty`を採用する。

Cygwin/MSYS2対応が有用であり、charmbracelet系の内部でも依存されているため実質的に新規依存とならない。

## 結果

### 良い影響

- Cygwin/MSYS2環境でも正しくTTY判定できる

### 悪い影響

- 外部依存が1つ増える（ただしcharmbracelet系の推移的依存として既に存在する見込み）