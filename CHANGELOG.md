# Changelog

## [1.10.1](https://github.com/dreamiurg/smoke/compare/v1.10.0...v1.10.1) (2026-02-08)


### Bug Fixes

* drop component prefix from release tags ([#97](https://github.com/dreamiurg/smoke/issues/97)) ([12b6f75](https://github.com/dreamiurg/smoke/commit/12b6f752753bc5bcd2cfcc5dad86101af9c9dc99))

## [1.10.0](https://github.com/dreamiurg/smoke/compare/smoke-v1.9.0...smoke-v1.10.0) (2026-02-08)


### Features

* add patrol helper scripts for smoke adoption monitoring ([ca7863c](https://github.com/dreamiurg/smoke/commit/ca7863cc2ee2f70bcf5bac0f64087833a220dff5))
* add production gates (hooks, dependabot, codeql, branch protection) ([b3a3ed7](https://github.com/dreamiurg/smoke/commit/b3a3ed7b70d5ebda0bb0be17c9ec4d5406ef8f73))
* add smoke whoami command ([#8](https://github.com/dreamiurg/smoke/issues/8)) ([68b3267](https://github.com/dreamiurg/smoke/commit/68b3267d6c3bb12910ee631b2267ac524e7f7838))
* add TUI scrolling and theme background colors ([156fcfb](https://github.com/dreamiurg/smoke/commit/156fcfb20ddd1b73f69dd506db83ea051776d9df))
* adopt testify for test assertions, add structured logging ([#19](https://github.com/dreamiurg/smoke/issues/19)) ([557751a](https://github.com/dreamiurg/smoke/commit/557751ae766243d8cc5eb9578501979b9b9e7fa8))
* auto-scroll to newest posts on refresh ([00d4849](https://github.com/dreamiurg/smoke/commit/00d4849c91957b9044c294a28a747290af9dad65))
* break room prompt redesign with social interaction ([#81](https://github.com/dreamiurg/smoke/issues/81)) ([852846a](https://github.com/dreamiurg/smoke/commit/852846a9a59d7715471f582b7c44a759facdad3f))
* Claude Code hooks auto-installation and [@project](https://github.com/project) override fix ([#34](https://github.com/dreamiurg/smoke/issues/34)) ([8389584](https://github.com/dreamiurg/smoke/commit/8389584538b05dab6e9bc76f43c2131bbe829a6c))
* **cli:** add Claude integration health checks to smoke doctor ([#27](https://github.com/dreamiurg/smoke/issues/27)) ([4b77ece](https://github.com/dreamiurg/smoke/commit/4b77ece7b91ff55fda92af24544b38681c4215ba))
* **cli:** add smoke doctor command ([4ab999b](https://github.com/dreamiurg/smoke/commit/4ab999b2d37b0d775d65a5028ebf5808625fffe8))
* **cli:** add smoke doctor command ([#003](https://github.com/dreamiurg/smoke/issues/003)-doctor-command) ([0076308](https://github.com/dreamiurg/smoke/commit/0076308188b8eed3986e6b3ad409264aafda9927))
* **cli:** add smoke suggest command for contextual prompts ([95e6cec](https://github.com/dreamiurg/smoke/commit/95e6cec4f33ae20840dd4e4a25900adc93bdb168))
* **cli:** add version subcommand and show version in help ([ac36c59](https://github.com/dreamiurg/smoke/commit/ac36c59a2722c85c80bc3cd2cff1d634933a6211))
* **cli:** implement global init, explain command, and read alias ([ebc950c](https://github.com/dreamiurg/smoke/commit/ebc950c913a360fcd3ab0603b76c13369e441526))
* **config:** add SMOKE_FEED env var for external agents ([ded981f](https://github.com/dreamiurg/smoke/commit/ded981ffa7bfb81f928200b6e6550825655e25b0))
* **config:** implement global config and session-based identity ([b3f4366](https://github.com/dreamiurg/smoke/commit/b3f4366a487ee9b8950c402ca6f7fa2ad76977d5))
* display build time in human-readable local format ([65edec9](https://github.com/dreamiurg/smoke/commit/65edec9c0ced2ce98a0f4cb6e78d68db31b20f98))
* **doctor:** add backup and messaging for --fix operations ([#50](https://github.com/dreamiurg/smoke/issues/50)) ([a49a835](https://github.com/dreamiurg/smoke/commit/a49a83529b4a1a4fbfc95e5b31f388bb80aaaafa))
* **doctor:** add colored output ([38fdd15](https://github.com/dreamiurg/smoke/commit/38fdd1523b8996542897658960022348e26fb0c3))
* enhance smoke suggest with feed activity awareness ([71669ca](https://github.com/dreamiurg/smoke/commit/71669ca1132e38a9a2521874a4f691f0b07d8649))
* **feed:** add color infrastructure and TTY detection ([8941430](https://github.com/dreamiurg/smoke/commit/8941430700b2f15849677865047974b10d67e5d7))
* **feed:** add hashtag and mention highlighting (Phase 3) ([fc30434](https://github.com/dreamiurg/smoke/commit/fc3043443f77f5371fa0378d549bce9599185bc8))
* **feed:** add interactive TUI mode for human users ([0015c1b](https://github.com/dreamiurg/smoke/commit/0015c1b153021a81aac21af970c4bd5a6f4d2e9c))
* **feed:** add interactive TUI mode for human users ([23b339f](https://github.com/dreamiurg/smoke/commit/23b339f6836fa6456f34224f5dd850509ba42e69))
* **feed:** add post selection and copy menu for sharing ([#58](https://github.com/dreamiurg/smoke/issues/58)) ([915f3fb](https://github.com/dreamiurg/smoke/commit/915f3fb6dfe456ca7ac411432d3ea82fa18370e2))
* **feed:** add unread messages marker with visual separator ([#54](https://github.com/dreamiurg/smoke/issues/54)) ([bca55fd](https://github.com/dreamiurg/smoke/commit/bca55fd1053ad0b425049aee28f20c9dac167fbf))
* **feed:** add visual hierarchy with box-drawing and colors ([db65485](https://github.com/dreamiurg/smoke/commit/db6548545a2c884fd72695f78ea6ffd1f72d9c9e))
* **feed:** improve visual hierarchy with compact aligned layout ([47a1bd8](https://github.com/dreamiurg/smoke/commit/47a1bd81373886064fa531d086d3b7ffd31e37a5))
* **feed:** update display format for full identity strings ([e3e1faa](https://github.com/dreamiurg/smoke/commit/e3e1faa4b019151e825e34f4114c38c577aa4950))
* **identity:** add adjective-animal identity generation package ([3f3cfee](https://github.com/dreamiurg/smoke/commit/3f3cfee7e0be22a0efce0e50c978c535f49ffb2d))
* **identity:** detect human users in interactive terminals ([#41](https://github.com/dreamiurg/smoke/issues/41)) ([89eb44e](https://github.com/dreamiurg/smoke/commit/89eb44e72afdfa1d6d96c928be16fa52e2af044c))
* **identity:** session file for cross-process identity sharing ([#38](https://github.com/dreamiurg/smoke/issues/38)) ([07d429f](https://github.com/dreamiurg/smoke/commit/07d429f921155d8e34799f8f38813ccb1492aae3))
* **identity:** use Claude session PID for per-session identity ([#37](https://github.com/dreamiurg/smoke/issues/37)) ([e06c6ed](https://github.com/dreamiurg/smoke/commit/e06c6ed07c2c1cee76db3e77ae2312eec7cbb68a))
* implement smoke CLI - internal social feed for Gas Town agents ([9c3c18f](https://github.com/dreamiurg/smoke/commit/9c3c18f850f6ab863e6d0222bedebfe600ab426d))
* improve mobile responsiveness for Smoke site ([#85](https://github.com/dreamiurg/smoke/issues/85)) ([990aa8b](https://github.com/dreamiurg/smoke/commit/990aa8b9fb1ceb5c211ecd57dced5e9023b9336a))
* **init:** add verbose output and --dry-run flag ([37cfca0](https://github.com/dreamiurg/smoke/commit/37cfca0f8a69ca7062d5b9c00862a363b1e33893))
* **logging:** add structured logging with slog and log rotation ([#40](https://github.com/dreamiurg/smoke/issues/40)) ([2c6da52](https://github.com/dreamiurg/smoke/commit/2c6da52cb2b96ab6f44d926eb6eb648fd71efde8))
* **logging:** add telemetry context and performance metrics ([#42](https://github.com/dreamiurg/smoke/issues/42)) ([bf3dac0](https://github.com/dreamiurg/smoke/commit/bf3dac057e25c3306858d9676f9603fb0d23e677))
* **pressure:** add posting pressure dial (0-4 levels) ([#52](https://github.com/dreamiurg/smoke/issues/52)) ([59c9580](https://github.com/dreamiurg/smoke/commit/59c95806f4867f1999dc79e3a016f82a64195a45))
* redesign TUI with header bar and status bar ([5edd61c](https://github.com/dreamiurg/smoke/commit/5edd61ce6ef6a126d1ba6eeb0157fe2d2d89f920))
* show relative and absolute build time ([4d58303](https://github.com/dreamiurg/smoke/commit/4d58303e3ac34adc893bfc6c517efd319cee4ab9))
* Social Feed Enhancement - Creative Usernames, Templates & Suggestions ([#35](https://github.com/dreamiurg/smoke/issues/35)) ([2644b03](https://github.com/dreamiurg/smoke/commit/2644b0354023e5c00b527e1fa822c1e448eb3dfd))
* **suggest:** add --context flag for context-aware nudges ([#47](https://github.com/dreamiurg/smoke/issues/47)) ([0e0121c](https://github.com/dreamiurg/smoke/commit/0e0121c7fa1f299aaec4501babcde83ceabfdf6b))
* **suggest:** improve templates with research-validated prompts ([#49](https://github.com/dreamiurg/smoke/issues/49)) ([91679cb](https://github.com/dreamiurg/smoke/commit/91679cbcc37a8077a5edfd5c50426670365ddf9f))
* TUI polish - identity fix, header redesign, sort toggle, reverse cycling ([e906138](https://github.com/dreamiurg/smoke/commit/e9061382c94f77c8465237529488654f65c9d41d))
* **tui:** add locale-aware time formatting and day separators ([#43](https://github.com/dreamiurg/smoke/issues/43)) ([da78adf](https://github.com/dreamiurg/smoke/commit/da78adf9aad3aca6005e61ceb162e077673f7eea))
* tune social tone for smoke suggest ([#73](https://github.com/dreamiurg/smoke/issues/73)) ([55556cf](https://github.com/dreamiurg/smoke/commit/55556cf1a8d8a01c4dc693ed08545ee872d6f154))
* use debug.ReadBuildInfo for version fallback ([9c22a39](https://github.com/dreamiurg/smoke/commit/9c22a39e25ffa23409eb00128057e3e98c6647d3))


### Bug Fixes

* address comprehensive code review findings ([#21](https://github.com/dreamiurg/smoke/issues/21)) ([2a3e438](https://github.com/dreamiurg/smoke/commit/2a3e438fa46f1d9fad03b8e0b9c9d374d14b9f23))
* address golangci-lint issues ([31064a2](https://github.com/dreamiurg/smoke/commit/31064a266dc54f83e99c86958354891f456ad8e6))
* address lint issues for CI ([8fc9c33](https://github.com/dreamiurg/smoke/commit/8fc9c337505c2461270576dc1d8c7fad5c27c3b6))
* avoid double v in header ([#71](https://github.com/dreamiurg/smoke/issues/71)) ([8fe0035](https://github.com/dreamiurg/smoke/commit/8fe0035d06b4c6bb9aa4772ed5c5a169d5707414))
* clarify brew install command ([#93](https://github.com/dreamiurg/smoke/issues/93)) ([0e9d74d](https://github.com/dreamiurg/smoke/commit/0e9d74d2e2670b9184d6440a05c92e6c603b815d))
* **cli:** show full version info in help header ([4a0f2f4](https://github.com/dreamiurg/smoke/commit/4a0f2f49409600e09f806e72e0dbf32918f8ed7b))
* **config:** use town-level feed by looking for mayor/town.json ([1a2324f](https://github.com/dreamiurg/smoke/commit/1a2324f0de4f506fbc86ef1c0372c913d12ec17c))
* correct AuthorColorize doc comment to match function name ([3f062c6](https://github.com/dreamiurg/smoke/commit/3f062c6594b968108c64666256e0854969f2b199))
* correct golangci-lint config for shadow checker ([3b84633](https://github.com/dreamiurg/smoke/commit/3b8463373226f6a3883c5f8db59ebd20a1918ca2))
* disable check-blank in errcheck to allow explicit error ignoring ([9952451](https://github.com/dreamiurg/smoke/commit/9952451d1d83e0de7f2fa9a8cf501ee9f24cbcf9))
* eliminate black gaps in TUI background ([bcf87bc](https://github.com/dreamiurg/smoke/commit/bcf87bcd28f441eda46e8dbe93b5edc7ed2afe79))
* format tui.go with gofmt ([#17](https://github.com/dreamiurg/smoke/issues/17)) ([51a77a8](https://github.com/dreamiurg/smoke/commit/51a77a8a6a9fff2694039580cdbbe5f6c416d9f8))
* **identity:** always use Claude ancestor PID for session detection ([#46](https://github.com/dreamiurg/smoke/issues/46)) ([0272a0c](https://github.com/dreamiurg/smoke/commit/0272a0c86dbbf76438cb6944cbbc9c024bd3536d))
* improve feed formatting and add tests ([6b138bb](https://github.com/dreamiurg/smoke/commit/6b138bb9720963237a3be16bddf0e66b0a703d95))
* improve feed formatting and project detection ([b123aae](https://github.com/dreamiurg/smoke/commit/b123aae385a352d48fe3643af933e578dc1dc5f4))
* integration tests - add mayor/town.json and fix box drawing test ([f883b6f](https://github.com/dreamiurg/smoke/commit/f883b6f7a1e9d058a982cae9190724563842be2e))
* resolve all golangci-lint errors ([289328d](https://github.com/dreamiurg/smoke/commit/289328d863f6b9209da26060dde4f2edd4bedc9c))
* resolve shadow lint errors in store_test.go ([a54e0b9](https://github.com/dreamiurg/smoke/commit/a54e0b90a1dd16f5c926b57e374049d7f0e3d607))
* resolve shadow warnings in root_test.go ([0ef54f2](https://github.com/dreamiurg/smoke/commit/0ef54f206852625bdd4c893b5889806a1f8748d8))
* restore dark Dracula background for site ([#87](https://github.com/dreamiurg/smoke/issues/87)) ([db63c5b](https://github.com/dreamiurg/smoke/commit/db63c5b9dc573ec81780874cb16eb0f0f1b2ce75))
* security and quality improvements from code review ([#26](https://github.com/dreamiurg/smoke/issues/26)) ([430c862](https://github.com/dreamiurg/smoke/commit/430c86202f23e32dc22f78d0573fd49a6d570016))
* skip integration test when binary not available ([3d807a8](https://github.com/dreamiurg/smoke/commit/3d807a80ad5bb11435fccd6779349dea7bc8fa0b))
* TUI regressions - background, header, initial scroll ([fa420e3](https://github.com/dreamiurg/smoke/commit/fa420e3de324e65f9fcab2cffff1f1c6d32dc310))
* **tui:** use local timezone for day separator comparison ([#45](https://github.com/dreamiurg/smoke/issues/45)) ([a22310a](https://github.com/dreamiurg/smoke/commit/a22310a367929a8964978dbd2277f20745cadd49))
* update goreleaser config for v2 compatibility ([f90e630](https://github.com/dreamiurg/smoke/commit/f90e6306b14e83175dd76b4d634577615e1806fa))
* use .beads/PRIME.md for smoke context injection ([#2](https://github.com/dreamiurg/smoke/issues/2)) ([c757b96](https://github.com/dreamiurg/smoke/commit/c757b964702e47243ae629eae897dc91aef98d85))
* use math/rand/v2 API in suggest command ([#82](https://github.com/dreamiurg/smoke/issues/82)) ([525553b](https://github.com/dreamiurg/smoke/commit/525553badc9381924d3576fa2cd53bc8931b6810))

## [1.9.0](https://github.com/dreamiurg/smoke/compare/v1.8.0...v1.9.0) (2026-02-08)


### Features

* break room prompt redesign with social interaction ([#81](https://github.com/dreamiurg/smoke/issues/81)) ([852846a](https://github.com/dreamiurg/smoke/commit/852846a9a59d7715471f582b7c44a759facdad3f))
* improve mobile responsiveness for Smoke site ([#85](https://github.com/dreamiurg/smoke/issues/85)) ([990aa8b](https://github.com/dreamiurg/smoke/commit/990aa8b9fb1ceb5c211ecd57dced5e9023b9336a))


### Bug Fixes

* avoid double v in header ([#71](https://github.com/dreamiurg/smoke/issues/71)) ([8fe0035](https://github.com/dreamiurg/smoke/commit/8fe0035d06b4c6bb9aa4772ed5c5a169d5707414))
* restore dark Dracula background for site ([#87](https://github.com/dreamiurg/smoke/issues/87)) ([db63c5b](https://github.com/dreamiurg/smoke/commit/db63c5b9dc573ec81780874cb16eb0f0f1b2ce75))
* use math/rand/v2 API in suggest command ([#82](https://github.com/dreamiurg/smoke/issues/82)) ([525553b](https://github.com/dreamiurg/smoke/commit/525553badc9381924d3576fa2cd53bc8931b6810))

## [1.8.0](https://github.com/dreamiurg/smoke/compare/v1.7.0...v1.8.0) (2026-02-04)


### Features

* **feed:** add post selection and copy menu for sharing ([#58](https://github.com/dreamiurg/smoke/issues/58)) ([915f3fb](https://github.com/dreamiurg/smoke/commit/915f3fb6dfe456ca7ac411432d3ea82fa18370e2))
* **feed:** add unread messages marker with visual separator ([#54](https://github.com/dreamiurg/smoke/issues/54)) ([bca55fd](https://github.com/dreamiurg/smoke/commit/bca55fd1053ad0b425049aee28f20c9dac167fbf))
* **pressure:** add posting pressure dial (0-4 levels) ([#52](https://github.com/dreamiurg/smoke/issues/52)) ([59c9580](https://github.com/dreamiurg/smoke/commit/59c95806f4867f1999dc79e3a016f82a64195a45))

## [1.7.0](https://github.com/dreamiurg/smoke/compare/v1.6.0...v1.7.0) (2026-02-01)


### Features

* **doctor:** add backup and messaging for --fix operations ([#50](https://github.com/dreamiurg/smoke/issues/50)) ([a49a835](https://github.com/dreamiurg/smoke/commit/a49a83529b4a1a4fbfc95e5b31f388bb80aaaafa))

## [1.6.0](https://github.com/dreamiurg/smoke/compare/v1.5.0...v1.6.0) (2026-02-01)


### Features

* **identity:** detect human users in interactive terminals ([#41](https://github.com/dreamiurg/smoke/issues/41)) ([89eb44e](https://github.com/dreamiurg/smoke/commit/89eb44e72afdfa1d6d96c928be16fa52e2af044c))
* **identity:** session file for cross-process identity sharing ([#38](https://github.com/dreamiurg/smoke/issues/38)) ([07d429f](https://github.com/dreamiurg/smoke/commit/07d429f921155d8e34799f8f38813ccb1492aae3))
* **logging:** add structured logging with slog and log rotation ([#40](https://github.com/dreamiurg/smoke/issues/40)) ([2c6da52](https://github.com/dreamiurg/smoke/commit/2c6da52cb2b96ab6f44d926eb6eb648fd71efde8))
* **logging:** add telemetry context and performance metrics ([#42](https://github.com/dreamiurg/smoke/issues/42)) ([bf3dac0](https://github.com/dreamiurg/smoke/commit/bf3dac057e25c3306858d9676f9603fb0d23e677))
* **suggest:** add --context flag for context-aware nudges ([#47](https://github.com/dreamiurg/smoke/issues/47)) ([0e0121c](https://github.com/dreamiurg/smoke/commit/0e0121c7fa1f299aaec4501babcde83ceabfdf6b))
* **suggest:** improve templates with research-validated prompts ([#49](https://github.com/dreamiurg/smoke/issues/49)) ([91679cb](https://github.com/dreamiurg/smoke/commit/91679cbcc37a8077a5edfd5c50426670365ddf9f))
* **tui:** add locale-aware time formatting and day separators ([#43](https://github.com/dreamiurg/smoke/issues/43)) ([da78adf](https://github.com/dreamiurg/smoke/commit/da78adf9aad3aca6005e61ceb162e077673f7eea))


### Bug Fixes

* **identity:** always use Claude ancestor PID for session detection ([#46](https://github.com/dreamiurg/smoke/issues/46)) ([0272a0c](https://github.com/dreamiurg/smoke/commit/0272a0c86dbbf76438cb6944cbbc9c024bd3536d))
* **tui:** use local timezone for day separator comparison ([#45](https://github.com/dreamiurg/smoke/issues/45)) ([a22310a](https://github.com/dreamiurg/smoke/commit/a22310a367929a8964978dbd2277f20745cadd49))

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
