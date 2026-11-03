# Project・Milestone 同期を任意化する判断

状態: #167 で方針承認済み。入力・実行コード・Go 回帰テストを実装済み、未公開。
関連: [#136](https://github.com/kudoas/sync-issue-field/issues/136)、[#137](https://github.com/kudoas/sync-issue-field/issues/137)、[#21](https://github.com/kudoas/sync-issue-field/issues/21)。

現行ブランチの入力・既定値・同期範囲・認証・移行手順の正本は [README](../README.md#optional-milestone-and-project-synchronization)。本書は採用理由と検証上の留意点を記録する。新しい動作は本変更を含むバージョンの公開後に利用できる。

## 決定と理由

担当者・ラベルを基本とし、Milestone と Project は独立して既定で無効にする。無効時は取得と更新を両方省き、子の独自値を保持する。Project を有効にしない利用へ追加の Projects 認証を要求しない。

子を対象とする実行時に親の値を反映する。親の変更から子全体への自動伝播は追加しない。Project 有効時の先頭1件への追加という既存制約と、既存の親選択は維持する。全所属やカスタムフィールドの同期は約束しない。

| 案 | 判断 |
|---|---|
| 全項目を常に同期 | 基本機能にも Project 取得と追加認証が結び付くため不採用 |
| Project・Milestone 同期を削除 | チェックリスト互換や明示的な再同期の利用まで失うため不採用 |
| 独立して任意化 | 基本機能の認証を単純にし、必要な項目だけ追加できるため採用 |

## 変更前の根拠と公式仕様

変更前の対象は `b9bbb05ced3fcd69996527e922b3eebcefe56b7a`。

- [ユースケース](https://github.com/kudoas/sync-issue-field/blob/b9bbb05ced3fcd69996527e922b3eebcefe56b7a/usecase/sync_issue_fields.go#L22-L70) は入力の Issue のみを更新し、Project には `ProjectIDs[0]` を使用していた。
- [クライアント](https://github.com/kudoas/sync-issue-field/blob/b9bbb05ced3fcd69996527e922b3eebcefe56b7a/infrastructure/github/client.go#L87-L153) は Milestone・Projects・未使用 FieldValues を無条件で取得していた。今回の実装では任意フィールドを有効時だけ個別に取得し、FieldValues を削除した。基本クエリを単純に保つため、任意項目を有効にすると各項目につき追加の読み取りが1回必要になる。
- [GitHub の公式告知](https://github.blog/changelog/2025-09-11-a-rest-api-for-github-projects-sub-issues-improvements-and-more/) は Sub-issue が親の Project・Milestone を既定で継承すると説明する。作成後の追従や競合時の全条件、チェックリスト互換でも同じ動作になるとはこの告知だけでは判断しない。
- [GitHub の認証仕様](https://docs.github.com/en/issues/planning-and-tracking-with-projects/automating-your-project/automating-projects-using-actions) では `GITHUB_TOKEN` は Projects にアクセスできず、GitHub App または PAT が必要。

## 互換性と復旧

既定動作の変更は 0.x の次の minor で公開する。従来の同期範囲が必要な利用者は両オプションを明示的に有効にし、Projects 対応の認証を設定する。問題発生時は従来の公開バージョンへ戻せる。既存タグ・バイナリは変更しない。

Issue 更新と Project 追加は別 API で、原子性は保証しない。任意フィールドの取得エラーでは子を更新しないが、Project 追加時に失敗すると Issue 更新は済んでいる。この場合は途中更新を明示するエラーを返す。

## 検証と残る確認

Go の設定・launcher・HTTP fixture・配布テストで、4通りの有効／無効、不正値、無効項目の無取得・無更新、有効な Milestone の設定と解除、Project 追加、取得権限エラーと途中更新を確認する。実際の実行結果は実装 PR に記録する。

標準継承との実API比較は未実施。隔離した Issue・Project で Action なしの継承と明示実行を比較し、親変更・子独自値・親の値解除を読み戻して記録する必要がある。GitHub App・PAT の両認証による Projects の検証も別途必要。本実装だけで #136 / #137 の実API確認まで完了したとは扱わない。

比較結果により前提が変われば、本書・README・関連 Issue を更新する。標準機能が完全に代替するとの仮定に依存して任意化したわけではない。
