---
status: proposed
date: 2026-02-13
---

# ADR-0006: TUIフレームワークの選定

## コンテキスト

[TUI Logging設計](../design/tui-logging.md)に基づき、TTY環境ではspinner付きステップ表示とコマンド出力の折りたたみ表示を提供する。
この実現にはTUIフレームワークが必要である。

## 検討した選択肢

### 選択肢1: Bubble Tea (charmbracelet/bubbletea)

Elm Architectureベースの関数型TUIフレームワーク。
Model-Update-Viewサイクルでメッセージ駆動のUI更新を行う。
Bubbles（spinner等のコンポーネント）、Lip Gloss（スタイリング）とエコシステムを形成する。

#### 良い点

- spinner・進捗表示等のカスタムUIを柔軟に構築できる
- Bubbles/Lip Glossとの統合でスタイリングが容易
- GitLabのCLIがtviewからBubble Teaへ移行するなど大規模プロジェクトでの採用実績がある

#### 悪い点

- Elm Architectureに馴染みがなければ学習コストがある
- シグナルハンドリング（SIGINT等）を自前で処理する必要がある

### 選択肢2: tview (rivo/tview)

tcellベースのウィジェット型TUIフレームワーク。
テーブル、フォーム、リスト等のプリビルトウィジェットを提供する。

#### 良い点

- 伝統的なウィジェット型APIで学習コストが低い
- マルチペインレイアウトが得意

#### 悪い点

- 今回必要なspinner・折りたたみログのようなカスタムUIはウィジェットとして提供されておらず自前実装が必要
- 本プロジェクトではマルチペインレイアウトは不要

## 決定

Bubble Tea (charmbracelet/bubbletea) + Bubbles + Lip Glossを採用する。

今回必要なUIはspinner付きステップ表示とコマンド出力の折りたたみでありBubble Teaのカスタム描画が適している。
tviewのウィジェットベースのアプローチはマルチペインレイアウト向けであり本ユースケースには合わない。

## 結果

### 良い影響

- 柔軟なカスタムUIにより設計書通りの表示を実現できる

### 悪い影響

- charmbracelet系への依存が複数増える（bubbletea, bubbles, lipgloss）