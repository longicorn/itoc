# /cli

## 目的

`/cli` は探索、候補の準備、表示をひとつの処理として実行するcommand-line entry pointである。

## 現在の動作

実行ファイルはrepository内のfile pathをちょうど1つ受け取る。`parseCLI` はoptionを解析し、出力形式、0以上のscore threshold、0以上のgroupごとのlimitを検証する。繰り返し指定できるinclude/exclude flagは `cliConfig` に集約する。

`run` は引数解析、登録済み方式による候補探索、候補準備、textまたはJSON出力を順に実行する。`--discovery` は方式名ひとつ、カンマ区切りの複数方式、または `all` を受け取る（既定は `git`）。選択方式を並行実行し、追加方式でも見つかった候補ごとに暫定で1点加算する。document modeはMarkdown/MDXとreStructuredTextを別々のextractorへ振り分ける。既定noise thresholdを超える候補は隠し、`--show-noisy` または `--max-noise-score` で調整する。詳細証拠は `--verbose` で表示する。errorは `main` に返し、stderrに表示して失敗終了する。

対応optionや例は[README](../README.ja.md)を参照。このページでは一覧を重複させず、責務の境界を説明する。

## 関連path

- [`/discovery`](discovery.ja.md) は候補を取得・分析する。
- [`/filter`](filter.ja.md) は候補集合を整える。
- [`/output`](output.ja.md) は結果を表示形式にする。
