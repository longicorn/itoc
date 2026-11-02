# /discovery

## 目的

`/discovery` は、証拠の取得元から独立して、関連pathの証拠を取得し候補に変換する共通境界を定める。

## 現在の動作

CLIは方式名のカンマ区切り、または `all` を受け取る。利用可能な実装の唯一の登録元を方式registryとし、`all` は登録済み方式をすべて選ぶ。指定方式を並行実行し、repository-relative pathで結果を統合して各方式の証拠を保持する。同じ候補を見つけた追加方式ごとに暫定+1一致bonusを加える。確率として較正した値ではなく、今後評価する初期heuristicである。エラーには失敗した方式名を含める。

`DocumentPathLister` は文書拡張子に応じて形式別extractorを選ぶ。Markdown/MDXとreStructuredTextに対応し、リンク、インラインコード/literal、強調/roleを抽出する。ローカルパスは直接解決し、強調またはコード表記の語句はコードファイル名とソース内容で照合する。意味上の言い換えは推論しない。入力拡張子が未対応の場合、複数方式の実行ではdocument方式だけを対象外としてスキップする。document単独実行では未対応形式をエラーで示す。

文書探索ではリポジトリを一度だけ走査してsource indexを作る。source fileは上限付きworker poolで読み込み、抽出語句はメモリ上のindexに照合する。語句ごとにリポジトリ全体を再走査・再読込することを避ける。

## 関連path

- [`/discovery/git`](discovery/git.ja.md) は現在のGit adapterを実装する。
- [`/analysis`](analysis.ja.md) は証拠をscore付き候補に変換する。
