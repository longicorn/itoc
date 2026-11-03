# /output

## 目的

`/output` は準備済み候補を、人または後続toolが利用できる形式に変換する。表示処理では候補の探索、score付け、filterを行わない。

## 現在の動作

`renderCandidates` はtextまたはJSON形式を選ぶ。text出力はimplementation、test、documentの見出しで候補を分ける。共変更commit IDや文書参照などの詳細証拠は既定で隠し、`--verbose` を指定すると表示する。候補が空の場合は該当するmessageを表示する。

JSON出力はcandidate recordの配列で、path、score、共変更commit数、filename bonus、kind、noise score、任意のnoise termを含む。testらしい候補のkindは `test` となる。CLIはrender済みbyte列をstdoutへ書く。output fileへの保存は未実装である。

## 関連path

- [`/cli`](cli/prd.ja.md) はformatを選び、結果を書き出す。
- [`/filter`](filter.ja.md) はrender前に候補集合を整える。
- [`/analysis`](analysis/prd.ja.md) はcandidate fieldと証拠を定める。
