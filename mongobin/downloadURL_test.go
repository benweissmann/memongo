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

package mongobin_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tryvium-travels/memongo/mongobin"
)

// Change this to true to issue a HEAD request in each test to make
// sure the file is there and accessible. We leave this off for reliability,
// but it can be turned on if you want to test that the generated URLs point
// to real files.
const testHTTPHead = true

func TestGetDownloadURL(t *testing.T) {
	mongoVersionsToTest := []string{
		"3.2.0", "3.2.22", "3.4.0", "3.4.19", "3.6.0", "3.6.10", "4.0.0", "4.0.13", "4.2.1",
	}

	tests := map[string]struct {
		spec          *mongobin.DownloadSpec
		mongoVersions []string

		expectedURL string
	}{
		"mac-ssl": {
			spec: &mongobin.DownloadSpec{
				Platform:       "osx",
				Arch:           "x86_64",
				OSName:         "",
				SSLBuildNeeded: true,
			},
			mongoVersions: []string{
				"3.2.0", "3.2.22", "3.4.0", "3.4.19", "3.6.0", "3.6.10", "4.0.0", "4.0.13",
			},
			expectedURL: "https://fastdl.mongodb.org/osx/mongodb-osx-ssl-x86_64-VERSION.tgz",
		},
		"mac": {
			spec: &mongobin.DownloadSpec{
				Platform: "osx",
				Arch:     "x86_64",
				OSName:   "",
			},
			mongoVersions: []string{
				"4.2.1",
			},
			expectedURL: "https://fastdl.mongodb.org/osx/mongodb-macos-x86_64-VERSION.tgz",
		},
		"ubuntu 18.04": {
			spec: &mongobin.DownloadSpec{
				Platform: "linux",
				Arch:     "x86_64",
				OSName:   "ubuntu1804",
			},
			mongoVersions: []string{"4.0.1", "4.0.13", "4.2.1"},
			expectedURL:   "https://fastdl.mongodb.org/linux/mongodb-linux-x86_64-ubuntu1804-VERSION.tgz",
		},
		"ubuntu 16.04": {
			spec: &mongobin.DownloadSpec{
				Platform: "linux",
				Arch:     "x86_64",
				OSName:   "ubuntu1604",
			},
			mongoVersions: []string{
				"3.2.7", "3.4.0", "3.4.19", "3.6.0", "3.6.10", "4.0.0", "4.0.13", "4.2.1",
			},
			expectedURL: "https://fastdl.mongodb.org/linux/mongodb-linux-x86_64-ubuntu1604-VERSION.tgz",
		},
		"ubuntu 14.04": {
			spec: &mongobin.DownloadSpec{
				Platform: "linux",
				Arch:     "x86_64",
				OSName:   "ubuntu1404",
			},
			mongoVersions: []string{
				"3.2.0", "3.2.22", "3.4.0", "3.4.19", "3.6.0", "3.6.10", "4.0.0", "4.0.13",
			},
			expectedURL: "https://fastdl.mongodb.org/linux/mongodb-linux-x86_64-ubuntu1404-VERSION.tgz",
		},
		"SUSE 12": {
			spec: &mongobin.DownloadSpec{
				Platform: "linux",
				Arch:     "x86_64",
				OSName:   "suse12",
			},
			mongoVersions: mongoVersionsToTest,
			expectedURL:   "https://fastdl.mongodb.org/linux/mongodb-linux-x86_64-suse12-VERSION.tgz",
		},
		"RHEL 7": {
			spec: &mongobin.DownloadSpec{
				Platform: "linux",
				Arch:     "x86_64",
				OSName:   "rhel70",
			},
			mongoVersions: mongoVersionsToTest,
			expectedURL:   "https://fastdl.mongodb.org/linux/mongodb-linux-x86_64-rhel70-VERSION.tgz",
		},
		"RHEL 6": {
			spec: &mongobin.DownloadSpec{
				Platform: "linux",
				Arch:     "x86_64",
				OSName:   "rhel62",
			},
			mongoVersions: mongoVersionsToTest,
			expectedURL:   "https://fastdl.mongodb.org/linux/mongodb-linux-x86_64-rhel62-VERSION.tgz",
		},
		"Debian stretch": {
			spec: &mongobin.DownloadSpec{
				Platform: "linux",
				Arch:     "x86_64",
				OSName:   "debian92",
			},
			mongoVersions: []string{
				"3.6.5", "3.6.10", "4.0.0", "4.0.13", "4.2.1",
			},
			expectedURL: "https://fastdl.mongodb.org/linux/mongodb-linux-x86_64-debian92-VERSION.tgz",
		},
		"Debian jessie": {
			spec: &mongobin.DownloadSpec{
				Platform: "linux",
				Arch:     "x86_64",
				OSName:   "debian81",
			},
			mongoVersions: []string{
				"3.2.8", "3.2.22", "3.4.0", "3.4.19", "3.6.0", "3.6.10", "4.0.0", "4.0.13",
			},
			expectedURL: "https://fastdl.mongodb.org/linux/mongodb-linux-x86_64-debian81-VERSION.tgz",
		},
		"Amazon Linux": {
			spec: &mongobin.DownloadSpec{
				Platform: "linux",
				Arch:     "x86_64",
				OSName:   "amazon",
			},
			mongoVersions: mongoVersionsToTest,
			expectedURL:   "https://fastdl.mongodb.org/linux/mongodb-linux-x86_64-amazon-VERSION.tgz",
		},
		"Amazon Linux 2": {
			spec: &mongobin.DownloadSpec{
				Platform: "linux",
				Arch:     "x86_64",
				OSName:   "amazon2",
			},
			mongoVersions: []string{
				"4.0.0", "4.0.13", "4.2.1",
			},
			expectedURL: "https://fastdl.mongodb.org/linux/mongodb-linux-x86_64-amazon2-VERSION.tgz",
		},
		"Other Linux": {
			spec: &mongobin.DownloadSpec{
				Platform: "linux",
				Arch:     "x86_64",
				OSName:   "",
			},
			mongoVersions: []string{
				"3.2.0", "3.2.22", "3.4.0", "3.4.19", "3.6.0", "3.6.10", "4.0.0", "4.0.13",
			},
			expectedURL: "https://fastdl.mongodb.org/linux/mongodb-linux-x86_64-VERSION.tgz",
		},
	}

	for testName, test := range tests {
		for _, mongoVersion := range test.mongoVersions {
			t.Run(testName+"_"+mongoVersion, func(t *testing.T) {
				spec := &mongobin.DownloadSpec{
					Version:        mongoVersion,
					Platform:       test.spec.Platform,
					Arch:           test.spec.Arch,
					OSName:         test.spec.OSName,
					SSLBuildNeeded: test.spec.SSLBuildNeeded,
				}

				expectedURL := strings.Replace(test.expectedURL, "VERSION", mongoVersion, -1)
				actualURL := spec.GetDownloadURL()

				if testHTTPHead {
					resp, err := http.Head(actualURL)
					assert.NoError(t, err)
					assert.Equal(t, 200, resp.StatusCode)
				}

				assert.Equal(t, expectedURL, actualURL)
			})
		}
	}
}
