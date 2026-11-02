# /analysis

## 目的

`/analysis` は探索で得た証拠を、順位づけ・分類された成果物候補に変換する。証拠の収集と解釈を分離し、探索アダプターがリポジトリ情報の取得に集中できるようにする。

## 現在の動作

現行実装は、リポジトリ相対path、対象と同時に変更されたcommitの集合、commit messageから得たノイズ情報を持つ `cochangeEvidence` を受け取る。`analyzeCandidates` は各証拠から候補を作る。

scoreは共変更commit数とfilenameの類似bonusの合計である。noise scoreは別に管理する。commit messageのsignalは小さく加点し、benchmark/performance系pathなどの内蔵ルールは大きく加点する。既定noise thresholdを超える候補は隠し、`--show-noisy` で含められる。`--max-noise-score` でthresholdを変えられる。noise ruleは利用者のinclude/exclude filterとは独立して追加できる。

これらはGit由来の証拠に対するheuristicであり、scoreはsemanticな関係を確定しない。現在の実装はsourceの意味解析やLLMを利用しない。

## 処理の流れ

`discoverWithLister` が `PathLister` から証拠を取得し、`analyzeCandidates` を呼ぶ。候補の準備と表示は [`/filter`](filter.md) と [`/output`](output.md) を参照。

## 関連する方針

[PRD](prd.ja.md) はより広い証拠と不確実性の扱いを定める。将来分析を拡張する場合も、確認可能な証拠を保ち、候補を確定したlinkとして示さない。
