# /filter

## 目的

`/filter` は利用者の選択条件、順位づけ、groupごとの件数制限を適用し、探索候補を表示用に整える。

## 現在の動作

`prepareCandidates` は次の順に処理する。extension/pathのinclude・exclude、候補またはgroupの完全一致除外、最低score、score降順（同点はpath順）のsort、groupごとの件数制限。limit 0は無制限を意味する。groupはimplementation、document、testであり、testらしい候補は他のcode候補から分ける。

通常のpath filterはrepository相対pathの部分一致である。glob記号を含むpatternはGoの `filepath.Match` syntaxでpath全体に照合し、`*` は `/` をまたがない。extension filterは小文字化したextensionを比較する。`--exclude-candidate` はgroup名または完全一致するrepository相対candidate pathを受け取る。

## 関連path

- [`/cli`](cli.ja.md) はoptionを定義する。
- [`/analysis`](analysis.ja.md) はscoreと分類を定義する。
- [`/output`](output.ja.md) は準備済み候補を表示する。
