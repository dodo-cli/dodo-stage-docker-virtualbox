![release](https://img.shields.io/github/v/release/wabenet/dodo-stage-virtualbox?sort=semver)
![build](https://img.shields.io/github/workflow/status/wabenet/dodo-stage-virtualbox/CI?logo=github)
![codecov](https://img.shields.io/codecov/c/github/wabenet/dodo-stage-virtualbox?logo=codecov)
![codeclimate](https://img.shields.io/codeclimate/maintainability/wabenet/dodo-stage-virtualbox?logo=codeclimate)
![license](https://img.shields.io/github/license/wabenet/dodo-stage-virtualbox)

# dodo stage-virtualbox plugin

Adds support for VirtualBox as a dodo stage plugin.

## installation

If you want to compile your own dodo distribution, you can add this plugin with the
following generate config:

```yaml
plugins:
  - import: github.com/wabenet/dodo-stage-virtualbox/pkg/plugin
```

Alternatively, you can install it as a standalone plugin by downloading the
correct file for your system from the [releases page](https://github.com/wabenet/dodo-stage-virtualbox/releases),
then copy it into the dodo plugin directory (`${HOME}/.dodo/plugins`).

## license & authors

```text
Copyright 2022 Ole Claussen

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
```
