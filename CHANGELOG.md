# Changelog

## [0.1.0](https://github.com/syntasso/kratix-cli/compare/v0.8.0...v0.1.0) (2025-09-16)


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
* **#10:** add image to convert api to operator crd, and support non-split ([c9d2d89](https://github.com/syntasso/kratix-cli/commit/c9d2d8972717faf429847dc33732c40ee178476e))
* **#10:** add readme to operator-promise ([7a2724f](https://github.com/syntasso/kratix-cli/commit/7a2724f3193dff4f24cdf9afa402d6c47b795ca8))
* **#10:** generate workflows ([b1fab2e](https://github.com/syntasso/kratix-cli/commit/b1fab2e01c8ab51a680a1070252a512cebb7c35b))
* **#10:** kratix init operator-promise ([ea0124f](https://github.com/syntasso/kratix-cli/commit/ea0124fa6660477bc45c958faa60813dfdf5fc39))
* **#10:** kratix init operator-promise ([94192b5](https://github.com/syntasso/kratix-cli/commit/94192b5ea0dca6d568e9a5359bda750892b3437f))
* **#10:** refactor operator promise in go ([d8eb655](https://github.com/syntasso/kratix-cli/commit/d8eb655ddf5d9f4d2175a4348626bdee69268e87))
* **#10:** support running without --split in operator-promise ([f4ddf11](https://github.com/syntasso/kratix-cli/commit/f4ddf1141bfa0d91dd579cd94c21fff23421cb18))
* **#2:** boostrap cli repo w/ help ([508c2a8](https://github.com/syntasso/kratix-cli/commit/508c2a888815b036074f0a72144d5a3f59b2a50e))
* **#3:** kratix init promise command ([2f2f8d4](https://github.com/syntasso/kratix-cli/commit/2f2f8d4f433967356a522d7ec17c7e764322e7dc))
* `go install` will now install a `kratix` binary ([#66](https://github.com/syntasso/kratix-cli/issues/66)) ([5a84127](https://github.com/syntasso/kratix-cli/commit/5a841271180ef6dda336579b0d0d78319f354d7b)), closes [#65](https://github.com/syntasso/kratix-cli/issues/65)
* add containers works when the init command uses the --split flag ([0a734b0](https://github.com/syntasso/kratix-cli/commit/0a734b0eb397aba9e9180042c18fa4b0842db5ba))
* Add defaults to primative tf variables ([#113](https://github.com/syntasso/kratix-cli/issues/113)) ([96976ca](https://github.com/syntasso/kratix-cli/commit/96976ca33288dab69e5373dae3ce322812eee111))
* add support for boolean types ([63244bf](https://github.com/syntasso/kratix-cli/commit/63244bf40109c6070e99af418d0afc2c017e5bea))
* Add support for modules in mono repos ([#130](https://github.com/syntasso/kratix-cli/issues/130)) ([0f10579](https://github.com/syntasso/kratix-cli/commit/0f10579eab2a22cc071f512c1d6136e91efb3571))
* add support for object types ([#44](https://github.com/syntasso/kratix-cli/issues/44)) ([3edd274](https://github.com/syntasso/kratix-cli/commit/3edd2748762d6daa964c361a76008c1187c23997))
* allow deps to be added as workflows ([#22](https://github.com/syntasso/kratix-cli/issues/22)) ([2ff36f2](https://github.com/syntasso/kratix-cli/commit/2ff36f2f003b1d2689691b56bd3194d91df1fbf6))
* default to versioning all initialised promises ([49113fb](https://github.com/syntasso/kratix-cli/commit/49113fb4d2d465f1645e94be0c0dc03a9eb37283))
* dependencies&operator manifest can take a file ([#26](https://github.com/syntasso/kratix-cli/issues/26)) ([3af2f89](https://github.com/syntasso/kratix-cli/commit/3af2f89145725c86dc2218fb3024509653d3d6ee))
* do not include dependencies in init-operator ([#23](https://github.com/syntasso/kratix-cli/issues/23)) ([50198d7](https://github.com/syntasso/kratix-cli/commit/50198d7e3544544766eeb513be8eb6c4fefaf4b6))
* helm template configure aspect ([3b87321](https://github.com/syntasso/kratix-cli/commit/3b87321b6fec8c00bff9c5259cec9b60a961e02b))
* implement update api ([11930ba](https://github.com/syntasso/kratix-cli/commit/11930ba30602d429a6766d9500209702461317a8))
* improvements to add container cmd ([#20](https://github.com/syntasso/kratix-cli/issues/20)) ([53977fb](https://github.com/syntasso/kratix-cli/commit/53977fbb6361f12e0f7591414316cb14907c10df))
* Init Promise from Crossplane Composition ([#78](https://github.com/syntasso/kratix-cli/issues/78)) ([5223d31](https://github.com/syntasso/kratix-cli/commit/5223d317cfe09008659bb4399e697ece65b226e4))
* Init Promise from Terraform Module ([#77](https://github.com/syntasso/kratix-cli/issues/77)) ([76dac2e](https://github.com/syntasso/kratix-cli/commit/76dac2eb0786cf086925991de26315b4378a7612))
* introduce 'add container' command ([#7](https://github.com/syntasso/kratix-cli/issues/7)) ([729f905](https://github.com/syntasso/kratix-cli/commit/729f905cef2e8c7b6335a0f8b66c4e01de2efeea))
* introduce 'build container' command [#36](https://github.com/syntasso/kratix-cli/issues/36) ([5882bb1](https://github.com/syntasso/kratix-cli/commit/5882bb1787c374ad51d3746188746791f900afb7))
* introduce 'init helm-promise' command ([#9](https://github.com/syntasso/kratix-cli/issues/9)) ([fd6e8d6](https://github.com/syntasso/kratix-cli/commit/fd6e8d655a544bc02cdfd39dedf833a210fca130))
* introduce language flag for 'add container' command [#121](https://github.com/syntasso/kratix-cli/issues/121) ([3d9c9c2](https://github.com/syntasso/kratix-cli/commit/3d9c9c2ca85d576ea5762edf3dfa5cde21f567f3))
* kratix update destination-selector ([#19](https://github.com/syntasso/kratix-cli/issues/19)) ([cf388d8](https://github.com/syntasso/kratix-cli/commit/cf388d82a23f14ec24f5c05c12d9e3c65fd69f09))
* output an informative message after init promise ([0a6c257](https://github.com/syntasso/kratix-cli/commit/0a6c257773b812256ff6ab8ecff96ec2d0b3b6be))
* refactor loadworkflows ([2a7c4bb](https://github.com/syntasso/kratix-cli/commit/2a7c4bb6ce37eebdf23c46083629416175f2b552))
* set label&annos from request in operator promise ([#64](https://github.com/syntasso/kratix-cli/issues/64)) ([c9af9c2](https://github.com/syntasso/kratix-cli/commit/c9af9c20a66002aaf3a3a2a6b2b73c26d81094bc))


### Bug Fixes

* 'add containers' does not allow for duplicated container names ([e6aee5e](https://github.com/syntasso/kratix-cli/commit/e6aee5ef08434c494e2f9e27a644cc441844aeaf))
* ([#37](https://github.com/syntasso/kratix-cli/issues/37)) handle null value in helm values files ([6e8a7d6](https://github.com/syntasso/kratix-cli/commit/6e8a7d6a4ed6abad9bbc48cc023aa2a566f46956))
* ([#37](https://github.com/syntasso/kratix-cli/issues/37)) handle null value in helm values files ([bef9200](https://github.com/syntasso/kratix-cli/commit/bef920047fab5691b61f8af446f8ff58c77523a8))
* ([#8](https://github.com/syntasso/kratix-cli/issues/8)) manage missing/empty api file on build ([8b52ad9](https://github.com/syntasso/kratix-cli/commit/8b52ad9dd38b48270a105d8900dc748f5caf4548))
* ([#8](https://github.com/syntasso/kratix-cli/issues/8)) remove --split from update dependencies ([5aad737](https://github.com/syntasso/kratix-cli/commit/5aad737769c50c157d3e45d75b3f8a3cb8858b7e))
* **#28:** ensure workflow.yaml is always an array of Pipelines ([115da10](https://github.com/syntasso/kratix-cli/commit/115da10432cf2c71e3166f80955d7826a057b43b))
* **#29:** better error message for invalid pipeline input ([7e94e0c](https://github.com/syntasso/kratix-cli/commit/7e94e0c1ee6675df9d16a6e583b4c6b786221c77))
* **#29:** better error message for invalid pipeline input ([da940d4](https://github.com/syntasso/kratix-cli/commit/da940d443c21a94a0920feedd2e73b7bfa8bfe53))
* **#2:** set resource_class to small ([b14a0cb](https://github.com/syntasso/kratix-cli/commit/b14a0cb1cc7fc901b5e324f88a736eee4b5c45bc))
* **#3:** test optional flags in `init promise` ([8d51a73](https://github.com/syntasso/kratix-cli/commit/8d51a7365a8839bdfd54d0ad430ea63128656949))
* add --split to persistent flags ([d1d1dbd](https://github.com/syntasso/kratix-cli/commit/d1d1dbd2f5086e71a3f0e805f0764e2f8af23575))
* add container --python templates ([#124](https://github.com/syntasso/kratix-cli/issues/124)) ([1f30310](https://github.com/syntasso/kratix-cli/commit/1f303108682a7563dfc76a2fe42532687c21c4f2))
* add release-please-config.json ([8b25c2e](https://github.com/syntasso/kratix-cli/commit/8b25c2e090d9c516087a37345b1b585cc20b28a2))
* better README and cli help usage message ([db2e1b6](https://github.com/syntasso/kratix-cli/commit/db2e1b6c03a1c23f7062b2c4a9e8f3c59684ecec))
* build a statically linked binary ([2dbcc96](https://github.com/syntasso/kratix-cli/commit/2dbcc9635f7a2e874973aa3083f1f2d3720a5622))
* build helm image multi arch ([410e279](https://github.com/syntasso/kratix-cli/commit/410e279c0e26b7db6d9b8a2ea9d64dad9b99e91d))
* bump go version to 1.23.0 ([40d6dfa](https://github.com/syntasso/kratix-cli/commit/40d6dfa6b5e284a2ff626754d12c57b1af56cc03))
* bump the resource class to medium when running tests ([2c976c8](https://github.com/syntasso/kratix-cli/commit/2c976c83e160479304ee5fc2c5358156baf408bb))
* can fetch chart from a specific version ([e708b8f](https://github.com/syntasso/kratix-cli/commit/e708b8f6745b21d38ac1a684f11d3f17c2b7b04f))
* can fetch chart from a specific version ([8f0d8eb](https://github.com/syntasso/kratix-cli/commit/8f0d8ebb9568cb2f435edbb63ff4e3523ac440ad))
* clean dir after installing goreleaser ([f1e70fd](https://github.com/syntasso/kratix-cli/commit/f1e70fd2308044c838fdcccc6bd2c85a853f4573))
* construct the correct docker push command ([8e6467e](https://github.com/syntasso/kratix-cli/commit/8e6467e98b34c2e30da67c7a9660748d50447de0))
* create and cleanup test dirs ([a0dd70e](https://github.com/syntasso/kratix-cli/commit/a0dd70e22353b55342d84202bf0ff8157e3e7885))
* create valid workflows for plain promise init ([3e979b6](https://github.com/syntasso/kratix-cli/commit/3e979b6d6fe0b994eacf82bebadc4c3285d24bf0))
* create valid workflows for plain promise init ([9d57553](https://github.com/syntasso/kratix-cli/commit/9d575537dde75697745b64d6f4a21fb15dc574b0))
* do not indent the pipeline.sh file content ([e19231a](https://github.com/syntasso/kratix-cli/commit/e19231a9636a19470fdc32a6706d025808b38fab))
* dont error when .spec isn't provided in terraform request ([88e030b](https://github.com/syntasso/kratix-cli/commit/88e030b71743c28955de7260c65dacea951f189f))
* ensure 'build promise' does not fail when there are no workflows and when logging that no promise.yaml has been found ([0dd8a90](https://github.com/syntasso/kratix-cli/commit/0dd8a907d951c2da6bf70f3ff6ef451cce894544))
* error gracefully when no workflows exist when building a container ([d0a2810](https://github.com/syntasso/kratix-cli/commit/d0a28104245eed60b5dc3afaac75a546b56d0ec9))
* error if the --name flag specifies a non-existent container name ([63aaefa](https://github.com/syntasso/kratix-cli/commit/63aaefafc20b3270bc6673ecf7b58a8d38f28bd9))
* handle multiple containers in the same pipeline ([165c075](https://github.com/syntasso/kratix-cli/commit/165c07508b60603c1b7c4262b3060c6153404191))
* handle no promise or workflow file existing ([725c8c2](https://github.com/syntasso/kratix-cli/commit/725c8c29c40a0a63596b1fd80fb90842d342d13e))
* improve help of update api subcommand ([42d6c46](https://github.com/syntasso/kratix-cli/commit/42d6c46d4da441c83801000b1e004e162049f658))
* include all dependency files from a directory when using --image ([3387b88](https://github.com/syntasso/kratix-cli/commit/3387b88a0e2c2d1673781edd1f3c9fcc2bab3a9c))
* include all dependency files from a directory when using --image ([cee117a](https://github.com/syntasso/kratix-cli/commit/cee117a654cef5c46db37bb7157357a515919bd9))
* include path to main.go in goreleaseer ([#73](https://github.com/syntasso/kratix-cli/issues/73)) ([600bd11](https://github.com/syntasso/kratix-cli/commit/600bd114c25301b3be603b3833b3cfa60050e8f3))
* increase no_output_timeout for helm operator tests ([29e4dac](https://github.com/syntasso/kratix-cli/commit/29e4dac408c5b413459d878b7c7a3d4ce9737d3c))
* make 'build container' command description more descriptive ([e6f5b99](https://github.com/syntasso/kratix-cli/commit/e6f5b994c125e9f374e4efe07476463b8acb409c))
* more error handling improvements for add container ([6a42202](https://github.com/syntasso/kratix-cli/commit/6a422023954064bef0e577e77a13cdeb0ea7da6e))
* proper treatment of dirs when building deps ([3249fc4](https://github.com/syntasso/kratix-cli/commit/3249fc4e87367dada8bfc3fa261045061ae5a82e))
* refactor parsing args logic ([e7ebf52](https://github.com/syntasso/kratix-cli/commit/e7ebf5233aa881e7046f549a92d93e6f46e37636))
* release v0.6.1 ([9fef9aa](https://github.com/syntasso/kratix-cli/commit/9fef9aa177fad0a8a38ca47c4bd33f6fda8b05fd))
* remove containers/ from path ([b0bde07](https://github.com/syntasso/kratix-cli/commit/b0bde0732ad78776dd7b764aefd2fe77e63b2076))
* remove duplicate 'type: object' from initialised Promise API ([3b6d97b](https://github.com/syntasso/kratix-cli/commit/3b6d97b9a1c38e4721de7a0866c0e96169890ebf))
* remove extra files from release tarball; remove previous release ([3a12b4f](https://github.com/syntasso/kratix-cli/commit/3a12b4fe980da02be21ecfad2ad55f79eac08d7e))
* remove workflow.yaml from init promise template ([20e4a7a](https://github.com/syntasso/kratix-cli/commit/20e4a7aa487dcfc7c03a365836ba3475327f57fe))
* revert v0.5.3 release in release-please-config.json ([65c00f1](https://github.com/syntasso/kratix-cli/commit/65c00f12d875861d49cdd719c43cfad29011d70f))
* stop on errors when parsing deps ([0daf7d2](https://github.com/syntasso/kratix-cli/commit/0daf7d2e2993acbd98246958e4b587ac7607de39))
* support empty openAPIV3Schema in xrd for the init-crossplane command ([#88](https://github.com/syntasso/kratix-cli/issues/88)) ([c4738ca](https://github.com/syntasso/kratix-cli/commit/c4738ca6e23b28922539da252f94c8c14dfa8f64))
* support empty spec.properties in xrd for the init-crossplane command ([d543cbf](https://github.com/syntasso/kratix-cli/commit/d543cbf17ae20fd8e1f05fdd63f2c2c44856392c))
* the 'promise build' command builds the workflows if workflow.yaml files are present ([0a20aa0](https://github.com/syntasso/kratix-cli/commit/0a20aa02c704a1275e8c7ae8b6391a6d235b338f))
* typo in helm promise aspect image name ([142e902](https://github.com/syntasso/kratix-cli/commit/142e902314fb2e2fb77c88fb75ea5025101324f5))
* typos and nil slice declaration ([3796c1c](https://github.com/syntasso/kratix-cli/commit/3796c1c323069091716026a29776428024d5504f))
* update api works with promise.yaml from build ([89c3427](https://github.com/syntasso/kratix-cli/commit/89c342733df63fb8a46727cc065d4749373c348c))
* update buildx example with build args ([2c69014](https://github.com/syntasso/kratix-cli/commit/2c6901448f81ba3faa41aec95b0e451433cd463e))
* update generated README from kratix init ([2960a5e](https://github.com/syntasso/kratix-cli/commit/2960a5e4e6ad554c76f950742a541ea389e420e0))
* use default namespace if not provided in operator resources ([d9d5a4b](https://github.com/syntasso/kratix-cli/commit/d9d5a4b0fb71a1439b1f95e4ede12b8f28940f6a))
* use the promise name from the cli ([3dc739d](https://github.com/syntasso/kratix-cli/commit/3dc739d02abcbcafa7d750fbf4a857641d64866b))
* work with container image without project name ([4d52288](https://github.com/syntasso/kratix-cli/commit/4d52288b7794b3f8841f26b63a7b5d4151ff21b2))
* WORKFLOWNAME -&gt; PIPELINENAME ([2e1d3a3](https://github.com/syntasso/kratix-cli/commit/2e1d3a399af7374215cf089e9c210b0eff90d311))
* write api.yaml in yaml format on update api ([b7aafac](https://github.com/syntasso/kratix-cli/commit/b7aafac074f814d5ff0575c56d08ca0ebdb5589e))
* write api.yaml in yaml format on update api ([e841e5f](https://github.com/syntasso/kratix-cli/commit/e841e5fc9362a50ea3a5d1474fa866ab386b45e5))


### Chores

* add e2e test for operator promise generation ([aac5229](https://github.com/syntasso/kratix-cli/commit/aac5229c2e7f93963255dc303cf563bcc6ffceb1))
* add required flags to init-helm help message ([#71](https://github.com/syntasso/kratix-cli/issues/71)) ([22a128c](https://github.com/syntasso/kratix-cli/commit/22a128c1323e8e434025427b7294676258f4b58a))
* adding check to ensure released version alignment ([728999d](https://github.com/syntasso/kratix-cli/commit/728999ded544f9099b8be9ba2be67f63fab981b1))
* aspect -&gt; stage ([#106](https://github.com/syntasso/kratix-cli/issues/106)) ([46a2dd7](https://github.com/syntasso/kratix-cli/commit/46a2dd782c53ae27fe4a37eebd0ffec45a7f568a))
* bump cli version ([0e4c3a2](https://github.com/syntasso/kratix-cli/commit/0e4c3a27042583cf1692312a275c6460998b0bde))
* bump cli version ([b8877d9](https://github.com/syntasso/kratix-cli/commit/b8877d967c73b83c0de012bf28448382cc61dd7e))
* bump cli version ([887685e](https://github.com/syntasso/kratix-cli/commit/887685e0a6be6a922e984115c53f90336dac2b9a))
* bump golang orb version ([99d11a4](https://github.com/syntasso/kratix-cli/commit/99d11a48d7415a0730bcbcf1ea5b8bc0ce99ef41))
* bump hardcoded cli version ([c738e90](https://github.com/syntasso/kratix-cli/commit/c738e90d0719aa66768b9daaa7fe0873f9c915d3))
* bump packages, use new kratix library helpers ([fa0143c](https://github.com/syntasso/kratix-cli/commit/fa0143ce364c66d4869a7401b7e529a74e53c249))
* bump timeout to get around ghcr rate limiting ([a44f252](https://github.com/syntasso/kratix-cli/commit/a44f252a5ef934fa1296978182801437f1e0d92a))
* checkout at a tag for goreleaser [skip ci] ([623812c](https://github.com/syntasso/kratix-cli/commit/623812ceac73d712ef81471dc3941c8d489da658))
* cleanup packages and unused imports ([ae8210b](https://github.com/syntasso/kratix-cli/commit/ae8210b9c9a947343161f08ab0cf5156d1fba4b4))
* document release process ([8f183da](https://github.com/syntasso/kratix-cli/commit/8f183dab75e5a8b70ecd18bcc221d04736b497f5))
* fix bash in release.yaml [skip ci] ([8852458](https://github.com/syntasso/kratix-cli/commit/885245834249995ab630e248c91d0a8c0aa88289))
* fix makefile ([7317a3f](https://github.com/syntasso/kratix-cli/commit/7317a3f9c934a5ed29e913ce1d0ab55dbb3cde39))
* fix release please github action [skip ci] ([29a9ce5](https://github.com/syntasso/kratix-cli/commit/29a9ce579965110265545b53a4d584a3d5f7d4a8))
* go mod tidy ([a1a27b0](https://github.com/syntasso/kratix-cli/commit/a1a27b09a5f562094a1a0faee4580dc19f5e3565))
* ignore release please branches ([947a0fe](https://github.com/syntasso/kratix-cli/commit/947a0fe2b04fe099fe53eb15b1c45094bf0a14a8))
* include chores in changelog ([d3fe2cc](https://github.com/syntasso/kratix-cli/commit/d3fe2cc2b2f1947a86ff460ae3a9825530cc0fec))
* **main:** release 0.1.0 ([2a7101c](https://github.com/syntasso/kratix-cli/commit/2a7101ca1556b7b900d5212889c8a03984a4c845))
* **main:** release 0.1.0 ([3d01749](https://github.com/syntasso/kratix-cli/commit/3d0174978ea14920a45227eadd96527840bc4ec5))
* **main:** release 0.1.0 ([3da0b0e](https://github.com/syntasso/kratix-cli/commit/3da0b0e6fe582b44d1043118494f4b96e775005f))
* **main:** release 0.1.0 ([da19408](https://github.com/syntasso/kratix-cli/commit/da194081c10f105632b7f9e333072dc8b989cc2f))
* **main:** release 0.1.0 ([a74b7e5](https://github.com/syntasso/kratix-cli/commit/a74b7e57664d72c7eff954b0e63e4fb838d8ccfc))
* **main:** release 0.1.0 ([1e91b9c](https://github.com/syntasso/kratix-cli/commit/1e91b9c27a9ed0e31fef0aec57e5580d35d51b3c))
* **main:** release 0.2.0 ([0a16c55](https://github.com/syntasso/kratix-cli/commit/0a16c55993a455174270a6ec45a6bb03da520dd2))
* **main:** release 0.2.0 ([6432ba4](https://github.com/syntasso/kratix-cli/commit/6432ba4a1a273f95a3579b60cfad06cb4e293c3e))
* **main:** release 0.2.1 ([#57](https://github.com/syntasso/kratix-cli/issues/57)) ([0b12bc9](https://github.com/syntasso/kratix-cli/commit/0b12bc9f2b57ec3b440de05752e788ff59b2df63))
* **main:** release 0.3.0 ([#62](https://github.com/syntasso/kratix-cli/issues/62)) ([9200af9](https://github.com/syntasso/kratix-cli/commit/9200af9a02ff18a831ada7dc5b12489111d1b297))
* **main:** release 0.4.0 ([4ab24a2](https://github.com/syntasso/kratix-cli/commit/4ab24a245458af289041d779e2eb86b50d3c076c))
* **main:** release 0.4.0 ([63f1b65](https://github.com/syntasso/kratix-cli/commit/63f1b65fc956e7cf5223bd9d4cd63d418653953a))
* **main:** release 0.5.0 ([e15c898](https://github.com/syntasso/kratix-cli/commit/e15c898761d0ebc1532955583c582e4e6de1b2e6))
* **main:** release 0.5.0 ([411bcd6](https://github.com/syntasso/kratix-cli/commit/411bcd6c76987c1094008cfaf3f5232ef721822c))
* **main:** release 0.5.1 ([#84](https://github.com/syntasso/kratix-cli/issues/84)) ([6d8a95f](https://github.com/syntasso/kratix-cli/commit/6d8a95f656f1d7e524046e34382f30a02ec406f8))
* **main:** release 0.5.2 ([d7d797e](https://github.com/syntasso/kratix-cli/commit/d7d797e5852a11fb2eee330729d1283ea4826bf0))
* **main:** release 0.5.2 ([daf931e](https://github.com/syntasso/kratix-cli/commit/daf931e5ea7fc19679f45a10fb5a1a99135b0524))
* **main:** release 0.5.3 ([#89](https://github.com/syntasso/kratix-cli/issues/89)) ([da1d94d](https://github.com/syntasso/kratix-cli/commit/da1d94d46bab44ab35dd148096a4d723dcf60ae2))
* **main:** release 0.5.3 ([#94](https://github.com/syntasso/kratix-cli/issues/94)) ([d30efe4](https://github.com/syntasso/kratix-cli/commit/d30efe435238907db02d22859b88ee5efc02f812))
* **main:** release 0.5.4 ([26c2ec9](https://github.com/syntasso/kratix-cli/commit/26c2ec9df2715f8453f02370fcca8345a8c8edd4))
* **main:** release 0.5.4 ([d405558](https://github.com/syntasso/kratix-cli/commit/d405558f3f0588a0ada3849abdcbd0c883921bbb))
* **main:** release 0.6.0 ([#105](https://github.com/syntasso/kratix-cli/issues/105)) ([b49a912](https://github.com/syntasso/kratix-cli/commit/b49a9120e48c973efa92bb503c42483abfc18ced))
* **main:** release 0.6.1 ([#114](https://github.com/syntasso/kratix-cli/issues/114)) ([0ebf260](https://github.com/syntasso/kratix-cli/commit/0ebf2601e0297cc14d09b47fa0a8b99f29a8fab9))
* **main:** release 0.7.0 ([#127](https://github.com/syntasso/kratix-cli/issues/127)) ([f9cfe11](https://github.com/syntasso/kratix-cli/commit/f9cfe11af105d5eb7392b6f9725737a059624859))
* **main:** release 0.8.0 ([#132](https://github.com/syntasso/kratix-cli/issues/132)) ([2449a71](https://github.com/syntasso/kratix-cli/commit/2449a716a1674f568f4abd54046f69ad958d7492))
* make it possible to trigger release manually ([d339bb8](https://github.com/syntasso/kratix-cli/commit/d339bb82814bcf5b741e9faff16472d3432d469e))
* migrate CI workflows to GH Actions ([#96](https://github.com/syntasso/kratix-cli/issues/96)) ([3b08cac](https://github.com/syntasso/kratix-cli/commit/3b08cac9a8bd09213130b6274cbca3bf5fdf6a67))
* minor change to helm-promise output ([7f438ad](https://github.com/syntasso/kratix-cli/commit/7f438ad0cadae23d4a3dfccab0b3a6f51aeabed6))
* refactorings ([862d56b](https://github.com/syntasso/kratix-cli/commit/862d56b207d7327eaa8957f343d5d965db3480db))
* release 0.1.0 ([d2fbcb9](https://github.com/syntasso/kratix-cli/commit/d2fbcb9327b96919b33de37439debaa940cdd510))
* release fixes to cut 0.5.4 correctly ([573b575](https://github.com/syntasso/kratix-cli/commit/573b5759565817bb11e756f401e3cb7305dfede3))
* remove references to bitnami ([#131](https://github.com/syntasso/kratix-cli/issues/131)) ([80488db](https://github.com/syntasso/kratix-cli/commit/80488db934c379b5b76d81c6b49d9694b9c9e0f7))
* rename workflow -&gt; lifecycle for consistency ([0580266](https://github.com/syntasso/kratix-cli/commit/0580266c775625932d6f261f1632a33d4cafedae))
* run check on release-please branches ([671e053](https://github.com/syntasso/kratix-cli/commit/671e053f83d9fd8d9cc340548b5ccdd2baf87c20))
* set version in code ([1de09d2](https://github.com/syntasso/kratix-cli/commit/1de09d2dc38f0c61f7989d36f62e24b40084917c))
* setup automated releases ([6c96ff0](https://github.com/syntasso/kratix-cli/commit/6c96ff0dc5a88ba3466b9d0b2c354afdec0e0d00))
* small markups ([baf1aae](https://github.com/syntasso/kratix-cli/commit/baf1aaeebfb7728164fa441d861dbd0e9b3dece7))
* stabilise flakey test ([5366542](https://github.com/syntasso/kratix-cli/commit/53665420f575b27050ac223002a084810524d968))
* update build and install steps in README ([ba096c7](https://github.com/syntasso/kratix-cli/commit/ba096c790c2d63c78aee98318d4364ff544a08b5))
* upgrade kratix dependency ([#108](https://github.com/syntasso/kratix-cli/issues/108)) ([98b3fe6](https://github.com/syntasso/kratix-cli/commit/98b3fe6e308a495b4b0473c0d6a6f9df16c22374))
* use the RELASE_CREATOR_TOKEN when creating release PRs ([3b25ff0](https://github.com/syntasso/kratix-cli/commit/3b25ff06816b3a9d055c32ff5e1825c1829105eb))
* use the RELASE_CREATOR_TOKEN when creating release PRs ([62a2f80](https://github.com/syntasso/kratix-cli/commit/62a2f803bbd0b238ab192747e60026a94957703d))


### Build System

* **deps:** bump github.com/containerd/containerd from 1.7.24 to 1.7.27 ([e3b9df0](https://github.com/syntasso/kratix-cli/commit/e3b9df013a3c15ba7c78609aaa49590c240474fb))
* **deps:** bump github.com/containerd/containerd from 1.7.24 to 1.7.27 ([4bf80a6](https://github.com/syntasso/kratix-cli/commit/4bf80a62cc47d4ffa5295545391bdef0a0619f4c))
* **deps:** bump github.com/docker/docker ([#58](https://github.com/syntasso/kratix-cli/issues/58)) ([d33c8b3](https://github.com/syntasso/kratix-cli/commit/d33c8b3ae516570d8b529d8f96c2d596dc66a7e2))
* **deps:** bump github.com/hashicorp/go-getter from 1.7.8 to 1.7.9 ([#129](https://github.com/syntasso/kratix-cli/issues/129)) ([a36040b](https://github.com/syntasso/kratix-cli/commit/a36040b03c7d58b88732ed451fe9c40d698465df))
* **deps:** bump golang.org/x/net from 0.37.0 to 0.38.0 ([69799fb](https://github.com/syntasso/kratix-cli/commit/69799fb76fa0c0da464f2ebf8aabacc025afe748))
* **deps:** bump golang.org/x/net from 0.37.0 to 0.38.0 ([e3fe824](https://github.com/syntasso/kratix-cli/commit/e3fe8246b4825e46fcb5f060201fb8b421912053))
* **deps:** bump golang.org/x/oauth2 from 0.24.0 to 0.27.0 ([a0320e2](https://github.com/syntasso/kratix-cli/commit/a0320e26e2dc179bd90b7bcc5fcdeb348178a74b))
* **deps:** bump golang.org/x/oauth2 from 0.24.0 to 0.27.0 ([959c175](https://github.com/syntasso/kratix-cli/commit/959c1753caf43182375b7e764416064dce4eb341))
* **deps:** bump helm.sh/helm/v3 from 3.15.2 to 3.17.3 ([461d877](https://github.com/syntasso/kratix-cli/commit/461d8774b5864645bddd9660bb3cb1decc3a5995))
* **deps:** bump helm.sh/helm/v3 from 3.15.2 to 3.17.3 ([735f500](https://github.com/syntasso/kratix-cli/commit/735f50016c533fc6bf05c5df25cda2978ba4e7d0))
* **deps:** bump helm.sh/helm/v3 from 3.17.3 to 3.17.4 ([e4347b8](https://github.com/syntasso/kratix-cli/commit/e4347b8cfccdae284932854feff52f607d140e96))
* **deps:** bump helm.sh/helm/v3 from 3.17.3 to 3.17.4 ([e415576](https://github.com/syntasso/kratix-cli/commit/e4155768e05708092de5bee678ba851ad96b3cc7))
* **deps:** bump helm.sh/helm/v3 from 3.17.4 to 3.18.5 ([#125](https://github.com/syntasso/kratix-cli/issues/125)) ([4a21ada](https://github.com/syntasso/kratix-cli/commit/4a21ada20aecb9a3ce8d42676e9ba662d18b0acf))

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
