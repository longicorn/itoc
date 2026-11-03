# /filter

## 目的

`/filter` は利用者の選択条件、順位づけ、groupごとの件数制限を適用し、探索候補を表示用に整える。

## 現在の動作

`prepareCandidates` はextension/path filter、候補/group除外、noise threshold、最低score、sort、groupごとの件数制限を適用する。noise scoreが既定値1を超える候補は隠す。`--max-noise-score` でthresholdを変え、`--show-noisy` ですべての候補を含められる。limit 0は無制限。

通常のpath filterはrepository相対pathの部分一致である。glob記号を含むpatternはGoの `filepath.Match` syntaxでpath全体に照合し、`*` は `/` をまたがない。extension filterは小文字化したextensionを比較する。`--exclude-candidate` はgroup名または完全一致するrepository相対candidate pathを受け取る。

## 関連path

- [`/cli`](cli/prd.ja.md) はoptionを定義する。
- [`/analysis`](analysis/prd.ja.md) はscoreと分類を定義する。
- [`/output`](output.ja.md) は準備済み候補を表示する。
