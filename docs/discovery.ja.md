# /discovery

## 目的

`/discovery` は、証拠の取得元から独立して、関連pathの証拠を取得し候補に変換する共通境界を定める。

## 現在の動作

`PathLister` がadapter contractである。`List` methodは入力pathを受け取り、共変更証拠、解決した対象path、errorを返す。`discoverWithLister` はadapterを呼び、証拠と対象pathを `analyzeCandidates` 経由で `/analysis` に渡す。

既定の `discover` は現在 `GitPathLister` を選ぶ。interfaceにより別の証拠取得元を注入できるが、Git以外のlisterは未実装である。また、現在のcontractはGit形式の共変更証拠を返す。他の証拠形式へ広げるにはmodelの明示的な設計が必要になる。

## 関連path

- [`/discovery/git`](discovery/git.ja.md) は現在のGit adapterを実装する。
- [`/analysis`](analysis.ja.md) は証拠をscore付き候補に変換する。
