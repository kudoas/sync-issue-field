# Changelog

## Unreleased

The changes below were previously recorded as 0.2.0, but no 0.2.0 tag or GitHub Release was published.


### Features

* Add check for changes before creating pull request with binary ([#82](https://github.com/kudoas/sync-issue-field/issues/82)) ([7505e7b](https://github.com/kudoas/sync-issue-field/commit/7505e7b75048d9f333263efc982507920b0ade4f))
* Add GitHub Actions workflow to build binary and create PR ([0dd8cae](https://github.com/kudoas/sync-issue-field/commit/0dd8caeb556749470053a5d9febbf9b537abd502))
* Optimize build process and update documentation ([#70](https://github.com/kudoas/sync-issue-field/issues/70)) ([b513e9b](https://github.com/kudoas/sync-issue-field/commit/b513e9be99a22a87c0f1148eb6dd8d411f4c5fa8))


### Bug Fixes

* update release workflow to build and commit binary after release-please step ([09bf0a9](https://github.com/kudoas/sync-issue-field/commit/09bf0a922709806bd4ce49ea634aee35c2100c7d))
* update release workflow to build and commit binary after release… ([#72](https://github.com/kudoas/sync-issue-field/issues/72)) ([09bf0a9](https://github.com/kudoas/sync-issue-field/commit/09bf0a922709806bd4ce49ea634aee35c2100c7d))

## [0.3.0](https://github.com/kudoas/sync-issue-field/compare/v0.2.0...v0.3.0) (2026-11-07)


### ⚠ BREAKING CHANGES

* make Project and Milestone synchronization opt-in
* milestones and Projects now default to disabled. Set sync-milestone and sync-projects to true with a Projects-capable token to retain both optional behaviors.

### Features

* make Project and Milestone synchronization opt-in ([92cce4e](https://github.com/kudoas/sync-issue-field/commit/92cce4ea800aa600a22864ce26f8a02c87701bad))
* Project・Milestone 同期を任意化する ([#169](https://github.com/kudoas/sync-issue-field/issues/169)) ([92cce4e](https://github.com/kudoas/sync-issue-field/commit/92cce4ea800aa600a22864ce26f8a02c87701bad))


### Bug Fixes

* synchronize complete assignee and label sets ([7a78c1b](https://github.com/kudoas/sync-issue-field/commit/7a78c1bcade715a2e32644d6000ce9acda0eb377))
* 親の担当者・ラベルを全件同期する ([#166](https://github.com/kudoas/sync-issue-field/issues/166)) ([7a78c1b](https://github.com/kudoas/sync-issue-field/commit/7a78c1bcade715a2e32644d6000ce9acda0eb377))

## [0.2.0](https://github.com/kudoas/sync-issue-field/compare/v0.2.0-rc.1...v0.2.0) (2026-11-02)


### ⚠ BREAKING CHANGES

* **deps:** Go・OAuth2・Docker の実行環境を更新 ([#163](https://github.com/kudoas/sync-issue-field/issues/163))

### Bug Fixes

* **deps:** update compatible dependencies and enforce three-day release age ([840bb90](https://github.com/kudoas/sync-issue-field/commit/840bb9065990e9faa7b446531aeee973adda98da))
* **deps:** 互換更新と公開後3日間の待機ルールを適用 ([#161](https://github.com/kudoas/sync-issue-field/issues/161)) ([840bb90](https://github.com/kudoas/sync-issue-field/commit/840bb9065990e9faa7b446531aeee973adda98da))
* prefer explicit Sub-issue parents for synchronization ([c3bee8b](https://github.com/kudoas/sync-issue-field/commit/c3bee8b9e04a6a345d357a41575b521b8f009faa))
* Sub-issueの親からIssueフィールドを同期する ([#151](https://github.com/kudoas/sync-issue-field/issues/151)) ([c3bee8b](https://github.com/kudoas/sync-issue-field/commit/c3bee8b9e04a6a345d357a41575b521b8f009faa))


### Build System

* **deps:** Go・OAuth2・Docker の実行環境を更新 ([#163](https://github.com/kudoas/sync-issue-field/issues/163)) ([979958b](https://github.com/kudoas/sync-issue-field/commit/979958b55e0ff26639ebceabe121de12d338ac36))

## [0.2.0-rc.1](https://github.com/kudoas/sync-issue-field/compare/v0.1.3...v0.2.0-rc.1) (2026-11-02)


### Features

* Add check for changes before creating pull request with binary ([#82](https://github.com/kudoas/sync-issue-field/issues/82)) ([7505e7b](https://github.com/kudoas/sync-issue-field/commit/7505e7b75048d9f333263efc982507920b0ade4f))
* Add environment variables for GitHub Actions inputs in action.yml ([597b298](https://github.com/kudoas/sync-issue-field/commit/597b2984f05699e8dcc4ac74db82e2b1cd9c3d54))
* Add GitHub Actions workflow to build binary and create PR ([0dd8cae](https://github.com/kudoas/sync-issue-field/commit/0dd8caeb556749470053a5d9febbf9b537abd502))
* Add GitHub Actions workflow to sync issue fields and update README ([2bc8414](https://github.com/kudoas/sync-issue-field/commit/2bc8414444a6d3cc73fc04273b34d3dcb7fd1e05))
* GitHub Releasesからビルド済みバイナリを配布する ([#142](https://github.com/kudoas/sync-issue-field/issues/142)) ([cfa719d](https://github.com/kudoas/sync-issue-field/commit/cfa719deb0f474e07e4b172e2334b01d1108b750))
* Optimize build process and update documentation ([#70](https://github.com/kudoas/sync-issue-field/issues/70)) ([b513e9b](https://github.com/kudoas/sync-issue-field/commit/b513e9be99a22a87c0f1148eb6dd8d411f4c5fa8))
* Update environment variable names in action.yml and env.go for consistency ([a1df13e](https://github.com/kudoas/sync-issue-field/commit/a1df13e88f47ae3e19e760a0e9c72232070b60ec))
* Upgrade Go version to 1.23 and refactor GitHub client interface ([#90](https://github.com/kudoas/sync-issue-field/issues/90)) ([28dcc29](https://github.com/kudoas/sync-issue-field/commit/28dcc2964dba00788e2ef67a9478c966851f2b74))


### Bug Fixes

* **deps:** update module golang.org/x/oauth2 to v0.29.0 ([#62](https://github.com/kudoas/sync-issue-field/issues/62)) ([cb95141](https://github.com/kudoas/sync-issue-field/commit/cb9514140e7c1753658ddfb138654ab61178077e))
* **deps:** update module golang.org/x/oauth2 to v0.30.0 ([#111](https://github.com/kudoas/sync-issue-field/issues/111)) ([52889fd](https://github.com/kudoas/sync-issue-field/commit/52889fdccc2d99fb9d8653e73e3c969f8ee913ca))
* update release workflow to build and commit binary after release-please step ([09bf0a9](https://github.com/kudoas/sync-issue-field/commit/09bf0a922709806bd4ce49ea634aee35c2100c7d))
* update release workflow to build and commit binary after release… ([#72](https://github.com/kudoas/sync-issue-field/issues/72)) ([09bf0a9](https://github.com/kudoas/sync-issue-field/commit/09bf0a922709806bd4ce49ea634aee35c2100c7d))

## [0.1.3](https://github.com/kudoas/sync-issue-field/compare/v0.1.2...v0.1.3) (2025-01-07)


### Bug Fixes

* **deps:** update github.com/shurcool/githubv4 digest to 4829585 ([#56](https://github.com/kudoas/sync-issue-field/issues/56)) ([0711980](https://github.com/kudoas/sync-issue-field/commit/0711980c877ae84fe9511ab6fcadc99ef97c121a))
* **deps:** update module golang.org/x/oauth2 to v0.22.0 ([#57](https://github.com/kudoas/sync-issue-field/issues/57)) ([346ae92](https://github.com/kudoas/sync-issue-field/commit/346ae92c2d012f11aca947543a322bacd5dfe7b5))
* set branding ([#69](https://github.com/kudoas/sync-issue-field/issues/69)) ([8ffc125](https://github.com/kudoas/sync-issue-field/commit/8ffc1253a69600718ee556e13e41d9eb138ebf0a))
* unwrap error  ([#61](https://github.com/kudoas/sync-issue-field/issues/61)) ([887f041](https://github.com/kudoas/sync-issue-field/commit/887f041db018fa2259494c1240d6fc795560bcbe))

## [0.1.2](https://github.com/kudoas/sync-issue-field/compare/v0.1.1...v0.1.2) (2024-06-05)


### Bug Fixes

* **deps:** update module golang.org/x/oauth2 to v0.21.0 ([#47](https://github.com/kudoas/sync-issue-field/issues/47)) ([3316b1f](https://github.com/kudoas/sync-issue-field/commit/3316b1f57cf2b53145c54079728502a61463d7e9))

## [0.1.1](https://github.com/kudoas/sync-issue-field/compare/v0.1.0...v0.1.1) (2024-05-04)


### Bug Fixes

* **deps:** update module golang.org/x/oauth2 to v0.20.0 ([#35](https://github.com/kudoas/sync-issue-field/issues/35)) ([d48d03e](https://github.com/kudoas/sync-issue-field/commit/d48d03e96bb046accd37e680ea93ed2620fce3e4))

## [0.1.0](https://github.com/kudoas/sync-issue-field/compare/v0.0.8...v0.1.0) (2024-04-30)


### Features

* validate env config ([#33](https://github.com/kudoas/sync-issue-field/issues/33)) ([7efd80f](https://github.com/kudoas/sync-issue-field/commit/7efd80ff5c5c9708958f6f983352334f0a0914e6))


### Bug Fixes

* **deps:** update github.com/shurcool/githubv4 digest to be2daab ([#32](https://github.com/kudoas/sync-issue-field/issues/32)) ([c6bd06f](https://github.com/kudoas/sync-issue-field/commit/c6bd06f97010de89705bf0a35962dcc9830b8d4b))


### Performance Improvements

* **request:** get issue information in a single request ([#30](https://github.com/kudoas/sync-issue-field/issues/30)) ([2240b79](https://github.com/kudoas/sync-issue-field/commit/2240b79ac6766f7687521637444d6917da6016c4))

## [0.0.8](https://github.com/kudoas/sync-issue-field/compare/v0.0.7...v0.0.8) (2024-04-23)


### Bug Fixes

* move run command ([#27](https://github.com/kudoas/sync-issue-field/issues/27)) ([808f84f](https://github.com/kudoas/sync-issue-field/commit/808f84f2a33c2acc37fdec08b63c43d1d2b1ef46))
* split github client package ([#24](https://github.com/kudoas/sync-issue-field/issues/24)) ([612fc87](https://github.com/kudoas/sync-issue-field/commit/612fc87672da69cdc629911ed83d7cf9b5af3187))

## [0.0.7](https://github.com/kudoas/sync-issue-field/compare/v0.0.6...v0.0.7) (2024-04-21)


### Bug Fixes

* use return early ([#19](https://github.com/kudoas/sync-issue-field/issues/19)) ([282d275](https://github.com/kudoas/sync-issue-field/commit/282d275d997b06fe4a0b2e944e27e9cf6f5f2b5b))

## [0.0.6](https://github.com/kudoas/sync-issue-field/compare/v0.0.5...v0.0.6) (2024-04-21)


### Bug Fixes

* **ci:** fix input ([#16](https://github.com/kudoas/sync-issue-field/issues/16)) ([54b51f8](https://github.com/kudoas/sync-issue-field/commit/54b51f8b4f278037e33958b1c359e98e4cec344f))
* **deps:** update github.com/shurcool/githubv4 digest to 18a1ae0 ([#9](https://github.com/kudoas/sync-issue-field/issues/9)) ([9f23f94](https://github.com/kudoas/sync-issue-field/commit/9f23f948f0bd0b73ce3923a4e12766076fa5c636))
* **deps:** update module golang.org/x/oauth2 to v0.19.0 ([#14](https://github.com/kudoas/sync-issue-field/issues/14)) ([b20db71](https://github.com/kudoas/sync-issue-field/commit/b20db71189c657a8a04fdc1e6e43ba422ecf2a01))
