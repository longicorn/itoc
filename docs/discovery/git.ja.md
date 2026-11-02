# /discovery/git

## 目的

`/discovery/git` はrepositoryのGit履歴を使って `PathLister` を実装する。現在、本番で探索証拠を取得する唯一のsourceである。

## 現在の動作

`GitPathLister` はrepository rootを特定し、入力をtrackedなrepository相対pathに解決する（削除済みtracked pathは履歴によるfallbackを使う）。対象pathに関係するcommitを `git log --follow` で取得し、各commitのsubject/bodyと変更pathを集める。Git command実行には `GitRunner` を注入でき、既定実装は `git` executableを呼ぶ。

`collectCochangeEvidence` は対象commitで同時に変更された対応済みcode/document pathを集約する。commit内の重複pathと対象自身、未対応file typeを除外し、commit集合を記録する。commit message由来のnoise termも蓄積する。score付けと分類は後段の [`/analysis`](../analysis.ja.md) が行う。

現在利用するのはcommitの共変更証拠である。[PRD](../prd.ja.md)に可能性として記載されたbranch、pull request、merge、diff内容、author、時系列、work itemの証拠はまだ使わない。

## 関連path

- [`/discovery`](../discovery.ja.md) はadapter境界を定める。
- [`/analysis`](../analysis.ja.md) は集約済み証拠を解釈する。
