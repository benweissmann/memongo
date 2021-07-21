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

import (
	"io/ioutil"
	"runtime"
	"strconv"
	"strings"

	"github.com/acobaugh/osrelease"
)

// We define these as package vars so we can override it in tests
var EtcOsRelease = "/etc/os-release"
var EtcRedhatRelease = "/etc/redhat-release"
var GoOS = runtime.GOOS
var GoArch = runtime.GOARCH

// DownloadSpec specifies what copy of MongoDB to download
type DownloadSpec struct {
	// Version is what version of MongoDB to download
	Version string

	// Platform is "osx" or "linux"
	Platform string

	// SSLBuildNeeded is "ssl" if we need to download the SSL build for macOS
	// (needed for <4.2)
	SSLBuildNeeded bool

	// Arch is always "x86_64"
	Arch string

	// OSName is one of:
	// - ubuntu1804
	// - ubuntu1604
	// - ubuntu1404
	// - debian92
	// - debian81
	// - suse12
	// - rhel70
	// - rhel62
	// - amazon
	// - amazon2
	// - "" for other linux or for MacOS
	OSName string
}

// MakeDownloadSpec returns a DownloadSpec for the current operating system
func MakeDownloadSpec(version string) (*DownloadSpec, error) {
	parsedVersion, versionErr := parseVersion(version)
	if versionErr != nil {
		return nil, versionErr
	}

	platform, platformErr := detectPlatform()
	if platformErr != nil {
		return nil, platformErr
	}

	ssl := false
	if platform == "osx" && !versionGTE(parsedVersion, []int{4, 2, 0}) {
		// pre-4.0, the MacOS builds had a special "ssl" designator in the URL
		ssl = true
	}

	arch, archErr := detectArch()
	if archErr != nil {
		return nil, archErr
	}

	osName := detectOSName(parsedVersion)

	if platform == "linux" && osName == "" && versionGTE(parsedVersion, []int{4, 2, 0}) {
		return nil, &UnsupportedSystemError{msg: "MongoDB 4.2 removed support for generic linux tarballs. Specify the download URL manually or use a supported distro. See: https://www.mongodb.com/blog/post/a-proposal-to-endoflife-our-generic-linux-tar-packages"}
	}

	return &DownloadSpec{
		Version:        version,
		Arch:           arch,
		SSLBuildNeeded: ssl,
		Platform:       platform,
		OSName:         osName,
	}, nil
}

func parseVersion(version string) ([]int, error) {
	versionParts := strings.Split(version, ".")
	if len(versionParts) < 3 {
		return nil, &UnsupportedMongoVersionError{
			version: version,
			msg:     "MongoDB version number must be in the form x.y.z",
		}
	}

	majorVersion, majErr := strconv.Atoi(versionParts[0])
	if majErr != nil {
		return nil, &UnsupportedMongoVersionError{
			version: version,
			msg:     "Could not parse major version",
		}
	}

	minorVersion, minErr := strconv.Atoi(versionParts[1])
	if minErr != nil {
		return nil, &UnsupportedMongoVersionError{
			version: version,
			msg:     "Could not parse minor version",
		}
	}

	patchVersion, patchErr := strconv.Atoi(versionParts[2])
	if patchErr != nil {
		return nil, &UnsupportedMongoVersionError{
			version: version,
			msg:     "Could not parse patch version",
		}
	}

	if (majorVersion < 3) || ((majorVersion == 3) && (minorVersion < 2)) {
		return nil, &UnsupportedMongoVersionError{
			version: version,
			msg:     "Only Mongo version 3.2 and above are supported",
		}
	}

	return []int{majorVersion, minorVersion, patchVersion}, nil
}

func detectPlatform() (string, error) {
	switch GoOS {
	case "darwin":
		return "osx", nil
	case "linux":
		return "linux", nil
	default:
		return "", &UnsupportedSystemError{msg: "your platform, " + GoOS + ", is not supported"}
	}
}

func detectArch() (string, error) {
	switch GoArch {
	case "amd64":
		return "x86_64", nil
	default:
		return "", &UnsupportedSystemError{msg: "your architecture, " + GoArch + ", is not supported"}
	}
}

func detectOSName(mongoVersion []int) string {
	if GoOS != "linux" {
		// Not on Linux
		return ""
	}

	osRelease, osReleaseErr := osrelease.ReadFile(EtcOsRelease)
	if osReleaseErr == nil {
		return osNameFromOsRelease(osRelease, mongoVersion)
	}

	// We control etcRedhatRelease
	//nolint:gosec
	redhatRelease, redhatReleaseErr := ioutil.ReadFile(EtcRedhatRelease)
	if redhatReleaseErr == nil {
		return osNameFromRedhatRelease(string(redhatRelease))
	}

	return ""
}

func versionGTE(a []int, b []int) bool {
	if a[0] > b[0] {
		return true
	}

	if a[0] < b[0] {
		return false
	}

	if a[1] > b[1] {
		return true
	}

	if a[1] < b[1] {
		return false
	}

	return a[2] >= b[2]
}

func osNameFromOsRelease(osRelease map[string]string, mongoVersion []int) string {
	id := osRelease["ID"]

	majorVersionString := strings.Split(osRelease["VERSION_ID"], ".")[0]
	majorVersion, err := strconv.Atoi(majorVersionString)
	if err != nil {
		return ""
	}

	switch id {
	case "ubuntu":
		if majorVersion >= 18 && versionGTE(mongoVersion, []int{4, 0, 1}) {
			return "ubuntu1804"
		}
		if majorVersion >= 16 && versionGTE(mongoVersion, []int{3, 2, 7}) {
			return "ubuntu1604"
		}
		if majorVersion >= 14 {
			return "ubuntu1404"
		}
	case "sles":
		if majorVersion >= 12 {
			return "suse12"
		}
	case "rhel":
		if majorVersion >= 7 {
			return "rhel70"
		}
	case "debian":
		if majorVersion >= 9 && versionGTE(mongoVersion, []int{3, 6, 5}) {
			return "debian92"
		}
		if majorVersion >= 8 && versionGTE(mongoVersion, []int{3, 2, 8}) {
			return "debian81"
		}
	case "amzn":
		if majorVersion == 2 && versionGTE(mongoVersion, []int{4, 0, 0}) {
			return "amazon2"
		}

		// Version before 2 has the release date, not a real version number
		return "amazon"
	}

	return ""
}

func osNameFromRedhatRelease(redhatRelease string) string {
	// RHEL 7 uses /etc/os-release, so we're just detecting RHEL 6 here
	if strings.Contains(redhatRelease, "release 6") {
		return "rhel62"
	}

	return ""
}
