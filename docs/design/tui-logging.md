# TUI Logging

## Summary

- TTYの有無、CI環境かどうかに応じて出力形式を自動で変更
- 全ての出力はstdoutに行う。ハッシュのみ取得したい場合は `--quiet` フラグを使用する
- ログメッセージは逐次追加方式で表示する（未来のステップは表示しない）

## 出力形式

stdoutがTTYの場合はTUI出力、TTYでない場合はテキストログを出力する。
CI環境によっては[独自の折りたたみ構文](#ci環境別の折りたたみ構文)を使用する。
CI環境の判定に関しては[CI環境判定](#ci環境判定)に記載する。

### TUI表示

TUI表示には2種類のログ要素がある。

#### ステップ表示

短時間で完了する操作は過去形で即座に `✓` 付きで表示する。
時間のかかる操作（ハッシュ計算、tarball作成等）はspinner付きで進行中表示し、完了時に `✓` に、エラー時は `✗` に変化する。

#### コマンド実行表示

`pnpm install`等の別コマンド呼び出し操作はspinner付きで折りたたまれた末尾数行のコマンド出力を表示する。
実行中はリングバッファで末尾N行を保持し表示する。バッファから溢れた行数を末尾に表示する。
正常終了時は1行に折りたたむ。
異常終了時はリングバッファの全内容を展開して表示する。

#### 表示例

##### 実行中

```text
✓ Loaded pnpm-lock.yaml from ./source/pnpm-lock.yaml
⠹ Executing `pnpm install` in 12s ...
  | Progress: resolved 123, reused 0, downloaded 32, added 32
  | hogehoge
  | fugafuga
  + ---------------------------------------------- (+123 lines) --
```

##### 正常終了

```text
✓ Loaded pnpm-lock.yaml from ./source/pnpm-lock.yaml
 `pnpm install` exit successfully in 12m 34s
✓ Normalized store
⠹ Computing hash...
```

##### 異常終了

```text
✓ Loaded pnpm-lock.yaml from ./source/pnpm-lock.yaml
 `pnpm install` failed with exit code 127 in 12m 34s
  | (リングバッファの全内容を展開)
  | ERROR: error message
  | hogehoge
  | fugafuga
  | ...
```

### テキストログ

TUI表示をしない場合のログ表示。
別コマンド呼び出しはスコープのような形式で区別する。

#### 表示例

```text
2006-01-02T15:04:05+07:00 INFO hogehoge
2006-01-02T15:04:05+07:00 INFO scope="pnpm install" fugafuga
```

## CI環境判定

CI環境の判定は固有の環境変数の有無とその値で行なう。
[CI環境別の折りたたみ構文](#ci環境別の折りたたみ構文)の対応のため、一部のCI環境かどうかの判定を個別に行なう必要がある。

### CI環境別の環境変数

| CI環境          | 環境変数名         | 値      |
| --------------- | ------------------ | ------- |
| 任意のCI環境    | `CI`               | `true`  |
| GitHub Actions  | `GITHUB_ACTIONS`   | `true`  |
| GitLab CI       | `GITLAB_CI`        | `true`  |
| Azure Pipelines | `TF_BUILD`         | `true`  |
| TeamCity        | `TEAMCITY_VERSION` | `!= ""` |
| Buildkite       | `BUILDKITE`        | `true`  |
| Travis CI       | `TRAVIS`           | `true`  |

### CI環境別の折りたたみ構文

| CI環境          | 開始                                              | 終了                                     |
| --------------- | ------------------------------------------------- | ---------------------------------------- |
| GitHub Actions  | `::group::{title}`                                | `::endgroup::`                           |
| GitLab CI       | `\e[0Ksection_start:{unix_ts}:{id}\r\e[0K{title}` | `\e[0Ksection_end:{unix_ts}:{id}\r\e[0K` |
| Azure Pipelines | `##[group]{title}`                                | `##[endgroup]`                           |
| TeamCity        | `##teamcity[blockOpened name='{title}']`          | `##teamcity[blockClosed name='{title}']` |
| Buildkite       | `--- {title}`                                     | (次のセクション開始で自動終了)           |
| Travis CI       | `travis_fold:start:{id}\n{title}`                 | `travis_fold:end:{id}`                   |
