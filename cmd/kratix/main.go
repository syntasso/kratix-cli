/*
Copyright 2024 Syntasso.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"github.com/syntasso/kratix-cli/cmd"
)

// release-please updates this version when it prepares a release (see extra-files
// in release-please-config.json), so it should not be edited by hand. `make
// check-version-alignment` fails the release PR if it does not match the manifest.
var version = "0.19.3" // x-release-please-version

func main() {
	cmd.Execute(version)
}
