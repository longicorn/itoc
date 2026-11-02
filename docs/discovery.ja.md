# /discovery

## 目的

`/discovery` は、証拠の取得元から独立して、関連pathの証拠を取得し候補に変換する共通境界を定める。

## 現在の動作

CLIは `--discovery git`、`--discovery document`、`--discovery both` をサポートする。`both` は独立した両方式を並行実行し、channelで結果を回収してrepository-relative pathで候補を統合し、各方式の証拠を保持する。方式の一致によるscore変更はまだ行わず、比較・評価用とする。エラーには失敗した方式名を含める。

`DocumentPathLister` は文書拡張子に応じて形式別extractorを選ぶ。Markdown/MDXとreStructuredTextに対応し、リンク、インラインコード/literal、強調/roleを抽出する。ローカルパスは直接解決し、強調またはコード表記の語句はコードファイル名とソース内容で照合する。意味上の言い換えは推論しない。

文書探索ではリポジトリを一度だけ走査してsource indexを作る。source fileは上限付きworker poolで読み込み、抽出語句はメモリ上のindexに照合する。語句ごとにリポジトリ全体を再走査・再読込することを避ける。

## 関連path

- [`/discovery/git`](discovery/git.ja.md) は現在のGit adapterを実装する。
- [`/analysis`](analysis.ja.md) は証拠をscore付き候補に変換する。
