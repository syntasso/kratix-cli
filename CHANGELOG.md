# Changelog

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
