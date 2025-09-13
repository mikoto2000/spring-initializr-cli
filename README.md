Spring Initializr CLI (Go)

Spring Initializr(https://start.spring.io/) をコマンドラインから使いやすくするための軽量 CLI ツールです。指定したオプションから URL を組み立ててプロジェクトをダウンロードし、任意で展開します。外部依存はなく、Go 標準ライブラリのみで動作します。

使い方
- ビルド: `go build -o spring-initializr-cli`（リポジトリ直下）
- ヘルプ: `./spring-initializr-cli -h`
- 対話（TUI）モード: `./spring-initializr-cli -i`

例
- ZIP をダウンロードのみ:
  `./spring-initializr-cli --type maven-project --language java --group-id com.example --artifact-id demo --dependencies web,data-jpa --output demo.zip`

- ダウンロードして展開（`--base-dir` 未指定なら `artifact-id` が展開先になる）:
  `./spring-initializr-cli --dependencies web,security --extract`

- URL のみ確認（ネットワーク不要）:
  `./spring-initializr-cli --dependencies web,data-jpa --dry-run`

TUI の操作（tview ベース）
- 画面上のフォームで各項目を編集（Tab/Shift+Tab で移動）。
- 「Select Dependencies」で依存関係の一覧を表示し、Enter で選択/解除、`d` で完了。
- 「Show URL」で生成 URL を表示。「Download」「Download+Extract」で実行。
- マウス操作にも対応しています。

依存関係の取得
- TUI は起動時に Spring Initializr のメタデータ（`/metadata/client` または `/dependencies`）を取得して一覧に表示します。
- ネットワークに接続できない場合は依存一覧の取得に失敗します。その際はコマンドラインの `--dependencies` 指定をご利用ください。

主なオプション
- `--type` : `maven-project` / `gradle-project` / `gradle-build`（デフォルト: `maven-project`）
- `--language` : `java` / `kotlin` / `groovy`（デフォルト: `java`）
- `--boot-version` : Spring Boot のバージョン（未指定なら Initializr のデフォルト）
- `--group-id`, `--artifact-id`, `--name`, `--description`, `--package-name`, `--packaging`(jar/war), `--java-version`
- `--dependencies` : 依存 ID のカンマ区切り（例: `web,data-jpa,security`）
- `--base-dir` : 展開時のプロジェクトルート名（未指定は `artifact-id`）
- `--output` : ZIP の保存先ファイル名（デフォルト: `<artifact-id>.zip`）
- `--extract` : ZIP をダウンロード後に展開
- `--dry-run` : 作成される URL を表示して終了（ダウンロードはしない）
- `--base-url` : Spring Initializr のベース URL（デフォルト: `https://start.spring.io`）
- `-v` : 冗長ログ

注意
- `--dry-run` はネットワーク不要です。`--extract` やダウンロードはネットワーク接続が必要です。
- `--dependencies` に指定する ID は Spring Initializr の依存 ID を用います（例: `web`, `data-jpa`, `security`, `postgresql` など）。

ライセンス
- 本ソフトウェアは MIT ライセンスです。詳細は `LICENSE` を参照してください。
