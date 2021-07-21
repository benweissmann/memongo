// Copyright 2021 Tryvium Travels LTD
// Copyright 2019-2020 Ben Weissmann
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package mongobin

import "fmt"

// GetDownloadURL returns the download URL to download the binary
// from the MongoDB website
func (spec *DownloadSpec) GetDownloadURL() string {
	archiveName := "mongodb-"

	if spec.Platform == "linux" {
		archiveName += "linux-" + spec.Arch + "-"

		if spec.OSName != "" {
			archiveName += spec.OSName + "-"
		}

		archiveName += spec.Version + ".tgz"
	} else {
		if spec.SSLBuildNeeded {
			archiveName += "osx-ssl-"
		} else {
			archiveName += "macos-"
		}

		archiveName += spec.Arch + "-" + spec.Version + ".tgz"
	}

	return fmt.Sprintf(
		"https://fastdl.mongodb.org/%s/%s",
		spec.Platform,
		archiveName,
	)
}
