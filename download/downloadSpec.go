package download

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"runtime"
	"strconv"
	"strings"

	"github.com/ONSdigital/log.go/v2/log"
)

const etcOsReleaseFileName = "/etc/os-release"

// We define these as package vars so we can override it in tests

var goOS = runtime.GOOS
var goArch = runtime.GOARCH

// DownloadSpec specifies what copy of MongoDB to download
type DownloadSpec struct {
	// Version is what version of MongoDB to download
	version *Version

	// Platform is "osx" or "linux"
	Platform string

	// Arch
	Arch string

	// OSName is one of:
	// - ubuntu2004
	// - ubuntu1804
	// - ubuntu1604
	// - debian10
	// - debian92
	// - "" for MacOS
	OSName string
}

// MakeDownloadSpec returns a DownloadSpec for the current operating system
func MakeDownloadSpec(version Version) (*DownloadSpec, error) {
	if !version.IsGreaterOrEqual(4, 4, 0) {
		return nil, &UnsupportedMongoVersionError{
			version: version.String(),
			msg:     "only version 4.4 and above are supported",
		}
	}

	arch, archErr := detectArch()
	if archErr != nil {
		return nil, archErr
	}

	platform, platformErr := detectPlatform()
	if platformErr != nil {
		return nil, platformErr
	}

	osName, osErr := detectLinuxId()
	if osErr != nil {
		return nil, osErr
	}

	return &DownloadSpec{
		version:  &version,
		Arch:     arch,
		Platform: platform,
		OSName:   osName,
	}, nil
}

// GetDownloadURL returns the download URL to download the binary
// from the MongoDB website
func (spec *DownloadSpec) GetDownloadURL() (string, error) {
	archiveName := "mongodb-"

	switch spec.Platform {
	case "linux":
		if spec.OSName == "" {
			return "", fmt.Errorf("invalid spec: OS name not provided")
		}
		archiveName += "linux-" + spec.Arch + "-" + spec.OSName
	case "osx":
		archiveName += "macos-" + spec.Arch
	default:
		return "", fmt.Errorf("invalid spec: unsupported platform " + spec.Platform)
	}

	return fmt.Sprintf(
		"https://fastdl.mongodb.org/%s/%s-%s.tgz",
		spec.Platform,
		archiveName,
		spec.Version(),
	), nil
}

// Version returns the MongoDb version
func (spec *DownloadSpec) Version() string {
	return spec.version.String()
}

func detectPlatform() (string, error) {
	switch goOS {
	case "darwin":
		return "osx", nil
	case "linux":
		return "linux", nil
	default:
		return "", &UnsupportedSystemError{msg: "OS " + goOS + " not supported"}
	}
}

func detectArch() (string, error) {
	switch goArch {
	case "amd64":
		return "x86_64", nil
	default:
		return "", &UnsupportedSystemError{msg: "architecture " + goArch + " not supported"}
	}
}

func detectLinuxId() (string, error) {
	if goOS != "linux" {
		// Not on Linux
		return "", nil
	}

	osreleaseFile, err := afs.Open(etcOsReleaseFileName)
	if err != nil {
		log.Error(context.Background(), "error reading "+etcOsReleaseFileName+" file", err)
		return "", err
	}
	defer osreleaseFile.Close()

	osRelease, osReleaseErr := readKeyValuePairs(osreleaseFile)
	if osReleaseErr != nil {
		return "", osReleaseErr
	}

	id := osRelease["ID"]
	versionString := strings.Split(osRelease["VERSION_ID"], ".")[0]
	version, versionErr := strconv.Atoi(versionString)
	if versionErr != nil {
		return "", &UnsupportedSystemError{msg: "invalid version number " + versionString}
	}
	switch id {
	case "ubuntu":
		if version >= 20 {
			return "ubuntu2004", nil
		}
		if version >= 18 {
			return "ubuntu1804", nil
		}
		if version >= 16 {
			return "ubuntu1604", nil
		}
		return "", &UnsupportedSystemError{msg: "invalid ubuntu version " + versionString + " (min 16)"}
	case "debian":
		if version >= 10 {
			return "debian10", nil
		}
		if version >= 9 {
			return "debian92", nil
		}
		return "", &UnsupportedSystemError{msg: "invalid debian version " + versionString + " (min 9)"}
	default:
		return "", &UnsupportedSystemError{msg: "invalid linux version '" + id + "'"}
	}
}

func readKeyValuePairs(r io.Reader) (map[string]string, error) {
	content := make(map[string]string)

	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := scanner.Text()

		if len(line) > 0 &&
			!strings.HasPrefix(line, "#") &&
			strings.Contains(line, "=") {
			// Skip empty lines, comments and malformed lines

			s := strings.Split(line, "=")
			key := strings.Trim(s[0], " ")
			key = strings.Trim(key, `"`)
			key = strings.Trim(key, `'`)

			value := strings.Trim(s[1], " ")
			value = strings.Trim(value, `"`)
			value = strings.Trim(value, `'`)
			// expand anything else that could be escaped
			value = strings.Replace(value, `\"`, `"`, -1)
			value = strings.Replace(value, `\$`, `$`, -1)
			value = strings.Replace(value, `\\`, `\`, -1)
			value = strings.Replace(value, "\\`", "`", -1)

			content[key] = value
		}
	}
	return content, scanner.Err()
}
