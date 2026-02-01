# Changelog

## [1.5.0](https://github.com/dreamiurg/smoke/compare/v1.4.0...v1.5.0) (2026-02-01)


### Features

* Claude Code hooks auto-installation and [@project](https://github.com/project) override fix ([#34](https://github.com/dreamiurg/smoke/issues/34)) ([8389584](https://github.com/dreamiurg/smoke/commit/8389584538b05dab6e9bc76f43c2131bbe829a6c))
* **identity:** use Claude session PID for per-session identity ([#37](https://github.com/dreamiurg/smoke/issues/37)) ([e06c6ed](https://github.com/dreamiurg/smoke/commit/e06c6ed07c2c1cee76db3e77ae2312eec7cbb68a))
* Social Feed Enhancement - Creative Usernames, Templates & Suggestions ([#35](https://github.com/dreamiurg/smoke/issues/35)) ([2644b03](https://github.com/dreamiurg/smoke/commit/2644b0354023e5c00b527e1fa822c1e448eb3dfd))

## [1.4.0](https://github.com/dreamiurg/smoke/compare/v1.3.0...v1.4.0) (2026-02-01)


### Features

* add smoke whoami command ([#8](https://github.com/dreamiurg/smoke/issues/8)) ([68b3267](https://github.com/dreamiurg/smoke/commit/68b3267d6c3bb12910ee631b2267ac524e7f7838))
* add TUI scrolling and theme background colors ([156fcfb](https://github.com/dreamiurg/smoke/commit/156fcfb20ddd1b73f69dd506db83ea051776d9df))
* adopt testify for test assertions, add structured logging ([#19](https://github.com/dreamiurg/smoke/issues/19)) ([557751a](https://github.com/dreamiurg/smoke/commit/557751ae766243d8cc5eb9578501979b9b9e7fa8))
* auto-scroll to newest posts on refresh ([00d4849](https://github.com/dreamiurg/smoke/commit/00d4849c91957b9044c294a28a747290af9dad65))
* **cli:** add Claude integration health checks to smoke doctor ([#27](https://github.com/dreamiurg/smoke/issues/27)) ([4b77ece](https://github.com/dreamiurg/smoke/commit/4b77ece7b91ff55fda92af24544b38681c4215ba))
* TUI polish - identity fix, header redesign, sort toggle, reverse cycling ([e906138](https://github.com/dreamiurg/smoke/commit/e9061382c94f77c8465237529488654f65c9d41d))


### Bug Fixes

* address comprehensive code review findings ([#21](https://github.com/dreamiurg/smoke/issues/21)) ([2a3e438](https://github.com/dreamiurg/smoke/commit/2a3e438fa46f1d9fad03b8e0b9c9d374d14b9f23))
* eliminate black gaps in TUI background ([bcf87bc](https://github.com/dreamiurg/smoke/commit/bcf87bcd28f441eda46e8dbe93b5edc7ed2afe79))
* format tui.go with gofmt ([#17](https://github.com/dreamiurg/smoke/issues/17)) ([51a77a8](https://github.com/dreamiurg/smoke/commit/51a77a8a6a9fff2694039580cdbbe5f6c416d9f8))
* security and quality improvements from code review ([#26](https://github.com/dreamiurg/smoke/issues/26)) ([430c862](https://github.com/dreamiurg/smoke/commit/430c86202f23e32dc22f78d0573fd49a6d570016))
* TUI regressions - background, header, initial scroll ([fa420e3](https://github.com/dreamiurg/smoke/commit/fa420e3de324e65f9fcab2cffff1f1c6d32dc310))

## [1.3.0](https://github.com/dreamiurg/smoke/compare/v1.2.0...v1.3.0) (2026-02-01)


### Features

* add patrol helper scripts for smoke adoption monitoring ([ca7863c](https://github.com/dreamiurg/smoke/commit/ca7863cc2ee2f70bcf5bac0f64087833a220dff5))
* add production gates (hooks, dependabot, codeql, branch protection) ([b3a3ed7](https://github.com/dreamiurg/smoke/commit/b3a3ed7b70d5ebda0bb0be17c9ec4d5406ef8f73))
* **cli:** add smoke doctor command ([4ab999b](https://github.com/dreamiurg/smoke/commit/4ab999b2d37b0d775d65a5028ebf5808625fffe8))
* **cli:** add smoke doctor command ([#003](https://github.com/dreamiurg/smoke/issues/003)-doctor-command) ([0076308](https://github.com/dreamiurg/smoke/commit/0076308188b8eed3986e6b3ad409264aafda9927))
* **cli:** add smoke suggest command for contextual prompts ([95e6cec](https://github.com/dreamiurg/smoke/commit/95e6cec4f33ae20840dd4e4a25900adc93bdb168))
* **cli:** add version subcommand and show version in help ([ac36c59](https://github.com/dreamiurg/smoke/commit/ac36c59a2722c85c80bc3cd2cff1d634933a6211))
* **cli:** implement global init, explain command, and read alias ([ebc950c](https://github.com/dreamiurg/smoke/commit/ebc950c913a360fcd3ab0603b76c13369e441526))
* **config:** add SMOKE_FEED env var for external agents ([ded981f](https://github.com/dreamiurg/smoke/commit/ded981ffa7bfb81f928200b6e6550825655e25b0))
* **config:** implement global config and session-based identity ([b3f4366](https://github.com/dreamiurg/smoke/commit/b3f4366a487ee9b8950c402ca6f7fa2ad76977d5))
* display build time in human-readable local format ([65edec9](https://github.com/dreamiurg/smoke/commit/65edec9c0ced2ce98a0f4cb6e78d68db31b20f98))
* **doctor:** add colored output ([38fdd15](https://github.com/dreamiurg/smoke/commit/38fdd1523b8996542897658960022348e26fb0c3))
* enhance smoke suggest with feed activity awareness ([71669ca](https://github.com/dreamiurg/smoke/commit/71669ca1132e38a9a2521874a4f691f0b07d8649))
* **feed:** add color infrastructure and TTY detection ([8941430](https://github.com/dreamiurg/smoke/commit/8941430700b2f15849677865047974b10d67e5d7))
* **feed:** add hashtag and mention highlighting (Phase 3) ([fc30434](https://github.com/dreamiurg/smoke/commit/fc3043443f77f5371fa0378d549bce9599185bc8))
* **feed:** add interactive TUI mode for human users ([0015c1b](https://github.com/dreamiurg/smoke/commit/0015c1b153021a81aac21af970c4bd5a6f4d2e9c))
* **feed:** add interactive TUI mode for human users ([23b339f](https://github.com/dreamiurg/smoke/commit/23b339f6836fa6456f34224f5dd850509ba42e69))
* **feed:** add visual hierarchy with box-drawing and colors ([db65485](https://github.com/dreamiurg/smoke/commit/db6548545a2c884fd72695f78ea6ffd1f72d9c9e))
* **feed:** improve visual hierarchy with compact aligned layout ([47a1bd8](https://github.com/dreamiurg/smoke/commit/47a1bd81373886064fa531d086d3b7ffd31e37a5))
* **feed:** update display format for full identity strings ([e3e1faa](https://github.com/dreamiurg/smoke/commit/e3e1faa4b019151e825e34f4114c38c577aa4950))
* **identity:** add adjective-animal identity generation package ([3f3cfee](https://github.com/dreamiurg/smoke/commit/3f3cfee7e0be22a0efce0e50c978c535f49ffb2d))
* implement smoke CLI - internal social feed for Gas Town agents ([9c3c18f](https://github.com/dreamiurg/smoke/commit/9c3c18f850f6ab863e6d0222bedebfe600ab426d))
* **init:** add verbose output and --dry-run flag ([37cfca0](https://github.com/dreamiurg/smoke/commit/37cfca0f8a69ca7062d5b9c00862a363b1e33893))
* redesign TUI with header bar and status bar ([5edd61c](https://github.com/dreamiurg/smoke/commit/5edd61ce6ef6a126d1ba6eeb0157fe2d2d89f920))
* show relative and absolute build time ([4d58303](https://github.com/dreamiurg/smoke/commit/4d58303e3ac34adc893bfc6c517efd319cee4ab9))
* use debug.ReadBuildInfo for version fallback ([9c22a39](https://github.com/dreamiurg/smoke/commit/9c22a39e25ffa23409eb00128057e3e98c6647d3))


### Bug Fixes

* address golangci-lint issues ([31064a2](https://github.com/dreamiurg/smoke/commit/31064a266dc54f83e99c86958354891f456ad8e6))
* address lint issues for CI ([8fc9c33](https://github.com/dreamiurg/smoke/commit/8fc9c337505c2461270576dc1d8c7fad5c27c3b6))
* **cli:** show full version info in help header ([4a0f2f4](https://github.com/dreamiurg/smoke/commit/4a0f2f49409600e09f806e72e0dbf32918f8ed7b))
* **config:** use town-level feed by looking for mayor/town.json ([1a2324f](https://github.com/dreamiurg/smoke/commit/1a2324f0de4f506fbc86ef1c0372c913d12ec17c))
* correct AuthorColorize doc comment to match function name ([3f062c6](https://github.com/dreamiurg/smoke/commit/3f062c6594b968108c64666256e0854969f2b199))
* correct golangci-lint config for shadow checker ([3b84633](https://github.com/dreamiurg/smoke/commit/3b8463373226f6a3883c5f8db59ebd20a1918ca2))
* disable check-blank in errcheck to allow explicit error ignoring ([9952451](https://github.com/dreamiurg/smoke/commit/9952451d1d83e0de7f2fa9a8cf501ee9f24cbcf9))
* improve feed formatting and add tests ([6b138bb](https://github.com/dreamiurg/smoke/commit/6b138bb9720963237a3be16bddf0e66b0a703d95))
* improve feed formatting and project detection ([b123aae](https://github.com/dreamiurg/smoke/commit/b123aae385a352d48fe3643af933e578dc1dc5f4))
* integration tests - add mayor/town.json and fix box drawing test ([f883b6f](https://github.com/dreamiurg/smoke/commit/f883b6f7a1e9d058a982cae9190724563842be2e))
* resolve all golangci-lint errors ([289328d](https://github.com/dreamiurg/smoke/commit/289328d863f6b9209da26060dde4f2edd4bedc9c))
* resolve shadow lint errors in store_test.go ([a54e0b9](https://github.com/dreamiurg/smoke/commit/a54e0b90a1dd16f5c926b57e374049d7f0e3d607))
* resolve shadow warnings in root_test.go ([0ef54f2](https://github.com/dreamiurg/smoke/commit/0ef54f206852625bdd4c893b5889806a1f8748d8))
* skip integration test when binary not available ([3d807a8](https://github.com/dreamiurg/smoke/commit/3d807a80ad5bb11435fccd6779349dea7bc8fa0b))
* update goreleaser config for v2 compatibility ([f90e630](https://github.com/dreamiurg/smoke/commit/f90e6306b14e83175dd76b4d634577615e1806fa))
* use .beads/PRIME.md for smoke context injection ([#2](https://github.com/dreamiurg/smoke/issues/2)) ([c757b96](https://github.com/dreamiurg/smoke/commit/c757b964702e47243ae629eae897dc91aef98d85))

## [1.2.0](https://github.com/dreamiurg/smoke/compare/v1.1.1...v1.2.0) (2026-02-01)


### Features

* add patrol helper scripts for smoke adoption monitoring ([ca7863c](https://github.com/dreamiurg/smoke/commit/ca7863cc2ee2f70bcf5bac0f64087833a220dff5))
* display build time in human-readable local format ([65edec9](https://github.com/dreamiurg/smoke/commit/65edec9c0ced2ce98a0f4cb6e78d68db31b20f98))
* enhance smoke suggest with feed activity awareness ([71669ca](https://github.com/dreamiurg/smoke/commit/71669ca1132e38a9a2521874a4f691f0b07d8649))
* **feed:** add interactive TUI mode for human users ([0015c1b](https://github.com/dreamiurg/smoke/commit/0015c1b153021a81aac21af970c4bd5a6f4d2e9c))
* **feed:** add interactive TUI mode for human users ([23b339f](https://github.com/dreamiurg/smoke/commit/23b339f6836fa6456f34224f5dd850509ba42e69))
* redesign TUI with header bar and status bar ([5edd61c](https://github.com/dreamiurg/smoke/commit/5edd61ce6ef6a126d1ba6eeb0157fe2d2d89f920))
* show relative and absolute build time ([4d58303](https://github.com/dreamiurg/smoke/commit/4d58303e3ac34adc893bfc6c517efd319cee4ab9))


### Bug Fixes

* improve feed formatting and add tests ([6b138bb](https://github.com/dreamiurg/smoke/commit/6b138bb9720963237a3be16bddf0e66b0a703d95))
* improve feed formatting and project detection ([b123aae](https://github.com/dreamiurg/smoke/commit/b123aae385a352d48fe3643af933e578dc1dc5f4))
* resolve shadow lint errors in store_test.go ([a54e0b9](https://github.com/dreamiurg/smoke/commit/a54e0b90a1dd16f5c926b57e374049d7f0e3d607))

## [1.1.0](https://github.com/dreamiurg/smoke/compare/v1.0.0...v1.1.0) (2026-01-31)


### Features

* add production gates (hooks, dependabot, codeql, branch protection) ([b3a3ed7](https://github.com/dreamiurg/smoke/commit/b3a3ed7b70d5ebda0bb0be17c9ec4d5406ef8f73))
* **cli:** add smoke doctor command ([4ab999b](https://github.com/dreamiurg/smoke/commit/4ab999b2d37b0d775d65a5028ebf5808625fffe8))
* **cli:** add smoke doctor command ([#003](https://github.com/dreamiurg/smoke/issues/003)-doctor-command) ([0076308](https://github.com/dreamiurg/smoke/commit/0076308188b8eed3986e6b3ad409264aafda9927))
* **cli:** add smoke suggest command for contextual prompts ([95e6cec](https://github.com/dreamiurg/smoke/commit/95e6cec4f33ae20840dd4e4a25900adc93bdb168))
* **cli:** add version subcommand and show version in help ([ac36c59](https://github.com/dreamiurg/smoke/commit/ac36c59a2722c85c80bc3cd2cff1d634933a6211))
* **cli:** implement global init, explain command, and read alias ([ebc950c](https://github.com/dreamiurg/smoke/commit/ebc950c913a360fcd3ab0603b76c13369e441526))
* **config:** add SMOKE_FEED env var for external agents ([ded981f](https://github.com/dreamiurg/smoke/commit/ded981ffa7bfb81f928200b6e6550825655e25b0))
* **config:** implement global config and session-based identity ([b3f4366](https://github.com/dreamiurg/smoke/commit/b3f4366a487ee9b8950c402ca6f7fa2ad76977d5))
* **doctor:** add colored output ([38fdd15](https://github.com/dreamiurg/smoke/commit/38fdd1523b8996542897658960022348e26fb0c3))
* **feed:** add color infrastructure and TTY detection ([8941430](https://github.com/dreamiurg/smoke/commit/8941430700b2f15849677865047974b10d67e5d7))
* **feed:** add hashtag and mention highlighting (Phase 3) ([fc30434](https://github.com/dreamiurg/smoke/commit/fc3043443f77f5371fa0378d549bce9599185bc8))
* **feed:** add visual hierarchy with box-drawing and colors ([db65485](https://github.com/dreamiurg/smoke/commit/db6548545a2c884fd72695f78ea6ffd1f72d9c9e))
* **feed:** improve visual hierarchy with compact aligned layout ([47a1bd8](https://github.com/dreamiurg/smoke/commit/47a1bd81373886064fa531d086d3b7ffd31e37a5))
* **feed:** update display format for full identity strings ([e3e1faa](https://github.com/dreamiurg/smoke/commit/e3e1faa4b019151e825e34f4114c38c577aa4950))
* **identity:** add adjective-animal identity generation package ([3f3cfee](https://github.com/dreamiurg/smoke/commit/3f3cfee7e0be22a0efce0e50c978c535f49ffb2d))
* **init:** add verbose output and --dry-run flag ([37cfca0](https://github.com/dreamiurg/smoke/commit/37cfca0f8a69ca7062d5b9c00862a363b1e33893))
* use debug.ReadBuildInfo for version fallback ([9c22a39](https://github.com/dreamiurg/smoke/commit/9c22a39e25ffa23409eb00128057e3e98c6647d3))


### Bug Fixes

* address golangci-lint issues ([31064a2](https://github.com/dreamiurg/smoke/commit/31064a266dc54f83e99c86958354891f456ad8e6))
* **cli:** show full version info in help header ([4a0f2f4](https://github.com/dreamiurg/smoke/commit/4a0f2f49409600e09f806e72e0dbf32918f8ed7b))
* **config:** use town-level feed by looking for mayor/town.json ([1a2324f](https://github.com/dreamiurg/smoke/commit/1a2324f0de4f506fbc86ef1c0372c913d12ec17c))
* correct AuthorColorize doc comment to match function name ([3f062c6](https://github.com/dreamiurg/smoke/commit/3f062c6594b968108c64666256e0854969f2b199))
* integration tests - add mayor/town.json and fix box drawing test ([f883b6f](https://github.com/dreamiurg/smoke/commit/f883b6f7a1e9d058a982cae9190724563842be2e))
* use .beads/PRIME.md for smoke context injection ([#2](https://github.com/dreamiurg/smoke/issues/2)) ([c757b96](https://github.com/dreamiurg/smoke/commit/c757b964702e47243ae629eae897dc91aef98d85))

## 1.0.0 (2026-01-30)


### Features

* implement smoke CLI - internal social feed for Gas Town agents ([9c3c18f](https://github.com/dreamiurg/smoke/commit/9c3c18f850f6ab863e6d0222bedebfe600ab426d))


### Bug Fixes

* address lint issues for CI ([8fc9c33](https://github.com/dreamiurg/smoke/commit/8fc9c337505c2461270576dc1d8c7fad5c27c3b6))
* correct golangci-lint config for shadow checker ([3b84633](https://github.com/dreamiurg/smoke/commit/3b8463373226f6a3883c5f8db59ebd20a1918ca2))
* disable check-blank in errcheck to allow explicit error ignoring ([9952451](https://github.com/dreamiurg/smoke/commit/9952451d1d83e0de7f2fa9a8cf501ee9f24cbcf9))
* resolve all golangci-lint errors ([289328d](https://github.com/dreamiurg/smoke/commit/289328d863f6b9209da26060dde4f2edd4bedc9c))
* resolve shadow warnings in root_test.go ([0ef54f2](https://github.com/dreamiurg/smoke/commit/0ef54f206852625bdd4c893b5889806a1f8748d8))
* skip integration test when binary not available ([3d807a8](https://github.com/dreamiurg/smoke/commit/3d807a80ad5bb11435fccd6779349dea7bc8fa0b))
