# Project・Milestone 同期を任意化する方針

状態: 採用方針。入力・実行コードへの反映は未実施。
関連: [#136](https://github.com/kudoas/sync-issue-field/issues/136)、[#137](https://github.com/kudoas/sync-issue-field/issues/137)、[#21](https://github.com/kudoas/sync-issue-field/issues/21)。

この文書は、任意化の理由・対象範囲・実装の完了条件を定める。公開済みの Action の使用方法は [README](../README.md) を参照する。未実装の入力を使用例として扱わない。

## 決定と提供する動作

担当者・ラベルを基本機能にする。Project と Milestone は独立して有効化する機能として残す。次の入力名と既定値を実装案とする。

| 入力 | 既定値 | 無効時 | 有効時 |
|---|---|---|---|
| `sync-milestone` | `false` | 親の Milestone を取得せず、子の値を変更しない | 実行時点の親の Milestone を子へ設定する。親が未設定なら子も解除する |
| `sync-projects` | `false` | Projects を一切取得・更新せず、子の所属を変更しない | 親の Project への追加処理を実行する。子独自の Project は削除しない |

担当者・ラベルは親の集合で子の集合を置き換える。親に値がなければ子も空にし、子独自の値を残す追加専用の動作にはしない。取得上限への対応は [#138](https://github.com/kudoas/sync-issue-field/issues/138) で独立して行う。

Project 有効時の初回実装は、現行の「取得した親の Project の先頭1件へ追加」を維持する。現行は親の Project を最大5件取得して先頭1件だけを使用するため、全所属の同期を保証しない。この制約を使用例にも明記し、複数所属の選択・全件対応を任意化と混同しない。カスタムフィールド値は同期対象にしない。

## 起動条件と同期方向

Action の入力は子 Issue を指定し、その子から親を取得する。実行時点の親の値を子へ反映する。親の変更から配下の子全体へ自動伝播する機能は提供しない。親を変更しただけで各子へ再同期されるとは案内しない。

子の独自値は、有効にした同期項目では上記の規則で変更される。無効にした項目では保持する。明示的な Sub-issue の親を優先し、親がなければ一意なチェックリスト追跡先へフォールバックする既存の親選択は維持する。

## 観測した現行実装と公式仕様

調査対象: `b9bbb05ced3fcd69996527e922b3eebcefe56b7a`。

- [ユースケース](https://github.com/kudoas/sync-issue-field/blob/b9bbb05ced3fcd69996527e922b3eebcefe56b7a/usecase/sync_issue_fields.go#L22-L70) は入力の Issue を更新する。子の一覧を取得して更新する処理はない。Project への追加には `ProjectIDs[0]` を使用する。
- [クライアント](https://github.com/kudoas/sync-issue-field/blob/b9bbb05ced3fcd69996527e922b3eebcefe56b7a/infrastructure/github/client.go#L87-L153) は任意化の入力なしに Milestone・Projects・未使用の FieldValues を取得する。無効時は mutation を省略するだけでなく、この取得も省く必要がある。
- [リポジトリのイベント設定](https://github.com/kudoas/sync-issue-field/blob/b9bbb05ced3fcd69996527e922b3eebcefe56b7a/.github/workflows/sync-issue.yml#L1-L30) は `opened` / `edited` / `labeled` / `unlabeled`。これだけで親のフィールド変更が子へ伝播するとは扱わない。
- [GitHub の公式告知](https://github.blog/changelog/2025-09-11-a-rest-api-for-github-projects-sub-issues-improvements-and-more/) は Sub-issue が親の Project・Milestone を既定で継承すると説明する。ただし、その告知だけでは作成後の追従、子の独自値、解除時の全条件を確定できない。チェックリスト互換機能にも同じ継承があるとは判断しない。
- [GitHub の認証仕様](https://docs.github.com/en/issues/planning-and-tracking-with-projects/automating-your-project/automating-projects-using-actions) では `GITHUB_TOKEN` は Projects にアクセスできず、GitHub App または PAT が必要。通常の Issue 同期へこの追加認証を要求しないため、Project の既定値を無効にする。

標準機能について、親変更後の追従や全競合条件の実API比較テストは未実施。この方針は「標準機能が完全に代替する」との仮定には依存しない。

## 選択肢と採用理由

| 案 | 利点 | 制約・採否 |
|---|---|---|
| 常に同期する現行仕様を維持 | 入力の追加が不要 | Issue 同期にも Projects 取得と追加認証が結び付くため不採用 |
| Project・Milestone 同期を削除 | 設定と認証が単純 | チェックリスト互換や明示的な再同期の利用まで失うため不採用 |
| 独立して任意化する | 基本機能は Issue 権限だけで利用でき、必要な項目を追加できる | 入力の検証・無効時の取得除外が必要。採用 |

## 互換性と反映順序

既定で Project・Milestone を同期しなくなることは、現在の利用者にとって互換性変更になる。0.x の次の minor リリースで提供し、既存利用者が従来の範囲を維持するには両入力を明示的に有効にする必要があるとリリースノートへ記載する。公開済みタグ・バイナリは変更しない。

1. 本方針をレビューする。#138 の取得上限修正は独立して反映できる。
2. #21 のうち今回必要な2入力を Action → launcher → 環境設定 → ユースケース／クエリへ通す。真偽値以外は入力エラーにする。
3. #137 で有効時の認証・権限不足の動作を整え、基本の使用例を `GITHUB_TOKEN` のみにする。無効時には Project 関連の要求を送らない。
4. #139 の classic 権限案内、#140 の未使用 FieldValues を整理する。Project ID の取得と FieldValues の取得を区別する。
5. 回帰テスト・任意の実API検証を行い、README の現行仕様と使用例を更新してから公開する。

Issue の更新と Project 追加は別 API 呼び出しであり、全体の原子性は保証できない。権限不足を含む失敗を成功として扱わず、途中更新があり得ることを有効時の仕様・テストに含める。

## 検証と Issue の完了条件

実装時の自動テストは Go で記述する。

- 既定値と4通りの有効／無効の組み合わせ、真偽値以外の入力エラーを検証する。
- HTTP fixture で、無効な項目が選択クエリ・mutation に含まれないこと、子の独自値を保持することを確認する。
- 有効な Milestone の設定・解除、有効な Project の追加、認証エラーと途中更新を確認する。
- launcher と実配布バイナリで入力が伝達されることを確認する。
- 隔離した親子 Issue・Project を使い、Action を実行しない標準継承と、明示実行した場合の結果を比較する。親変更・子の独自値・親の値解除をそれぞれ読み戻して記録する。専用リソースは検証後に片付ける。

本書のマージだけで #136 の実API比較や実装まで完了したとは扱わない。結果により前提が変われば本書と関連 Issue・README を更新する。問題発生時は利用側を従来の公開バージョンへ戻せる。
