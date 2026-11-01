# itoc

`itoc` は、意図とコードの関係を見つけるためのツールです。名前には **ito**（意図。日本語の「糸」も想起させます）と **code** を組み合わせ、「ito to code」「ito + code」の意味を込めています。

主な目的は [Nagi](https://github.com/longicorn/nagi) で使うことです。特に、semantic pathからコードを探す処理を、より再現可能で説明しやすくすることを目指します。一方、このトレーサビリティの課題はソフトウェア開発全般にもあるため、`itoc` 単独でも使えるツールを目指します。

初期段階ではForward開発を対象に、Git履歴などの証拠から、commitされた文書に対応する実装コードを探します。決定論的な証拠を中心にし、AIは解釈や曖昧なケースの補助に使います。結果には根拠と不確実性を示します。

現在の方針は[製品要求仕様書](docs/prd.md)を参照してください。

## 使い方

GoでCLIをビルドします。

```sh
go build -o itoc .
```

commit済みのリポジトリ内ファイルを指定すると、Git履歴上で同時に変更された候補を探します。

```sh
./itoc docs/cli/adr.md
```

結果は実装・テスト・文書の候補に分けて表示します。既定では各グループ上位10件です。`--limit 5`で件数を減らし、`--limit 0`で全件を表示できます。JSON形式は`--format json`で指定します。
`--min-score N`で最低scoreを指定できます。`--limit`と併用すると、scoreで絞り込んだ後に各グループ上位N件を表示します。

filterは複数回指定できます。

```sh
./itoc --include-ext .go --include-path internal docs/cli/adr.md
./itoc --exclude-path vendor --exclude-ext .md docs/cli/adr.md
./itoc --exclude-candidate internal/router_test.go docs/cli/adr.md
```

glob記号を含まない`--include-path`と`--exclude-path`は、リポジトリ相対pathの部分一致です。globを含む場合はGoの`filepath.Match`形式でpath全体に照合します（`*`は`/`をまたぎません）。`--exclude-candidate`は`implementation`・`document`・`test`の候補group全体、またはrepository-relative pathが完全一致する個別候補を非表示にします。すべてのoptionは`./itoc --help`で確認できます。
