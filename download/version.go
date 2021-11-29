package download

import (
	"fmt"
	"strconv"
	"strings"
)

// Version represents a version (Major.Minor.Patch)
type Version struct {
	Major int
	Minor int
	Patch int
}

// NewVersion parses the version string and creates a Version object
func NewVersion(version string) (*Version, error) {
	versionParts := strings.Split(version, ".")
	if len(versionParts) != 3 {
		return nil, &UnsupportedMongoVersionError{
			version: version,
			msg:     "MongoDB version number must be in the form x.y.z",
		}
	}

	majorVersion, majErr := strconv.Atoi(versionParts[0])
	if majErr != nil {
		return nil, &UnsupportedMongoVersionError{
			version: version,
			msg:     "could not parse major version",
		}
	}

	minorVersion, minErr := strconv.Atoi(versionParts[1])
	if minErr != nil {
		return nil, &UnsupportedMongoVersionError{
			version: version,
			msg:     "could not parse minor version",
		}
	}

	patchVersion, patchErr := strconv.Atoi(versionParts[2])
	if patchErr != nil {
		return nil, &UnsupportedMongoVersionError{
			version: version,
			msg:     "could not parse patch version",
		}
	}

	return &Version{
		Major: majorVersion,
		Minor: minorVersion,
		Patch: patchVersion,
	}, nil
}

// String returns the string representation of a Version:
// Major.Minor.Patch
func (v *Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// IsGreaterOrEqual checks if the version is greater or equal than another version
func (v *Version) IsGreaterOrEqual(major int, minor int, patch int) bool {
	if v.Major > major {
		return true
	}
	if v.Major < major {
		return false
	}
	if v.Minor > minor {
		return true
	}
	if v.Minor < minor {
		return false
	}
	return v.Patch >= patch
}
