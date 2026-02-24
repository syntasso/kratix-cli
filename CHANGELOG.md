# Changelog

## [0.14.1](https://github.com/syntasso/kratix-cli/compare/v0.14.0...v0.14.1) (2026-02-24)


### Chores

* add pulumi stage to release pipeline ([ae8e1a3](https://github.com/syntasso/kratix-cli/commit/ae8e1a3c297bd12b7113020f26c67a464626ce5f))

## [0.14.0](https://github.com/syntasso/kratix-cli/compare/v0.13.0...v0.14.0) (2026-02-24)


### Features

* **init-pulumi:** scaffolding for pulumi cli command ([d9be96a](https://github.com/syntasso/kratix-cli/commit/d9be96aa3ff860e1bd65f2e975acb7741a710617))
* load pulumi schema and select component in init command ([13c622a](https://github.com/syntasso/kratix-cli/commit/13c622a38cdfefc6ef56efaac784cf7536b0297b))
* **pulumi-component-to-cli:** add initial Pulumi schema to OpenAPI translation ([b657da9](https://github.com/syntasso/kratix-cli/commit/b657da9e5c87c2e5437efd1e0ef1237dde386b6a))
* **pulumi-component-to-cli:** add resilient translation with component-scoped preflight ([5ae1fbc](https://github.com/syntasso/kratix-cli/commit/5ae1fbc0638b43e65bf2e5f6bd658dda5ae68c22))
* **pulumi-component-to-cli:** add verbose warn/info channels and document output contract ([6567e0e](https://github.com/syntasso/kratix-cli/commit/6567e0e50e871bf50ccb8d81810d95d3164e1b42))
* **pulumi-component-to-cli:** implement CRD identity overrides and derived defaults ([c729c02](https://github.com/syntasso/kratix-cli/commit/c729c029682f0b0bceff90190f4cb004903adbdf))
* **pulumi-component-to-cli:** implement task 1 CLI selection and scaffold emission ([7d3c486](https://github.com/syntasso/kratix-cli/commit/7d3c4865b208f09dadddb435d9b89e068c50ce5e))
* **pulumi-component-to-cli:** stage diagnostics behind --verbose ([e20c042](https://github.com/syntasso/kratix-cli/commit/e20c0424859a978b091ea91bc9ce9d315b88346a))
* **pulumi-component-to-cli:** support URL-based schema inputs ([1413c6e](https://github.com/syntasso/kratix-cli/commit/1413c6e2c1ca2e6cc357fabc4f210de91bc3d8a5))
* **pulumi-component-to-crd:** pass descrptions from component -&gt; crd output ([bab1be8](https://github.com/syntasso/kratix-cli/commit/bab1be89c6f09d9f665d9b18b43aeb38a289382a))
* **pulumi:** Add promise stage runtime and tests ([14233ca](https://github.com/syntasso/kratix-cli/commit/14233ca71cd6803a6ec97dda182d05433fa183ec))
* **pulumi:** generate files for pulumi init ([358ce08](https://github.com/syntasso/kratix-cli/commit/358ce08c130a2aa81fb739365b8ce4c96682d8fb))
* **pulumi:** translate component inputs to CRD spec schema ([5dead2c](https://github.com/syntasso/kratix-cli/commit/5dead2c881f4a50d69bf7db879c4926951c9d9f5))
* **pulumi:** warn when pulumi schema source is local ([571f8bf](https://github.com/syntasso/kratix-cli/commit/571f8bf31bd8615e428109a9a683131218b53cc7))
* wire pulumi stage release metadata and add e2e regression ([0db682a](https://github.com/syntasso/kratix-cli/commit/0db682acfbb545ad4828621adb56e4e16e4bc08f))


### Bug Fixes

* **crossplane:** tests referenced wrong init ([6247f43](https://github.com/syntasso/kratix-cli/commit/6247f4306d3c7966f48a7440bf6bad45b1f6a27d))
* **pulumi-component-to-cli:** support local resource refs in translation ([836f717](https://github.com/syntasso/kratix-cli/commit/836f7173598e5e755843ee2b38f5cfcbc5a99fcd))
* **pulumi-component-to-crd:** manage relative paths better ([cb62f17](https://github.com/syntasso/kratix-cli/commit/cb62f17561486bf777c1078d1924cefc7508dd73))
* **pulumi-translate:** reject unsupported keywords before $ref handling ([46beda0](https://github.com/syntasso/kratix-cli/commit/46beda06a05978842e6d762690e2f648052de5af))
* **pulumi-translate:** validate enum values against declared type ([53ef256](https://github.com/syntasso/kratix-cli/commit/53ef2562dd7f0c3b15b347b0903d0bc063b3f1af))
* **pulumi:** shell-quote README args and omit --dir replay ([d12822b](https://github.com/syntasso/kratix-cli/commit/d12822b1e801bc54f4459d6ffb0d643ad2072ca4))
* **tf:** remove panic on tf-module-init ([ec67918](https://github.com/syntasso/kratix-cli/commit/ec67918e551d2fd9fc57fc37f3a352bad0ff68e4))
* **tf:** remove panic on tf-module-init ([dbbf955](https://github.com/syntasso/kratix-cli/commit/dbbf955c3454d4679b6da6e5c2a7d19d8941eff8))


### Chores

* Add preview warning to init crossplane/helm/operator ([#204](https://github.com/syntasso/kratix-cli/issues/204)) ([1d4e23f](https://github.com/syntasso/kratix-cli/commit/1d4e23f8ad9d39b56ec8f7b7982f5b128b2419cb))
* bump cli version to v0.14.0 ([a7437ae](https://github.com/syntasso/kratix-cli/commit/a7437ae0b3f35a5919273de809dcaf700ca14b11))
* **component-to-crd:** rename to include pulumi in cli name ([f17661e](https://github.com/syntasso/kratix-cli/commit/f17661e8b29739f4a34b67d61dac8683aa768bf0))
* **component-to-crd:** shift manual tests to use more permanent remote test files ([44ccf97](https://github.com/syntasso/kratix-cli/commit/44ccf97b4a83a856cbd9e5e76e78364604d75c09))
* **init-pulumi:** plan for feature implementation ([8da9d17](https://github.com/syntasso/kratix-cli/commit/8da9d1774f3786ae7ad4cf3e3d1afb5675fbe802))
* **init-pulumi:** refactoring plan after scaffolding ([f618df5](https://github.com/syntasso/kratix-cli/commit/f618df522a70bd0709bce826c8d69617c4513b92))
* **main:** release pulumi-promise 0.1.0 ([87e0b5f](https://github.com/syntasso/kratix-cli/commit/87e0b5f8e06d1e3698735ef1fca923db80f247e0))
* **main:** release pulumi-promise 0.1.0 ([901bce0](https://github.com/syntasso/kratix-cli/commit/901bce03e8cafc511d4ad028a453c75fa2f65ece))
* **pulumi-component-to-cli:** migrate manual-test workflow to regression harness ([02b2588](https://github.com/syntasso/kratix-cli/commit/02b2588f5fed9483035ae84b5beb62bde2282592))
* **pulumi-translate:** split translator by concern ([841d3a9](https://github.com/syntasso/kratix-cli/commit/841d3a9b3f703ef7ebc8c6fcb03ef816d395b4fa))
* **pulumi:** add tests for translate lib ([ef7da82](https://github.com/syntasso/kratix-cli/commit/ef7da82367badb31179268440353d996d6013ba9))
* **pulumi:** align pulumi help example expectation with current output ([e22ac6c](https://github.com/syntasso/kratix-cli/commit/e22ac6c5624acd42ef250fe03a1611070cd4982a))
* **pulumi:** Apply suggestions from code review ([eb2795a](https://github.com/syntasso/kratix-cli/commit/eb2795afc0e9eab5a2f2f2c5ddef0b9939995856))
* **pulumi:** more e2e testing for url parsing ([cca9665](https://github.com/syntasso/kratix-cli/commit/cca9665fcd27befb730fa299641810342988fe1d))
* **pulumi:** remove completed tasks and documentation ([0412cb3](https://github.com/syntasso/kratix-cli/commit/0412cb33aa281a18ed788869236a7f0ff3893587))
* **pulumi:** remove initial conversion code after building into cli ([a0c5370](https://github.com/syntasso/kratix-cli/commit/a0c53704798d2448e0c039e128503b12b2fce25f))
* **pulumi:** remove unnecessary details in readme ([9bfc39e](https://github.com/syntasso/kratix-cli/commit/9bfc39e4e93255c29da41981aa231f96d58bdd50))
* **pulumi:** update command help text for better example ([7a475a8](https://github.com/syntasso/kratix-cli/commit/7a475a8b5838321d95c454e6ab17d4ddd35dbdb5))
* refactorings after pulumi load/select schema ([725611d](https://github.com/syntasso/kratix-cli/commit/725611da4dbe89771dc61d63322c6cbdb65b43b0))
* reorg all pulumi files to correct directories ([a645dbe](https://github.com/syntasso/kratix-cli/commit/a645dbee7b9bbf5d404f13bd20d95e0402d290ed))


### Build System

* **pulumi-component-to-cli:** build test artifacts via binary helper script ([76d6dc1](https://github.com/syntasso/kratix-cli/commit/76d6dc1512a9402dcfcf4348b52e8eacce4eb663))
* **pulumi-component-to-cli:** docker build ([78fa31f](https://github.com/syntasso/kratix-cli/commit/78fa31f7435de4158df30b059d666b07d5bd86b8))

## [0.13.0](https://github.com/syntasso/kratix-cli/compare/v0.12.1...v0.13.0) (2026-01-26)


### Features

* terraform init promise includes providers block ([ee49fe2](https://github.com/syntasso/kratix-cli/commit/ee49fe298c7ac101a414d71222f2087d51dc83c1))


### Chores

* add version alignment check back ([85231af](https://github.com/syntasso/kratix-cli/commit/85231afc339efc5b27706ca91f76312470c50d06))
* bump cli version to v0.13.0 ([20cb98c](https://github.com/syntasso/kratix-cli/commit/20cb98c52cfaae0c3167464168f39fc3ce1f1103))
* dont include chores in release notes, use simpler regex for releasing ([93a90b3](https://github.com/syntasso/kratix-cli/commit/93a90b3444770034a30f913af02c405aa440872a))
* **main:** release terraform-module-promise 0.5.0 ([#203](https://github.com/syntasso/kratix-cli/issues/203)) ([9c66723](https://github.com/syntasso/kratix-cli/commit/9c66723a0da930e898e037f8e480c2421abfbd71))

## [0.12.1](https://github.com/syntasso/kratix-cli/compare/v0.12.0...v0.12.1) (2026-01-05)


### Bug Fixes

* document correct examples in init tf-module-promise command ([#198](https://github.com/syntasso/kratix-cli/issues/198)) ([e83d67d](https://github.com/syntasso/kratix-cli/commit/e83d67dccca74cab915e8e9f158e9fd7e8f7b3f1))


## [0.12.0](https://github.com/syntasso/kratix-cli/compare/v0.11.2...v0.12.0) (2026-01-05)


### Bug Fixes

* Ensure Version field is handled correctly when bootstrapping TF module Promise ([#191](https://github.com/syntasso/kratix-cli/issues/191)) ([cbcfa51](https://github.com/syntasso/kratix-cli/commit/cbcfa514a18c1d15abee2813164ceb58676e97ad))


### Chores

* **main:** release terraform-module-promise 0.4.0 ([#192](https://github.com/syntasso/kratix-cli/issues/192)) ([0b894dd](https://github.com/syntasso/kratix-cli/commit/0b894ddb1e1d4794b5ebcdd23f25971932ba2d17))

## [0.11.2](https://github.com/syntasso/kratix-cli/compare/v0.11.1...v0.11.2) (2025-12-15)


### Chores

* **main:** release crossplane-promise 0.2.1 ([0027919](https://github.com/syntasso/kratix-cli/commit/00279196cf2cc918452bf52b602cd6fbad79fc56))
* **main:** release crossplane-promise 0.2.1 ([d341dcf](https://github.com/syntasso/kratix-cli/commit/d341dcf098ebc7aa35cb176826959e9383009ab5))
* **main:** release helm-promise 0.3.1 ([dc8a152](https://github.com/syntasso/kratix-cli/commit/dc8a152a2eeedad4b622d8f14e1b3a64ce17027f))
* **main:** release helm-promise 0.3.1 ([8235ffa](https://github.com/syntasso/kratix-cli/commit/8235ffa93146c01b07c5b73e4690ea60efbdbe05))

## [0.11.1](https://github.com/syntasso/kratix-cli/compare/v0.11.0...v0.11.1) (2025-12-08)


### Chores

* 159/release stage images ([#173](https://github.com/syntasso/kratix-cli/issues/173)) ([e39f21b](https://github.com/syntasso/kratix-cli/commit/e39f21ba13c918e7cbdb8ee0095ba90f8aab6de2))
* add helm-promise image to release-please config and manaifest ([15bc290](https://github.com/syntasso/kratix-cli/commit/15bc290c63afed47815caec8f8cb1ff311723b26))
* add makefile for testing and releasing helm-promise ([1180a3d](https://github.com/syntasso/kratix-cli/commit/1180a3d0d1d354064fc622d89215ea67cc85206d))
* add release type for helm-promise release ([4c6aa1a](https://github.com/syntasso/kratix-cli/commit/4c6aa1a5e982e422b401a71c1ca665ac98349498))
* add release-please config for crossplane-promise ([d1edbf1](https://github.com/syntasso/kratix-cli/commit/d1edbf10be6d1ebc0be354b9f3d52213de0a3693))
* add release-please configuration for terraform module promise ([8b0fa81](https://github.com/syntasso/kratix-cli/commit/8b0fa8182c43d8d998dc42949e1ee2d8561c15ff))
* added dry-run to jobs ([250ac28](https://github.com/syntasso/kratix-cli/commit/250ac289c33a09ea72fda7bd1e8b053ceada246c))
* added help to location makefiles ([b210cb9](https://github.com/syntasso/kratix-cli/commit/b210cb990bf6309a933af4aec6b051f29647c1a2))
* calling targets from stage makefiles ([977a3c2](https://github.com/syntasso/kratix-cli/commit/977a3c2dc6c257c39a97e1778aadbd2ea49e0291))
* cleaned changelog ([3a31fcd](https://github.com/syntasso/kratix-cli/commit/3a31fcd4489d4f0d1fca385a913a6e4587ee87a7))
* configure helm-promise to release-as 0.3.0 ([9138fdc](https://github.com/syntasso/kratix-cli/commit/9138fdc76af101b2a86ec88df144b835626bf509))
* confirm govulncheck installation ([d660b72](https://github.com/syntasso/kratix-cli/commit/d660b72732960f82acbe5abfcf5cfa6aca79abb3))
* correct path to stages directory in release-image action ([bbcc007](https://github.com/syntasso/kratix-cli/commit/bbcc007a025d16367e16ee7ca73ed1aaf4dbaca7))
* correct path to stages directory in release-image action ([5ae57ec](https://github.com/syntasso/kratix-cli/commit/5ae57ecc9ffc281129619b015295ba4c77442518))
* ensure a component name is set in helm-promise release config ([fa28504](https://github.com/syntasso/kratix-cli/commit/fa2850405efe7219bbb9c45e2ed73779b75a9703))
* ensure buildx builder is created before building images ([1716305](https://github.com/syntasso/kratix-cli/commit/1716305087fca8bf19e4af88da80d20f452c5234))
* ensure release-image action prints the package and version in outputs ([0fae459](https://github.com/syntasso/kratix-cli/commit/0fae4599ac23a31dd49b2939ea76e84030d138d1))
* ensuring that image building works in each subdirectory ([9794f76](https://github.com/syntasso/kratix-cli/commit/9794f76e3c8c4be0cabe06fcb158d0dae431637b))
* fix variable parsing in release-image action ([ded5a4a](https://github.com/syntasso/kratix-cli/commit/ded5a4a0af6f60d7ce8cfc98eda779a26e7664b8))
* fixed wrong variable names ([d85a8c7](https://github.com/syntasso/kratix-cli/commit/d85a8c73a8925aa4be40110868a4ef8e7866ddaa))
* fixing jobs to run from Makefiles on stages locations ([92f3360](https://github.com/syntasso/kratix-cli/commit/92f3360416bf292641c060a11523170084af07c2))
* **main:** release crossplane-promise 0.2.0 ([27b82a1](https://github.com/syntasso/kratix-cli/commit/27b82a15cdc42e201cc200db67b68e6efc90a008))
* **main:** release crossplane-promise 0.2.0 ([6b44225](https://github.com/syntasso/kratix-cli/commit/6b44225dbbffa02caded1f0f9bff456d82274595))
* **main:** release crossplane-promise 0.2.0 ([3234df9](https://github.com/syntasso/kratix-cli/commit/3234df97697ce3a4eed2456ce8bee9bca1861b60))
* **main:** release crossplane-promise 0.2.0 ([8adef0a](https://github.com/syntasso/kratix-cli/commit/8adef0aa224804cdb3dd0f8cd339ae837397410c))
* **main:** release crossplane-promise 0.2.0 ([c1fe883](https://github.com/syntasso/kratix-cli/commit/c1fe88377acbdfae8d8d152a5da6e3cec2162e16))
* **main:** release crossplane-promise 0.2.0 ([6062dd9](https://github.com/syntasso/kratix-cli/commit/6062dd9ebf42dce0197c9b86f9519ea3d96f60d2))
* **main:** release crossplane-promise 0.2.0 ([4897205](https://github.com/syntasso/kratix-cli/commit/489720596ba810d10a31abdebb10148bcacc50fb))
* **main:** release crossplane-promise 0.2.0 ([43de305](https://github.com/syntasso/kratix-cli/commit/43de30563015d082dde23dc087d6e962c7686579))
* **main:** release crossplane-promise 0.2.0 ([#176](https://github.com/syntasso/kratix-cli/issues/176)) ([42117c7](https://github.com/syntasso/kratix-cli/commit/42117c798e6e0027c47095c62fd8e8308c5741dc))
* **main:** release helm-promise 0.3.0 ([#172](https://github.com/syntasso/kratix-cli/issues/172)) ([2cf409d](https://github.com/syntasso/kratix-cli/commit/2cf409d856516633ed79bf76aae034efbb1b41b6))
* **main:** release operator-promise 0.3.0 ([fe95627](https://github.com/syntasso/kratix-cli/commit/fe9562708c92de42999b4b4617212c5cf749f506))
* **main:** release operator-promise 0.3.0 ([90e7a66](https://github.com/syntasso/kratix-cli/commit/90e7a66ed474587b1f730ec718966b63e5fbc9fe))
* release-please configuration for operator-promise ([a796eb9](https://github.com/syntasso/kratix-cli/commit/a796eb9a09c14ea9eb0abd276cbea4ddbda6f8c4))
* remove unnecessary test annotation ([4a7f825](https://github.com/syntasso/kratix-cli/commit/4a7f825467159a534ae08dc7d03a4149c0d92b97))
* removed version pinning ([bc9c93e](https://github.com/syntasso/kratix-cli/commit/bc9c93e5f4240c9aba8ce1efc5fe6bfe39077b8d))
* removed version pinning for operator-promise release ([2716131](https://github.com/syntasso/kratix-cli/commit/271613152bda8916129b9f6ce10c7a8f9cd7c751))
* reverting release-please configuration ([07c694b](https://github.com/syntasso/kratix-cli/commit/07c694be9334630a6b871a5a104d3046385bd2da))
* set initial version of helm-promise to 0.3.0 ([49bbdb6](https://github.com/syntasso/kratix-cli/commit/49bbdb659d51d5e961090699597a2b18227b3766))
* set the current version of the helm-promise to 0.0.0 ([f8da4eb](https://github.com/syntasso/kratix-cli/commit/f8da4eb77f19d2e1d9467b63c7bf8aea94f31313))
* speed up cross-arch compilation ([0882954](https://github.com/syntasso/kratix-cli/commit/0882954bd30ade5cac329c5adbf23a667290db04))
* unpin helm-promise from v0.3.0 ([16a5e2d](https://github.com/syntasso/kratix-cli/commit/16a5e2d4dd7870f76c1c06583d935f3cce2d917b))
* update release please config to uopen separate PRs for each component ([285be29](https://github.com/syntasso/kratix-cli/commit/285be29430bb77a8d5a5463aa4dc694280bd2d28))
* validate files scanned by govulncheck ([560ff11](https://github.com/syntasso/kratix-cli/commit/560ff11faae77e37983bff5745a30551b8385646))


### Build System

* **deps:** bump golang.org/x/crypto from 0.41.0 to 0.45.0 ([#177](https://github.com/syntasso/kratix-cli/issues/177)) ([c36e577](https://github.com/syntasso/kratix-cli/commit/c36e57748d56a1b547fe20a66b032110a9b2b6ad))

## [0.11.0](https://github.com/syntasso/kratix-cli/compare/v0.10.0...v0.11.0) (2025-12-03)


### Features

* refactor Terraform module resolution via terraform init to support terraform private registry modules ([#167](https://github.com/syntasso/kratix-cli/issues/167)) ([4fce203](https://github.com/syntasso/kratix-cli/commit/4fce203c44c25deb59f3a98ac0cff1906262bbc0))
* use isPromiseWorkflow/is_promise_workflow functions within generated workflow templates ([7c99c4a](https://github.com/syntasso/kratix-cli/commit/7c99c4a28d4529b5696600fd6e22c7ff098919e4))
* use isPromiseWorkflow/is_promise_workflow functions within generated workflow templates ([0e3a1a3](https://github.com/syntasso/kratix-cli/commit/0e3a1a3f8e134979726c8e1cccd4f6fcad1f1d1a))


### Bug Fixes

* **#154:** ensure tf defaults are parsed correctly ([#156](https://github.com/syntasso/kratix-cli/issues/156)) ([d51804a](https://github.com/syntasso/kratix-cli/commit/d51804a923aa2470f288a96652cff440aa3060e4))
* copilot codereview ([5f65fc0](https://github.com/syntasso/kratix-cli/commit/5f65fc068dec29ea63d054b3c1b38819be0ce79e))
* defaults for list(map(string)) ([0963e27](https://github.com/syntasso/kratix-cli/commit/0963e277efea43ccf14e0be6f34ffe4492537999))
* do not assume terraform module source protocol ([#157](https://github.com/syntasso/kratix-cli/issues/157)) ([12ef49d](https://github.com/syntasso/kratix-cli/commit/12ef49d42662a7fd992197fd4762013bcc9c174e))
* proper parsing of defaults for maps ([41a770c](https://github.com/syntasso/kratix-cli/commit/41a770c6d438b7f51a1987488aaaf7cfccb92874))
* sets a default value to .spec ([0676981](https://github.com/syntasso/kratix-cli/commit/067698145713a14afef022e4c7f26b9025a586e8)), closes [#155](https://github.com/syntasso/kratix-cli/issues/155)


### Chores

* add a space in help msg ([18f4974](https://github.com/syntasso/kratix-cli/commit/18f4974ea8f4d5285275ad7d9248c6278db65e57))
* bootstrap releases for path: stages/helm-promise ([#168](https://github.com/syntasso/kratix-cli/issues/168)) ([fcaaf37](https://github.com/syntasso/kratix-cli/commit/fcaaf37b96dace81a1fde3e21b05664cfe3e2356))
* install python SDK from pypi instead of github ([35d2289](https://github.com/syntasso/kratix-cli/commit/35d2289cf4fb740e16f3587f69ef72a945425477))
* install python SDK from pypi instead of github ([0b7af0b](https://github.com/syntasso/kratix-cli/commit/0b7af0b6b7e4199ad0453274f001b9b6318d68a2))
* temporary fix for release a terraform generate 0.3.0 ([#164](https://github.com/syntasso/kratix-cli/issues/164)) ([eba557e](https://github.com/syntasso/kratix-cli/commit/eba557e16b5d590f3ebe22bfa2c8c570f5ffcff8))
* update version to match next release version ([1b72ede](https://github.com/syntasso/kratix-cli/commit/1b72ede0e8695c5aad2516caf8da225d0e2ea77f))

## [0.10.0](https://github.com/syntasso/kratix-cli/compare/v0.9.2...v0.10.0) (2025-10-30)


### Features

* containerrun helper can take in envvars ([fc71234](https://github.com/syntasso/kratix-cli/commit/fc7123471c57a7beb0d38d6a8c04b455d2fd2d39))


### Chores

* bump dependencies ([#152](https://github.com/syntasso/kratix-cli/issues/152)) ([f8a92a3](https://github.com/syntasso/kratix-cli/commit/f8a92a33cf8bd5fe112b6ab4f9b89922a4d70ef2))
* ensure LoadPromiseWithAPI reads the promise.yaml when it exists ([0fdea55](https://github.com/syntasso/kratix-cli/commit/0fdea55f09c067061a158dd98c2dec5459268f7c))
* introduce LoadPromiseWithAPI for parsing promises.yaml and api.yaml ([d368f99](https://github.com/syntasso/kratix-cli/commit/d368f993f94ad97ba69cf8ad915333789e33651f))
* provide env var and --env flag as two separate args the run command in ForkRunCommand ([cebf1be](https://github.com/syntasso/kratix-cli/commit/cebf1bedcb1b81d51dbbb2f661f8068b886d662d))

## [0.9.2](https://github.com/syntasso/kratix-cli/compare/v0.9.1...v0.9.2) (2025-10-09)


### Chores

* bump cli version ([537e089](https://github.com/syntasso/kratix-cli/commit/537e0899c832ed9d1209a79b0de627bd123ee9b4))
* remove debug output ([52859bc](https://github.com/syntasso/kratix-cli/commit/52859bca3acf9d0ed21227dad3e5e16149a6d55f))

## [0.9.1](https://github.com/syntasso/kratix-cli/compare/v0.9.0...v0.9.1) (2025-10-08)


### Chores

* document build container in README ([b15ae46](https://github.com/syntasso/kratix-cli/commit/b15ae4654b9c8f849ad1aa78c2b88e0e5d402712))

## [0.9.0](https://github.com/syntasso/kratix-cli/compare/v0.8.0...v0.9.0) (2025-10-08)


### Features

* **#119:** introduce 'platform get resources' command to detail resource request ([6eaeeff](https://github.com/syntasso/kratix-cli/commit/6eaeeff11f22069d1cbf6f9d7d657a4e0cb3e1c1))
* kratix cli supports plugins ([#148](https://github.com/syntasso/kratix-cli/issues/148)) ([4106791](https://github.com/syntasso/kratix-cli/commit/4106791af19fb332d6d056e39ed2adce6b75f189))


### Chores

* bumping main.go version ([4522421](https://github.com/syntasso/kratix-cli/commit/452242109345146dff5ca419f4eac45f8a9f7db3))
* bumping main.go version to 0.9.0 ([40d1529](https://github.com/syntasso/kratix-cli/commit/40d15292093e07eabb1619436d5964fb4eb3fe54))
* fix operator-promise dockerfile ([cbec6cf](https://github.com/syntasso/kratix-cli/commit/cbec6cf6896c2591d370d51428040af5ca2cda82))
* Introduce utils packages ([#147](https://github.com/syntasso/kratix-cli/issues/147)) ([e1cebb8](https://github.com/syntasso/kratix-cli/commit/e1cebb8431ef702fee9838dfb8c6a87eb048442d))
* **main:** release 0.9.0 ([11e2642](https://github.com/syntasso/kratix-cli/commit/11e2642fd0b80fac65bf9e9f1bfe8e9a81fd0a2d))
* **main:** release 0.9.0 ([6ac6f2f](https://github.com/syntasso/kratix-cli/commit/6ac6f2fe24a3f1af8e088ee48b17f93ad3cbdec5))
* **main:** release 0.9.0 ([6313167](https://github.com/syntasso/kratix-cli/commit/63131676c80976a8de6712a36de3bfee616f166f))
* **main:** release 0.9.0 ([17a0ca4](https://github.com/syntasso/kratix-cli/commit/17a0ca4445a24c048b8ccb3ac60d15ba8f83a8f3))

## [0.8.0](https://github.com/syntasso/kratix-cli/compare/v0.7.0...v0.8.0) (2025-09-16)


### Features

* Add support for modules in mono repos ([#130](https://github.com/syntasso/kratix-cli/issues/130)) ([0f10579](https://github.com/syntasso/kratix-cli/commit/0f10579eab2a22cc071f512c1d6136e91efb3571))


### Chores

* remove references to bitnami ([#131](https://github.com/syntasso/kratix-cli/issues/131)) ([80488db](https://github.com/syntasso/kratix-cli/commit/80488db934c379b5b76d81c6b49d9694b9c9e0f7))


### Build System

* **deps:** bump github.com/hashicorp/go-getter from 1.7.8 to 1.7.9 ([#129](https://github.com/syntasso/kratix-cli/issues/129)) ([a36040b](https://github.com/syntasso/kratix-cli/commit/a36040b03c7d58b88732ed451fe9c40d698465df))

## [0.7.0](https://github.com/syntasso/kratix-cli/compare/v0.6.1...v0.7.0) (2025-08-15)


### Features

* introduce language flag for 'add container' command [#121](https://github.com/syntasso/kratix-cli/issues/121) ([3d9c9c2](https://github.com/syntasso/kratix-cli/commit/3d9c9c2ca85d576ea5762edf3dfa5cde21f567f3))


### Bug Fixes

* add container --python templates ([#124](https://github.com/syntasso/kratix-cli/issues/124)) ([1f30310](https://github.com/syntasso/kratix-cli/commit/1f303108682a7563dfc76a2fe42532687c21c4f2))


### Chores

* bump hardcoded cli version ([c738e90](https://github.com/syntasso/kratix-cli/commit/c738e90d0719aa66768b9daaa7fe0873f9c915d3))
* use the RELASE_CREATOR_TOKEN when creating release PRs ([3b25ff0](https://github.com/syntasso/kratix-cli/commit/3b25ff06816b3a9d055c32ff5e1825c1829105eb))
* use the RELASE_CREATOR_TOKEN when creating release PRs ([62a2f80](https://github.com/syntasso/kratix-cli/commit/62a2f803bbd0b238ab192747e60026a94957703d))


### Build System

* **deps:** bump golang.org/x/oauth2 from 0.24.0 to 0.27.0 ([a0320e2](https://github.com/syntasso/kratix-cli/commit/a0320e26e2dc179bd90b7bcc5fcdeb348178a74b))
* **deps:** bump golang.org/x/oauth2 from 0.24.0 to 0.27.0 ([959c175](https://github.com/syntasso/kratix-cli/commit/959c1753caf43182375b7e764416064dce4eb341))
* **deps:** bump helm.sh/helm/v3 from 3.17.3 to 3.17.4 ([e4347b8](https://github.com/syntasso/kratix-cli/commit/e4347b8cfccdae284932854feff52f607d140e96))
* **deps:** bump helm.sh/helm/v3 from 3.17.3 to 3.17.4 ([e415576](https://github.com/syntasso/kratix-cli/commit/e4155768e05708092de5bee678ba851ad96b3cc7))
* **deps:** bump helm.sh/helm/v3 from 3.17.4 to 3.18.5 ([#125](https://github.com/syntasso/kratix-cli/issues/125)) ([4a21ada](https://github.com/syntasso/kratix-cli/commit/4a21ada20aecb9a3ce8d42676e9ba662d18b0acf))

## [0.6.1](https://github.com/syntasso/kratix-cli/compare/v0.6.0...v0.6.1) (2025-06-23)


### Bug Fixes

* release v0.6.1 ([9fef9aa](https://github.com/syntasso/kratix-cli/commit/9fef9aa177fad0a8a38ca47c4bd33f6fda8b05fd))

## [0.6.0](https://github.com/syntasso/kratix-cli/compare/v0.5.4...v0.6.0) (2025-06-23)


### Features

* Add defaults to primative tf variables ([#113](https://github.com/syntasso/kratix-cli/issues/113)) ([96976ca](https://github.com/syntasso/kratix-cli/commit/96976ca33288dab69e5373dae3ce322812eee111))


### Chores

* aspect -&gt; stage ([#106](https://github.com/syntasso/kratix-cli/issues/106)) ([46a2dd7](https://github.com/syntasso/kratix-cli/commit/46a2dd782c53ae27fe4a37eebd0ffec45a7f568a))
* checkout at a tag for goreleaser [skip ci] ([623812c](https://github.com/syntasso/kratix-cli/commit/623812ceac73d712ef81471dc3941c8d489da658))
* fix bash in release.yaml [skip ci] ([8852458](https://github.com/syntasso/kratix-cli/commit/885245834249995ab630e248c91d0a8c0aa88289))
* make it possible to trigger release manually ([d339bb8](https://github.com/syntasso/kratix-cli/commit/d339bb82814bcf5b741e9faff16472d3432d469e))
* release fixes to cut 0.5.4 correctly ([573b575](https://github.com/syntasso/kratix-cli/commit/573b5759565817bb11e756f401e3cb7305dfede3))
* upgrade kratix dependency ([#108](https://github.com/syntasso/kratix-cli/issues/108)) ([98b3fe6](https://github.com/syntasso/kratix-cli/commit/98b3fe6e308a495b4b0473c0d6a6f9df16c22374))

## [0.5.4](https://github.com/syntasso/kratix-cli/compare/v0.5.3...v0.5.4) (2025-05-13)


### Bug Fixes

* include all dependency files from a directory when using --image ([3387b88](https://github.com/syntasso/kratix-cli/commit/3387b88a0e2c2d1673781edd1f3c9fcc2bab3a9c))
* include all dependency files from a directory when using --image ([cee117a](https://github.com/syntasso/kratix-cli/commit/cee117a654cef5c46db37bb7157357a515919bd9))

## [0.5.3](https://github.com/syntasso/kratix-cli/compare/v0.5.2...v0.5.3) (2025-04-10)


### Bug Fixes

* bump go version to 1.23.0 ([40d6dfa](https://github.com/syntasso/kratix-cli/commit/40d6dfa6b5e284a2ff626754d12c57b1af56cc03))
* revert v0.5.3 release in release-please-config.json ([65c00f1](https://github.com/syntasso/kratix-cli/commit/65c00f12d875861d49cdd719c43cfad29011d70f))
* support empty openAPIV3Schema in xrd for the init-crossplane command ([#88](https://github.com/syntasso/kratix-cli/issues/88)) ([c4738ca](https://github.com/syntasso/kratix-cli/commit/c4738ca6e23b28922539da252f94c8c14dfa8f64))


### Chores

* bump cli version ([b8877d9](https://github.com/syntasso/kratix-cli/commit/b8877d967c73b83c0de012bf28448382cc61dd7e))
* **main:** release 0.5.3 ([#89](https://github.com/syntasso/kratix-cli/issues/89)) ([da1d94d](https://github.com/syntasso/kratix-cli/commit/da1d94d46bab44ab35dd148096a4d723dcf60ae2))


### Build System

* **deps:** bump helm.sh/helm/v3 from 3.15.2 to 3.17.3 ([461d877](https://github.com/syntasso/kratix-cli/commit/461d8774b5864645bddd9660bb3cb1decc3a5995))
* **deps:** bump helm.sh/helm/v3 from 3.15.2 to 3.17.3 ([735f500](https://github.com/syntasso/kratix-cli/commit/735f50016c533fc6bf05c5df25cda2978ba4e7d0))

## [0.5.3](https://github.com/syntasso/kratix-cli/compare/v0.5.2...v0.5.3) (2025-04-10)


### Bug Fixes

* support empty openAPIV3Schema in xrd for the init-crossplane command ([#88](https://github.com/syntasso/kratix-cli/issues/88)) ([c4738ca](https://github.com/syntasso/kratix-cli/commit/c4738ca6e23b28922539da252f94c8c14dfa8f64))


### Chores

* bump cli version ([b8877d9](https://github.com/syntasso/kratix-cli/commit/b8877d967c73b83c0de012bf28448382cc61dd7e))


### Build System

* **deps:** bump helm.sh/helm/v3 from 3.15.2 to 3.17.3 ([461d877](https://github.com/syntasso/kratix-cli/commit/461d8774b5864645bddd9660bb3cb1decc3a5995))
* **deps:** bump helm.sh/helm/v3 from 3.15.2 to 3.17.3 ([735f500](https://github.com/syntasso/kratix-cli/commit/735f50016c533fc6bf05c5df25cda2978ba4e7d0))

## [0.5.2](https://github.com/syntasso/kratix-cli/compare/v0.5.1...v0.5.2) (2025-04-09)


### Bug Fixes

* support empty spec.properties in xrd for the init-crossplane command ([d543cbf](https://github.com/syntasso/kratix-cli/commit/d543cbf17ae20fd8e1f05fdd63f2c2c44856392c))


### Chores

* adding check to ensure released version alignment ([728999d](https://github.com/syntasso/kratix-cli/commit/728999ded544f9099b8be9ba2be67f63fab981b1))
* bump cli version ([887685e](https://github.com/syntasso/kratix-cli/commit/887685e0a6be6a922e984115c53f90336dac2b9a))
* bump golang orb version ([99d11a4](https://github.com/syntasso/kratix-cli/commit/99d11a48d7415a0730bcbcf1ea5b8bc0ce99ef41))
* ignore release please branches ([947a0fe](https://github.com/syntasso/kratix-cli/commit/947a0fe2b04fe099fe53eb15b1c45094bf0a14a8))
* run check on release-please branches ([671e053](https://github.com/syntasso/kratix-cli/commit/671e053f83d9fd8d9cc340548b5ccdd2baf87c20))

## [0.5.1](https://github.com/syntasso/kratix-cli/compare/v0.5.0...v0.5.1) (2025-03-31)


### Bug Fixes

* dont error when .spec isn't provided in terraform request ([88e030b](https://github.com/syntasso/kratix-cli/commit/88e030b71743c28955de7260c65dacea951f189f))

## [0.5.0](https://github.com/syntasso/kratix-cli/compare/v0.4.0...v0.5.0) (2025-03-19)


### Features

* default to versioning all initialised promises ([49113fb](https://github.com/syntasso/kratix-cli/commit/49113fb4d2d465f1645e94be0c0dc03a9eb37283))
* Init Promise from Crossplane Composition ([#78](https://github.com/syntasso/kratix-cli/issues/78)) ([5223d31](https://github.com/syntasso/kratix-cli/commit/5223d317cfe09008659bb4399e697ece65b226e4))
* Init Promise from Terraform Module ([#77](https://github.com/syntasso/kratix-cli/issues/77)) ([76dac2e](https://github.com/syntasso/kratix-cli/commit/76dac2eb0786cf086925991de26315b4378a7612))


### Bug Fixes

* remove duplicate 'type: object' from initialised Promise API ([3b6d97b](https://github.com/syntasso/kratix-cli/commit/3b6d97b9a1c38e4721de7a0866c0e96169890ebf))


### Chores

* include chores in changelog ([d3fe2cc](https://github.com/syntasso/kratix-cli/commit/d3fe2cc2b2f1947a86ff460ae3a9825530cc0fec))
* set version in code ([1de09d2](https://github.com/syntasso/kratix-cli/commit/1de09d2dc38f0c61f7989d36f62e24b40084917c))

## [0.4.0](https://github.com/syntasso/kratix-cli/compare/v0.3.0...v0.4.0) (2025-01-27)


### Features

* add support for boolean types ([63244bf](https://github.com/syntasso/kratix-cli/commit/63244bf40109c6070e99af418d0afc2c017e5bea))


### Bug Fixes

* improve help of update api subcommand ([42d6c46](https://github.com/syntasso/kratix-cli/commit/42d6c46d4da441c83801000b1e004e162049f658))
* include path to main.go in goreleaseer ([#73](https://github.com/syntasso/kratix-cli/issues/73)) ([600bd11](https://github.com/syntasso/kratix-cli/commit/600bd114c25301b3be603b3833b3cfa60050e8f3))

## [0.3.0](https://github.com/syntasso/kratix-cli/compare/v0.2.1...v0.3.0) (2024-12-19)


### Features

* `go install` will now install a `kratix` binary ([#66](https://github.com/syntasso/kratix-cli/issues/66)) ([5a84127](https://github.com/syntasso/kratix-cli/commit/5a841271180ef6dda336579b0d0d78319f354d7b)), closes [#65](https://github.com/syntasso/kratix-cli/issues/65)
* set label&annos from request in operator promise ([#64](https://github.com/syntasso/kratix-cli/issues/64)) ([c9af9c2](https://github.com/syntasso/kratix-cli/commit/c9af9c20a66002aaf3a3a2a6b2b73c26d81094bc))


### Bug Fixes

* create valid workflows for plain promise init ([3e979b6](https://github.com/syntasso/kratix-cli/commit/3e979b6d6fe0b994eacf82bebadc4c3285d24bf0))
* create valid workflows for plain promise init ([9d57553](https://github.com/syntasso/kratix-cli/commit/9d575537dde75697745b64d6f4a21fb15dc574b0))

## [0.2.1](https://github.com/syntasso/kratix-cli/compare/v0.2.0...v0.2.1) (2024-09-20)


### Bug Fixes

* build a statically linked binary ([2dbcc96](https://github.com/syntasso/kratix-cli/commit/2dbcc9635f7a2e874973aa3083f1f2d3720a5622))

## [0.2.0](https://github.com/syntasso/kratix-cli/compare/v0.1.0...v0.2.0) (2024-09-13)


### Features

* introduce 'build container' command [#36](https://github.com/syntasso/kratix-cli/issues/36) ([5882bb1](https://github.com/syntasso/kratix-cli/commit/5882bb1787c374ad51d3746188746791f900afb7))
* refactor loadworkflows ([2a7c4bb](https://github.com/syntasso/kratix-cli/commit/2a7c4bb6ce37eebdf23c46083629416175f2b552))


### Bug Fixes

* build helm image multi arch ([410e279](https://github.com/syntasso/kratix-cli/commit/410e279c0e26b7db6d9b8a2ea9d64dad9b99e91d))
* bump the resource class to medium when running tests ([2c976c8](https://github.com/syntasso/kratix-cli/commit/2c976c83e160479304ee5fc2c5358156baf408bb))
* construct the correct docker push command ([8e6467e](https://github.com/syntasso/kratix-cli/commit/8e6467e98b34c2e30da67c7a9660748d50447de0))
* ensure 'build promise' does not fail when there are no workflows and when logging that no promise.yaml has been found ([0dd8a90](https://github.com/syntasso/kratix-cli/commit/0dd8a907d951c2da6bf70f3ff6ef451cce894544))
* error gracefully when no workflows exist when building a container ([d0a2810](https://github.com/syntasso/kratix-cli/commit/d0a28104245eed60b5dc3afaac75a546b56d0ec9))
* error if the --name flag specifies a non-existent container name ([63aaefa](https://github.com/syntasso/kratix-cli/commit/63aaefafc20b3270bc6673ecf7b58a8d38f28bd9))
* handle no promise or workflow file existing ([725c8c2](https://github.com/syntasso/kratix-cli/commit/725c8c29c40a0a63596b1fd80fb90842d342d13e))
* increase no_output_timeout for helm operator tests ([29e4dac](https://github.com/syntasso/kratix-cli/commit/29e4dac408c5b413459d878b7c7a3d4ce9737d3c))
* make 'build container' command description more descriptive ([e6f5b99](https://github.com/syntasso/kratix-cli/commit/e6f5b994c125e9f374e4efe07476463b8acb409c))
* refactor parsing args logic ([e7ebf52](https://github.com/syntasso/kratix-cli/commit/e7ebf5233aa881e7046f549a92d93e6f46e37636))
* update buildx example with build args ([2c69014](https://github.com/syntasso/kratix-cli/commit/2c6901448f81ba3faa41aec95b0e451433cd463e))

## [0.1.0](https://github.com/syntasso/kratix-cli/compare/v0.1.0...v0.1.0) (2024-07-30)


### chore

* release 0.1.0 ([d2fbcb9](https://github.com/syntasso/kratix-cli/commit/d2fbcb9327b96919b33de37439debaa940cdd510))


### Features

* 'add container' autogenerates a Dockerfile and empty resources directory ([aadc956](https://github.com/syntasso/kratix-cli/commit/aadc956b35c2cdc7e93bca54c99a0a488b9a913a))
* ([#31](https://github.com/syntasso/kratix-cli/issues/31)) better help message for init commands ([f0efdf6](https://github.com/syntasso/kratix-cli/commit/f0efdf61044b93a32756ddb28b20c8815504e889))
* ([#4](https://github.com/syntasso/kratix-cli/issues/4)) implement --split flag for kratix init ([8de695c](https://github.com/syntasso/kratix-cli/commit/8de695cbf0b2592c7bc80ce6cef334e065834a7a))
* ([#5](https://github.com/syntasso/kratix-cli/issues/5)) add kratix build promise ([cc96f41](https://github.com/syntasso/kratix-cli/commit/cc96f417d4ff8d94e14dcdf28527c238e1cdc4c0))
* ([#6](https://github.com/syntasso/kratix-cli/issues/6)) kratix update api remove properties ([686d166](https://github.com/syntasso/kratix-cli/commit/686d1663cb9b94ee805af37a5d38f73bff41060f))
* ([#8](https://github.com/syntasso/kratix-cli/issues/8)) add kratix update dependencies to readme ([137577a](https://github.com/syntasso/kratix-cli/commit/137577a126af1416f4310648a9e08c14de877cc8))
* ([#8](https://github.com/syntasso/kratix-cli/issues/8)) implement kratix update dependencies ([7077a08](https://github.com/syntasso/kratix-cli/commit/7077a08255f77e1b3a40362933b0ee0a2abcd397))
* ([#8](https://github.com/syntasso/kratix-cli/issues/8)) kratix build will include dependencies in Promise ([a5441c7](https://github.com/syntasso/kratix-cli/commit/a5441c737d61d11d7b38a6bd349b0f9502f66d5c))
* [#12](https://github.com/syntasso/kratix-cli/issues/12) support integer as valid property type ([aefa154](https://github.com/syntasso/kratix-cli/commit/aefa1548cc79df58b8648d95d99b9e814c79d6eb))
* [#12](https://github.com/syntasso/kratix-cli/issues/12) update api works with split files promise ([53f7b76](https://github.com/syntasso/kratix-cli/commit/53f7b76d8941fd4b233534fb5338f2dfe6051dea))
* [#12](https://github.com/syntasso/kratix-cli/issues/12) update example resource when gvk changes ([fcffaf4](https://github.com/syntasso/kratix-cli/commit/fcffaf486f521de8333cd06d6dad0ebc245c42f7))
* [#9](https://github.com/syntasso/kratix-cli/issues/9) build and push helm-resource-configure image ([3cfaa24](https://github.com/syntasso/kratix-cli/commit/3cfaa244924d32fdf6daaad109baab37be36e406))
* [#9](https://github.com/syntasso/kratix-cli/issues/9) generate api schema from chart ([b3f698d](https://github.com/syntasso/kratix-cli/commit/b3f698d77723e9a51c434d541811d5785e00516a))
* [#9](https://github.com/syntasso/kratix-cli/issues/9) helper convert helm values to crd schema ([933c09e](https://github.com/syntasso/kratix-cli/commit/933c09e79fbf8f8419b7751a80baa92acea6e42b))
* [#9](https://github.com/syntasso/kratix-cli/issues/9) template helm resource configure workflow ([ee2bb7d](https://github.com/syntasso/kratix-cli/commit/ee2bb7d927fcd546a5ed35b63e90f93ad71b807a))
* **#10:** add example resource request ([cdbd940](https://github.com/syntasso/kratix-cli/commit/cdbd940287b7f338f6ef5d329d99bbcb21c00878))
* **#10:** add from-api-to-crd aspect for operator-promise ([9c9996a](https://github.com/syntasso/kratix-cli/commit/9c9996aa878f0d0844ea7d7bd97b1379e19fc0e0))
* **#10:** add readme to operator-promise ([7a2724f](https://github.com/syntasso/kratix-cli/commit/7a2724f3193dff4f24cdf9afa402d6c47b795ca8))
* **#10:** generate workflows ([b1fab2e](https://github.com/syntasso/kratix-cli/commit/b1fab2e01c8ab51a680a1070252a512cebb7c35b))
* **#10:** kratix init operator-promise ([94192b5](https://github.com/syntasso/kratix-cli/commit/94192b5ea0dca6d568e9a5359bda750892b3437f))
* **#10:** refactor operator promise in go ([d8eb655](https://github.com/syntasso/kratix-cli/commit/d8eb655ddf5d9f4d2175a4348626bdee69268e87))
* **#10:** support running without --split in operator-promise ([f4ddf11](https://github.com/syntasso/kratix-cli/commit/f4ddf1141bfa0d91dd579cd94c21fff23421cb18))
* **#2:** boostrap cli repo w/ help ([508c2a8](https://github.com/syntasso/kratix-cli/commit/508c2a888815b036074f0a72144d5a3f59b2a50e))
* **#3:** kratix init promise command ([2f2f8d4](https://github.com/syntasso/kratix-cli/commit/2f2f8d4f433967356a522d7ec17c7e764322e7dc))
* add containers works when the init command uses the --split flag ([0a734b0](https://github.com/syntasso/kratix-cli/commit/0a734b0eb397aba9e9180042c18fa4b0842db5ba))
* add support for object types ([#44](https://github.com/syntasso/kratix-cli/issues/44)) ([3edd274](https://github.com/syntasso/kratix-cli/commit/3edd2748762d6daa964c361a76008c1187c23997))
* allow deps to be added as workflows ([#22](https://github.com/syntasso/kratix-cli/issues/22)) ([2ff36f2](https://github.com/syntasso/kratix-cli/commit/2ff36f2f003b1d2689691b56bd3194d91df1fbf6))
* dependencies&operator manifest can take a file ([#26](https://github.com/syntasso/kratix-cli/issues/26)) ([3af2f89](https://github.com/syntasso/kratix-cli/commit/3af2f89145725c86dc2218fb3024509653d3d6ee))
* do not include dependencies in init-operator ([#23](https://github.com/syntasso/kratix-cli/issues/23)) ([50198d7](https://github.com/syntasso/kratix-cli/commit/50198d7e3544544766eeb513be8eb6c4fefaf4b6))
* helm template configure aspect ([3b87321](https://github.com/syntasso/kratix-cli/commit/3b87321b6fec8c00bff9c5259cec9b60a961e02b))
* implement update api ([11930ba](https://github.com/syntasso/kratix-cli/commit/11930ba30602d429a6766d9500209702461317a8))
* improvements to add container cmd ([#20](https://github.com/syntasso/kratix-cli/issues/20)) ([53977fb](https://github.com/syntasso/kratix-cli/commit/53977fbb6361f12e0f7591414316cb14907c10df))
* introduce 'add container' command ([#7](https://github.com/syntasso/kratix-cli/issues/7)) ([729f905](https://github.com/syntasso/kratix-cli/commit/729f905cef2e8c7b6335a0f8b66c4e01de2efeea))
* introduce 'init helm-promise' command ([#9](https://github.com/syntasso/kratix-cli/issues/9)) ([fd6e8d6](https://github.com/syntasso/kratix-cli/commit/fd6e8d655a544bc02cdfd39dedf833a210fca130))
* kratix update destination-selector ([#19](https://github.com/syntasso/kratix-cli/issues/19)) ([cf388d8](https://github.com/syntasso/kratix-cli/commit/cf388d82a23f14ec24f5c05c12d9e3c65fd69f09))
* output an informative message after init promise ([0a6c257](https://github.com/syntasso/kratix-cli/commit/0a6c257773b812256ff6ab8ecff96ec2d0b3b6be))


### Bug Fixes

* 'add containers' does not allow for duplicated container names ([e6aee5e](https://github.com/syntasso/kratix-cli/commit/e6aee5ef08434c494e2f9e27a644cc441844aeaf))
* ([#37](https://github.com/syntasso/kratix-cli/issues/37)) handle null value in helm values files ([bef9200](https://github.com/syntasso/kratix-cli/commit/bef920047fab5691b61f8af446f8ff58c77523a8))
* ([#8](https://github.com/syntasso/kratix-cli/issues/8)) manage missing/empty api file on build ([8b52ad9](https://github.com/syntasso/kratix-cli/commit/8b52ad9dd38b48270a105d8900dc748f5caf4548))
* ([#8](https://github.com/syntasso/kratix-cli/issues/8)) remove --split from update dependencies ([5aad737](https://github.com/syntasso/kratix-cli/commit/5aad737769c50c157d3e45d75b3f8a3cb8858b7e))
* **#28:** ensure workflow.yaml is always an array of Pipelines ([115da10](https://github.com/syntasso/kratix-cli/commit/115da10432cf2c71e3166f80955d7826a057b43b))
* **#29:** better error message for invalid pipeline input ([da940d4](https://github.com/syntasso/kratix-cli/commit/da940d443c21a94a0920feedd2e73b7bfa8bfe53))
* **#2:** set resource_class to small ([b14a0cb](https://github.com/syntasso/kratix-cli/commit/b14a0cb1cc7fc901b5e324f88a736eee4b5c45bc))
* **#3:** test optional flags in `init promise` ([8d51a73](https://github.com/syntasso/kratix-cli/commit/8d51a7365a8839bdfd54d0ad430ea63128656949))
* add --split to persistent flags ([d1d1dbd](https://github.com/syntasso/kratix-cli/commit/d1d1dbd2f5086e71a3f0e805f0764e2f8af23575))
* add release-please-config.json ([8b25c2e](https://github.com/syntasso/kratix-cli/commit/8b25c2e090d9c516087a37345b1b585cc20b28a2))
* better README and cli help usage message ([db2e1b6](https://github.com/syntasso/kratix-cli/commit/db2e1b6c03a1c23f7062b2c4a9e8f3c59684ecec))
* can fetch chart from a specific version ([8f0d8eb](https://github.com/syntasso/kratix-cli/commit/8f0d8ebb9568cb2f435edbb63ff4e3523ac440ad))
* create and cleanup test dirs ([a0dd70e](https://github.com/syntasso/kratix-cli/commit/a0dd70e22353b55342d84202bf0ff8157e3e7885))
* do not indent the pipeline.sh file content ([e19231a](https://github.com/syntasso/kratix-cli/commit/e19231a9636a19470fdc32a6706d025808b38fab))
* handle multiple containers in the same pipeline ([165c075](https://github.com/syntasso/kratix-cli/commit/165c07508b60603c1b7c4262b3060c6153404191))
* more error handling improvements for add container ([6a42202](https://github.com/syntasso/kratix-cli/commit/6a422023954064bef0e577e77a13cdeb0ea7da6e))
* proper treatment of dirs when building deps ([3249fc4](https://github.com/syntasso/kratix-cli/commit/3249fc4e87367dada8bfc3fa261045061ae5a82e))
* remove containers/ from path ([b0bde07](https://github.com/syntasso/kratix-cli/commit/b0bde0732ad78776dd7b764aefd2fe77e63b2076))
* remove extra files from release tarball; remove previous release ([3a12b4f](https://github.com/syntasso/kratix-cli/commit/3a12b4fe980da02be21ecfad2ad55f79eac08d7e))
* remove workflow.yaml from init promise template ([20e4a7a](https://github.com/syntasso/kratix-cli/commit/20e4a7aa487dcfc7c03a365836ba3475327f57fe))
* stop on errors when parsing deps ([0daf7d2](https://github.com/syntasso/kratix-cli/commit/0daf7d2e2993acbd98246958e4b587ac7607de39))
* the 'promise build' command builds the workflows if workflow.yaml files are present ([0a20aa0](https://github.com/syntasso/kratix-cli/commit/0a20aa02c704a1275e8c7ae8b6391a6d235b338f))
* typo in helm promise aspect image name ([142e902](https://github.com/syntasso/kratix-cli/commit/142e902314fb2e2fb77c88fb75ea5025101324f5))
* typos and nil slice declaration ([3796c1c](https://github.com/syntasso/kratix-cli/commit/3796c1c323069091716026a29776428024d5504f))
* update api works with promise.yaml from build ([89c3427](https://github.com/syntasso/kratix-cli/commit/89c342733df63fb8a46727cc065d4749373c348c))
* update generated README from kratix init ([2960a5e](https://github.com/syntasso/kratix-cli/commit/2960a5e4e6ad554c76f950742a541ea389e420e0))
* use default namespace if not provided in operator resources ([d9d5a4b](https://github.com/syntasso/kratix-cli/commit/d9d5a4b0fb71a1439b1f95e4ede12b8f28940f6a))
* use the promise name from the cli ([3dc739d](https://github.com/syntasso/kratix-cli/commit/3dc739d02abcbcafa7d750fbf4a857641d64866b))
* work with container image without project name ([4d52288](https://github.com/syntasso/kratix-cli/commit/4d52288b7794b3f8841f26b63a7b5d4151ff21b2))
* WORKFLOWNAME -&gt; PIPELINENAME ([2e1d3a3](https://github.com/syntasso/kratix-cli/commit/2e1d3a399af7374215cf089e9c210b0eff90d311))
* write api.yaml in yaml format on update api ([e841e5f](https://github.com/syntasso/kratix-cli/commit/e841e5fc9362a50ea3a5d1474fa866ab386b45e5))
