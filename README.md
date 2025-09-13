Spring Initializr CLI (Go)

<img width="1886" height="691" alt="image" src="https://github.com/user-attachments/assets/67fac3b1-1960-4dea-8a42-baec1f8398f4" />

Spring Initializr(https://start.spring.io/) をコマンドラインから使いやすくするための軽量 CLI ツールです。指定したオプションから URL を組み立ててプロジェクトをダウンロードし、任意で展開します。

使い方
- ビルド: `go build -o spring-initializr-cli`（リポジトリ直下）
- ヘルプ: `./spring-initializr-cli -h`
- バージョン表示: `./spring-initializr-cli --version` または `-V`
- ライセンス表示: `./spring-initializr-cli --license` または `-L`
- 対話（TUI）モード: `./spring-initializr-cli -i`

例
- ZIP をダウンロードのみ:
  `./spring-initializr-cli --type maven-project --language java --group-id com.example --artifact-id demo --dependencies web,data-jpa --output demo.zip`

- ダウンロードして展開（`--base-dir` 未指定なら `artifact-id` が展開先になる）:
  `./spring-initializr-cli --dependencies web,security --extract`

- URL のみ確認（ネットワーク不要）:
  `./spring-initializr-cli --dependencies web,data-jpa --dry-run`

TUI の操作（tview ベース）
- 起動時に Spring Initializr のメタデータ（`/metadata/client`）を取得してから画面を表示します。
  - Project Type / Language / Packaging / Boot Version / Java Version はメタデータの候補とデフォルトが反映されます。
- 画面上のフォームで各項目を編集（Tab/Shift+Tab で移動）。
- 依存選択（Select Dependencies）
  - グループごとに一覧表示され、Enter/Space で選択/解除できます。
  - フィルタ（Filter）で ID/名前/グループを絞り込み。
  - ショートカット: `Tab` で Filter と List を切替、`/` で Filter にフォーカス、`d` で完了、`Esc` で閉じる。
  - チェックを入れた直後はフィルタを空にして、Filter にフォーカスが戻ります。
- 「Show Selected」で現在選択している依存を「Name (ID) [Group]」形式で一覧表示。
- 「Show URL」で生成 URL を表示。「Download」「Download+Extract」で実行。

依存関係の取得
- TUI は起動時に Spring Initializr のメタデータ（まず `/metadata/client`、次にフォールバックで `/dependencies`）を取得します。
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
- `--version` / `-V` : バージョン表示
- `--license` / `-L` : アプリケーションおよび依存ライブラリのライセンス表示

注意
- `--dry-run` はネットワーク不要です。`--extract` やダウンロードはネットワーク接続が必要です。
- `--dependencies` に指定する ID は Spring Initializr の依存 ID を用います（例: `web`, `data-jpa`, `security`, `postgresql` など）。
 - TUI のブート/Java バージョンはメタデータのデフォルトが反映されます（ネットワーク未接続時は指定済み値のみ）。

ライセンス
- 本ソフトウェアは MIT ライセンスです。詳細は `LICENSE` を参照してください。
