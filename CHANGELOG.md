## Changelog

### v0.30.0 / 2020-12-07
- [#222](https://github.com/eremetic-framework/eremetic/pull/222) Random stuck (@ja8zyjits)
- [#218](https://github.com/eremetic-framework/eremetic/pull/218) added ca-certificates in Dockerfile (@sachinpk46)
- [#219](https://github.com/eremetic-framework/eremetic/pull/219) Bump go version used to 1.13 (@keis)
- [#216](https://github.com/eremetic-framework/eremetic/pull/216) remove deprecated -e flag from docker login (@alde)

### v0.29.0 / 2019-06-25
- [#213](https://github.com/eremetic-framework/eremetic/pull/213) Update to go 1.12 with modules (@alde)
- [#199](https://github.com/eremetic-framework/eremetic/pull/199) run go vet and golint in travis (@alde)
- [#198](https://github.com/eremetic-framework/eremetic/pull/198) Fix reportcard issues (@alde)
- [#193](https://github.com/eremetic-framework/eremetic/pull/193) Add support to filter tasks by name and list terminated tasks (@zmalik)
- [#196](https://github.com/eremetic-framework/eremetic/pull/196) Add support for labels (@zmalik)
- [#195](https://github.com/eremetic-framework/eremetic/pull/195) Drop support for golang 1.5 (@keis)
- [#192](https://github.com/eremetic-framework/eremetic/pull/192) Remove unused tasksCreated counter (@keis)
- [#189](https://github.com/eremetic-framework/eremetic/pull/189) Check launch tasks (@keis)
- [#190](https://github.com/eremetic-framework/eremetic/pull/190) Allow user specified task names (@zmalik)
- [#186](https://github.com/eremetic-framework/eremetic/pull/186) Add docker 'volumes-from' option support (#186) (@waynz0r)
- [#183](https://github.com/eremetic-framework/eremetic/pull/183) Add a 'privileged' flag to allow running privileged containers in eremetic. (@cybricio)
- [#182](https://github.com/eremetic-framework/eremetic/pull/182) Update changelog (@alde)

### v0.28.0 / 2017-06-23
- [#178](https://github.com/eremetic-framework/eremetic/pull/178) Add /api/v1/ routes (@alde)
- [#180](https://github.com/eremetic-framework/eremetic/pull/180) Test data builders (@keis)
- [#177](https://github.com/eremetic-framework/eremetic/pull/177) Switch to glide (@eremetic-framework)
- [#179](https://github.com/eremetic-framework/eremetic/pull/179) Build eremetic on 1.7 (non-rc) and 1.8 (@eremetic-framework)
- [#176](https://github.com/eremetic-framework/eremetic/pull/176) Scheduler driver restructure (@keis)
- [#174](https://github.com/eremetic-framework/eremetic/pull/174) Internal rename of slave -> agent (@keis)
- [#175](https://github.com/eremetic-framework/eremetic/pull/175) Update year in license (@alde)
- [#173](https://github.com/eremetic-framework/eremetic/pull/173) Move metric definitions to new package (@keis)
- [#170](https://github.com/eremetic-framework/eremetic/pull/170) upgrade version of base alpine image (@gorelikov)
- [#168](https://github.com/eremetic-framework/eremetic/pull/168) Basic support for maintenance mode of mesos (@zmalik)

### v0.27.0 / 2017-04-19
- [#166](https://github.com/eremetic-framework/eremetic/pull/166) Add URL Prefix support for DCOS package (@gisjedi)
- [#165](https://github.com/eremetic-framework/eremetic/pull/165) Adding icon assets to support DCOS Universe package. (@gisjedi)
- [#162](https://github.com/eremetic-framework/eremetic/pull/162) Update the vendored mesos-go (@keis)
- [#159](https://github.com/eremetic-framework/eremetic/pull/159) Add support for network host mode and pass dns to the docker (#159) (@zmalik)
- [#158](https://github.com/eremetic-framework/eremetic/pull/158) Add args and uris in the hermit client (@zmalik)
- [#157](https://github.com/eremetic-framework/eremetic/pull/157) add basic auth support (@heww)
- [#154](https://github.com/eremetic-framework/eremetic/pull/154) Add support for environment variables in hermit (@zmalik)
- [#152](https://github.com/eremetic-framework/eremetic/pull/152) Implemented delete task and added test for zk and bolt (@zmalik)
- [#140](https://github.com/eremetic-framework/eremetic/pull/140) Add killswitch for tasks (@alde)

### v0.26.0 / 2016-10-04
- [#149](https://github.com/eremetic-framework/eremetic/pull/149) Add port mapping when allocating unspecified port (@keis)
- [#147](https://github.com/eremetic-framework/eremetic/pull/147) Fix linter errors (@marcusolsson)
- [#146](https://github.com/eremetic-framework/eremetic/pull/146) Add Args (@Mongey)

### v0.25.0 / 2016-09-28
- [#143](https://github.com/eremetic-framework/eremetic/pull/143) Add host ports to task environment (@marcusolsson)
- [#144](https://github.com/eremetic-framework/eremetic/pull/144) Add usage help for hermit (@alde)
- [#137](https://github.com/eremetic-framework/eremetic/pull/137) hermit: Allow filtering output (@marcusolsson, @alde)
- [#142](https://github.com/eremetic-framework/eremetic/pull/142) Merge scheduler into eremetic pkg (@marcusolsson)
- [#141](https://github.com/eremetic-framework/eremetic/pull/141) Update readme with goreportcard (@alde)
- [#133](https://github.com/eremetic-framework/eremetic/pull/133) Add hermit cli (@marcusolsson)
- [#136](https://github.com/eremetic-framework/eremetic/pull/136) List tasks empty db (@keis)

### v0.24.1 / 2016-09-12
- [#134](https://github.com/eremetic-framework/eremetic/pull/134) Actually build the binary with make docker (@alde)

### v0.24.0 / 2016-09-12
- [#129](https://github.com/eremetic-framework/eremetic/pull/129) Extract dependency packages (@marcusolsson)
- [#131](https://github.com/eremetic-framework/eremetic/pull/131) Add callback tests and check for missing status (@marcusolsson)
- [#132](https://github.com/eremetic-framework/eremetic/pull/132) Extract scheduler tests into smaller ones (@marcusolsson)
- [#127](https://github.com/eremetic-framework/eremetic/pull/127) Fix readme (@alde)
- [#126](https://github.com/eremetic-framework/eremetic/pull/126) Refactor package layout (@marcusolsson)
- [#125](https://github.com/eremetic-framework/eremetic/pull/125) Version is not configurable (@keis)
- [#124](https://github.com/eremetic-framework/eremetic/pull/124) Remove dependency on spf13/viper and m4rw3r/uuid (@alde)
- [#123](https://github.com/eremetic-framework/eremetic/pull/123) Status view (@keis)
- [#121](https://github.com/eremetic-framework/eremetic/pull/121) Unify the building of routes (@alde)
- [#119](https://github.com/eremetic-framework/eremetic/pull/119) add port mapping (@mrares)
- [#120](https://github.com/eremetic-framework/eremetic/pull/120) Refactor createTaskInfo (@alde)
- [#107](https://github.com/eremetic-framework/eremetic/pull/107) Build with 1.7rc on travis (@keis)
- [#114](https://github.com/eremetic-framework/eremetic/pull/114) Add task queued state (@keis)
- [#115](https://github.com/eremetic-framework/eremetic/pull/115) Return empty array instead of `null` when tasks list is empty (@sheepkiller)
- [#112](https://github.com/eremetic-framework/eremetic/pull/112) Vendor all dependencies (@alde)
- [#113](https://github.com/eremetic-framework/eremetic/pull/113) fix ForcePullImage and add unit test (@sheepkiller)
- [#109](https://github.com/eremetic-framework/eremetic/pull/109) Refactor uris/fetch (@sheepkiller)
- [#108](https://github.com/eremetic-framework/eremetic/pull/108) Expose eremetic version (@alde)

### v0.23.0 / 2016-08-13
- [#106](https://github.com/eremetic-framework/eremetic/pull/106) Add ability to force docker to pull image before launching each task (@sheepkiller)
- [#105](https://github.com/eremetic-framework/eremetic/pull/105) Set sandbox path from reconcile status update (@keis)
- [#104](https://github.com/eremetic-framework/eremetic/pull/104) Internal refactor of offer matching on attr (@keis)
- [#102](https://github.com/eremetic-framework/eremetic/pull/102) Add `fetch` to minic marathon management of URIs, i.e. fine managemen… (@sheepkiller)

### v0.22.0 / 2016-07-15
- [#94](https://github.com/eremetic-framework/eremetic/pull/94) Add zookeeper database backend (@alde)

### v0.21.0 / 2016-06-14
- [#92](https://github.com/eremetic-framework/eremetic/pull/92) Task attribute constraints (@chuckg)

### v0.20.0 / 2016-06-09
- [#91](https://github.com/eremetic-framework/eremetic/pull/91) Provide access to sandbox logs (@alde)

### v0.19.1 / 2016-06-08
- [#90](https://github.com/eremetic-framework/eremetic/pull/90) Create a database abstraction (@alde)
- [#84](https://github.com/eremetic-framework/eremetic/pull/84) Update purpose in the readme (@alde)
- [#87](https://github.com/eremetic-framework/eremetic/pull/87) remove Name from request type (@wstrucke)
- [#86](https://github.com/eremetic-framework/eremetic/pull/86) Wipe out viper from scheduler module (@keis)

### v0.19.0 / 2016-05-16
- [#82](https://github.com/eremetic-framework/eremetic/pull/82) add route to list all running tasks (@justinclayton)
- [#85](https://github.com/eremetic-framework/eremetic/pull/85) More tests (@keis)
- [#79](https://github.com/eremetic-framework/eremetic/pull/79) Enable coveralls (@keis)

### v0.18.0 / 2016-05-11
- [#83](https://github.com/eremetic-framework/eremetic/pull/83) Make the queue size configurable (@alde)

### v0.17.0 / 2016-04-30
- [#81](https://github.com/eremetic-framework/eremetic/pull/81) Adding authentication support (#81) (@mcgin)
- [#77](https://github.com/eremetic-framework/eremetic/pull/77) Exit eremetic when framework driver closes (@keis)
- [#76](https://github.com/eremetic-framework/eremetic/pull/76) Add example usage (@keis)

### v0.16.2 / 2016-03-10
- [#75](https://github.com/eremetic-framework/eremetic/pull/75) Only use one protobuf library (@alde)

### v0.16.1 / 2016-03-08
- [#72](https://github.com/eremetic-framework/eremetic/pull/72) Make callback and metrics aware of retries (@keis)

### v0.16.0 / 2016-03-04
- [#70](https://github.com/eremetic-framework/eremetic/pull/70) Add support for Masked Environment variables (@alde)

### v0.15.1 / 2016-02-29
- [#67](https://github.com/eremetic-framework/eremetic/pull/67) Fix reconcile tasks test (@alde)

### v0.15.0 / 2016-02-29
- [#66](https://github.com/eremetic-framework/eremetic/pull/66) Run default docker command (@alde)
- [#65](https://github.com/eremetic-framework/eremetic/pull/65) Error on full queue (@keis)

### v0.14.1 / 2016-02-04
- [#61](https://github.com/eremetic-framework/eremetic/pull/61) Use a OS-appropriate temp dir for the test database (@alde)

### v0.14.0 / 2016-01-30
- [#60](https://github.com/eremetic-framework/eremetic/pull/60) Switch to logrus (@alde)

### v0.13.0 / 2016-01-25
- [#58](https://github.com/eremetic-framework/eremetic/pull/58) Use absolute url in location header (@keis)

### v0.12.0 / 2016-01-15
- [#54](https://github.com/eremetic-framework/eremetic/pull/54) Implement a new landing page (@alde)

### v0.11.1 / 2016-01-15
- [#53](https://github.com/eremetic-framework/eremetic/pull/53) Improve the Error message from POSTing a json (@alde)
- [#52](https://github.com/eremetic-framework/eremetic/pull/52) Add swagger describing the api (@keis)

### v0.11.0 / 2016-01-11
- [#50](https://github.com/eremetic-framework/eremetic/pull/50) Support adding uris to command info (@keis)
- [#49](https://github.com/eremetic-framework/eremetic/pull/49) Remove protobuf members from task struct (@keis)

### v0.10.0 / 2015-12-29
- [#48](https://github.com/eremetic-framework/eremetic/pull/48) Notify callback (@alde)
- [#44](https://github.com/eremetic-framework/eremetic/pull/44) Fix AssetInfo after update to go-bindata-assetfs (@alde)
- [#43](https://github.com/eremetic-framework/eremetic/pull/43) Extend metrics with a running tasks gauge (@keis)
- [#42](https://github.com/eremetic-framework/eremetic/pull/42) Metrics (@keis)

### v0.9.1 / 2015-12-15
- [#41](https://github.com/eremetic-framework/eremetic/pull/41) Publish tagged version to docker (@keis)

### v0.9.0 / 2015-12-15
- [#39](https://github.com/eremetic-framework/eremetic/pull/39) Reconcile (@keis)
- [#38](https://github.com/eremetic-framework/eremetic/pull/38) Fix messenger address in container (@gengmao)
- [#37](https://github.com/eremetic-framework/eremetic/pull/37) Don't modify the input when creating mesos task (@keis)
- [#35](https://github.com/eremetic-framework/eremetic/pull/35) Update readme with better build instructions (@alde)
- [#28](https://github.com/eremetic-framework/eremetic/pull/28) Publish docker (@keis)
- [#31](https://github.com/eremetic-framework/eremetic/pull/31) Filter offers (@keis)

### v0.8.0 / 2015-11-23
- [#30](https://github.com/eremetic-framework/eremetic/pull/30) Use bindata-assetfs to bundle assets and templates (@alde)

### v0.7.0 / 2015-11-20
- [#29](https://github.com/eremetic-framework/eremetic/pull/29) Restart task that failed to start (@keis)
- [#27](https://github.com/eremetic-framework/eremetic/pull/27) Change Status to a slice containing Status structs (@alde)

### v0.6.0 / 2015-11-18
- [#26](https://github.com/eremetic-framework/eremetic/pull/26) extract the scheduler from handler package (@keis)

### v0.5.0 / 2015-11-13
- [#24](https://github.com/eremetic-framework/eremetic/pull/24) Implement support for BoltDB database backing (@alde)
- [#23](https://github.com/eremetic-framework/eremetic/pull/23) remove some dead code (@keis)

### v0.4.0 / 2015-11-12
- [#22](https://github.com/eremetic-framework/eremetic/pull/22) automatic version based on git tag (@keis)
- [#16](https://github.com/eremetic-framework/eremetic/pull/16) Add support for text/html Accept header (@alde)
- [#19](https://github.com/eremetic-framework/eremetic/pull/19) always use task structure from map (@keis)
- [#20](https://github.com/eremetic-framework/eremetic/pull/20) Add task id as environment variable (@keis)
- [#18](https://github.com/eremetic-framework/eremetic/pull/18) Add logo (@alde)
- [#17](https://github.com/eremetic-framework/eremetic/pull/17) Actually exit after catching an interrupt (@alde)
- [#15](https://github.com/eremetic-framework/eremetic/pull/15) adding docker file (@alde, @keis)
- [#14](https://github.com/eremetic-framework/eremetic/pull/14) initialise status to `staging` (@keis)
- [#10](https://github.com/eremetic-framework/eremetic/pull/10) move creation of task structure to request handler (@keis)
- [#8](https://github.com/eremetic-framework/eremetic/pull/8) More data more better (@keis)
- [#7](https://github.com/eremetic-framework/eremetic/pull/7) Add tests (@alde)
- [#5](https://github.com/eremetic-framework/eremetic/pull/5) use default master detector over zook (@keis)
- [#4](https://github.com/eremetic-framework/eremetic/pull/4) Split task scheduler (@keis)
- [#3](https://github.com/eremetic-framework/eremetic/pull/3) make mesos port and published address configurable (@keis)
- [#2](https://github.com/eremetic-framework/eremetic/pull/2) enable automatic configuration from environment (@keis)
- [#1](https://github.com/eremetic-framework/eremetic/pull/1) detect master to use by lowest node id (@keis)
