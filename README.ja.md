# itoc

`itoc` は、意図とコードの関係を見つけるためのツールです。名前には **ito**（意図。日本語の「糸」も想起させます）と **code** を組み合わせ、「ito to code」「ito + code」の意味を込めています。

`itoc` は、意図を記した文書からコードを探す処理を、より再現可能で説明しやすくすることを目指します。このトレーサビリティの課題はソフトウェア開発全般にあるため、単独でも使えるツールとして開発しています。

初期段階ではForward開発を対象に、Git履歴などの証拠から、commitされた文書に対応する実装コードを探します。決定論的な証拠を中心にし、AIは解釈や曖昧なケースの補助に使います。結果には根拠と不確実性を示します。

現在の方針は[製品要求仕様書](docs/prd.ja.md)を参照してください。

## インストール

Go 1.26以降とGitを用意し、リポジトリのルートでCLIをビルドしてインストールします。

```sh
go install .
```

Goのバイナリディレクトリを`PATH`に追加してください。`GOBIN`を設定していない場合は`$(go env GOPATH)/bin`です。カレントディレクトリにビルドする場合は`go build -o itoc .`を実行します。

テスト候補の手掛かりを増やすには、JSON出力に対応した[Universal Ctags](https://ctags.io/)をインストールし、`ctags`を`PATH`に追加してください。Ctagsがなくてもファイル名とGitの同時変更から探索できますが、識別子の照合は省略されます。

## 使い方

Gitリポジトリ内で`itoc`を実行します。リポジトリ内のファイルパスを指定するか、追跡対象ファイルの内容からキーワードで探索の起点を探します。

```sh
itoc <path>
itoc --search <keyword>
```

`--search`で複数のファイルが一致した場合は、番号を選んで探索の起点を指定します。既定のGit探索は、そのファイルの履歴から関連候補を探します。すべてのオプションは`itoc --help`で確認できます。

## その他のオプション

結果は実装・テスト・文書の候補に分けて表示します。既定では各グループ上位10件です。`--limit 5`で件数を減らし、`--limit 0`で全件を表示できます。JSON形式は`--format json`で指定します。

実装候補からファイル名とGitの同時変更を手掛かりにテスト候補も探します。[Ctagsによる識別子の一致](docs/discovery/ctags/prd.ja.md)も手掛かりに加わります。機械的な探索には誤検知や見落としがあり、一致はテストによる検証を保証しません。テスト候補を隠すには`--exclude-candidate test`、実装候補からのテスト探索を止めるには`--discover-tests=false`を指定します。

`--min-score N`で最低scoreを指定できます。`--limit`と併用すると、scoreで絞り込んだ後に各グループ上位N件を表示します。

filterは複数回指定できます。

```sh
itoc --include-ext .go --include-path internal docs/cli/adr.md
itoc --exclude-path vendor --exclude-ext .md docs/cli/adr.md
itoc --exclude-candidate internal/router_test.go docs/cli/adr.md
```

探索方式は `--discovery git`（既定）、`--discovery git,document` のようなカンマ区切り、または `--discovery all` で指定できます。指定した方式を並行実行して候補と根拠を統合し、追加の方式でも見つかった候補には方式ごとに暫定で1点加算します。`all` は登録済みのすべての方式を実行します。文書探索はMarkdown/MDXとreStructuredTextに対応し、ローカルリンク、インラインコード、強調語を文字どおりの手掛かりとして使います。

```sh
itoc --discovery document docs/cli/adr.md
```

文書探索は意味上の言い換えを推論しません。一致したパス、ファイル名、ソース本文と文書行の詳細は `--verbose` で表示します。
内蔵ノイズルールにより、noise scoreが1を超える候補は既定で隠します。`--show-noisy` ですべて含めるか、`--max-noise-score N` でthresholdを変更できます。

glob記号を含まない`--include-path`と`--exclude-path`は、リポジトリ相対pathの部分一致です。globを含む場合はGoの`filepath.Match`形式でpath全体に照合します（`*`は`/`をまたぎません）。`--exclude-candidate`は`implementation`・`document`・`test`の候補group全体、またはrepository-relative pathが完全一致する個別候補を非表示にします。

## ドキュメント

製品要求と各コンポーネントの文書は[docs/](docs/)を参照してください。

Git探索は既定で現在のパスの履歴を使います。`--follow`を指定すると、Git探索とテスト探索で名前変更前の履歴もたどります。存在しない過去のパスに対しては、ファイル名、内容の完全一致、Ctagsの共通定義を手掛かりに、現在の追跡ファイルを移動先候補として示せます。移動を見逃したり、無関係のファイルが候補になったりすることがあります。
