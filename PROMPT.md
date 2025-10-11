# Prompt

1. Spring Initializr(https://start.spring.io/)の CLI 版を作成したいです。
2. MITライセンスにしてください。
3. 現在オプションで渡している各種情報を、 TUI で入力できるようにしたい。
4. TUI ライブラリを使い、 Dependency をリストから選択できるようにしたい
5. (tui.go の中身を削除し、)tviewを使って、各種情報を TUIで選択・入力できるようにしたいです。
6. 次に進めてください
7. フィルタの入力にキーが奪われて依存の選択ができません。タブキーでフィルタ入力と選択セクションを移動できるようにしてください。
8. Select Dependenciesで `/` キー押下でフィルター入力をフォーカスしてほしいです
9. Select Dependenciesでチェックを入れたら、フィルタを空にしてほしいです
10. Select Dependenciesでチェックを入て、フィルタを空にしたとき、フィルタ入力欄にフォーカスしてほしいです
11. 追加した依存の一覧を表示したいです。
12. 選択した依存一覧で、名前＋ID 形式で表示するように拡張してください。
13. Boot Versionも、 Project Type と同じようにセレクトできるようにしたいです
14. Java Version も、 Project Type と同じようにセレクトできるようにしたいです
15. GitHub Actions 用のリリースワークフローを作成してください。
16. version オプションでアプリケーションのバージョンを表示するようにしたいです。
17. license オプションでアプリケーションのライセンスと、NOTICE(依存ライブラリ(tview)のライセンス)を表示するようにしたいです。
18. 現状に合わせて READMEを更新してください
19. .gitignoreを作成してください。
20. Boot Version は、 `id` とも `name` とも違う値を入れなければならないようです、何を入れればいいかわかりますか？
21. `/metadata/client の bootVersion.values[].id` には　`3.5.5.RELEASE` の `.RELEASE`や、 `3.4.10.BUILD-SNAPSHOT` の `.BUILD` ように、不要な文字列が入っています。
22. `3.5.5.RELEASE` とすると、以下のように Not Found と怒られます
23. 引数無しで実行した際には、 interactive モードで起動するようにしたい
24. TUI モードで、`Group` と `Artifact ID` を入力された時点で、 `Package Name` を `Group` と `Artifact ID` を `.` で　JOIN した文字列にしたい
25. Project Type のデフォルトを maven-projectに変更したい
25. Java Version のデフォルトを 21 にしたい
26. Base Dir を削除し、 Artifact ID と同じ値を使う用に修正したい
