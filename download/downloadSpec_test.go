package download

import (
	"errors"
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/spf13/afero"
)

func TestGetDownloadURL(t *testing.T) {
	Convey("Given a DownloadSpec object", t, func() {
		spec := &DownloadSpec{
			version: &Version{
				Major: 5,
				Minor: 0,
				Patch: 2,
			},
			Arch: "x86_64",
		}
		Convey("When platform is Mac", func() {
			spec.Platform = "osx"
			Convey("Then GetDownloadURL builds the right url", func() {
				url, err := spec.GetDownloadURL()

				So(err, ShouldBeNil)
				So(url, ShouldEqual, "https://fastdl.mongodb.org/osx/mongodb-macos-x86_64-5.0.2.tgz")
			})
		})

		Convey("When platform is Linux", func() {
			spec.Platform = "linux"
			Convey("And a linux id is provided", func() {
				spec.OSName = "ubuntu2004"
				Convey("Then GetDownloadURL builds the right url", func() {
					url, err := spec.GetDownloadURL()

					So(err, ShouldBeNil)
					So(url, ShouldEqual, "https://fastdl.mongodb.org/linux/mongodb-linux-x86_64-ubuntu2004-5.0.2.tgz")
				})
			})
			Convey("And no linux id is provided", func() {
				spec.OSName = ""
				Convey("Then an error is thrown", func() {
					url, err := spec.GetDownloadURL()

					So(url, ShouldBeBlank)
					So(err, ShouldBeError)
					So(err, ShouldResemble, errors.New("invalid spec: OS name not provided"))
				})
			})
		})

		Convey("When platform is not supported", func() {
			spec.Platform = "win32"

			Convey("Then an error is thrown", func() {
				url, err := spec.GetDownloadURL()

				So(url, ShouldBeBlank)
				So(err, ShouldBeError)
				So(err, ShouldResemble, errors.New("invalid spec: unsupported platform win32"))
			})
		})
	})
}

func TestMakeDownloadSpec(t *testing.T) {
	var originalGoOs = goOS
	var originalGoArch = goArch

	// Use a memory backed filesystem (no persistence)
	afs = afero.Afero{Fs: afero.NewMemMapFs()}

	Convey("Given a valid MongoDB version", t, func() {
		version := Version{
			Major: 5,
			Minor: 0,
			Patch: 4,
		}

		Convey("When running on a x64 architecture", func() {
			goArch = "amd64"
			Convey("And on Mac", func() {
				goOS = "darwin"
				Convey("Then the returned spec is correct", func() {
					spec, err := MakeDownloadSpec(version)

					So(err, ShouldBeNil)
					So(spec, ShouldResemble, &DownloadSpec{
						version: &Version{
							Major: 5,
							Minor: 0,
							Patch: 4,
						},
						Arch:     "x86_64",
						Platform: "osx",
					})
				})
			})

			Convey("And on Linux", func() {
				goOS = "linux"

				tests := map[string]struct {
					linuxId      string
					linuxVersion string
					expectedSpec *DownloadSpec
					expectedErr  error
				}{
					"Ubuntu 20.04": {
						linuxId:      "ubuntu",
						linuxVersion: "20.04",
						expectedSpec: &DownloadSpec{
							version:  &version,
							Arch:     "x86_64",
							Platform: "linux",
							OSName:   "ubuntu2004",
						},
					},
					"Ubuntu 20.10": {
						linuxId:      "ubuntu",
						linuxVersion: "20.10",
						expectedSpec: &DownloadSpec{
							version:  &version,
							Arch:     "x86_64",
							Platform: "linux",
							OSName:   "ubuntu2004",
						},
					},
					"Ubuntu 18.04": {
						linuxId:      "ubuntu",
						linuxVersion: "18.04",
						expectedSpec: &DownloadSpec{
							version:  &version,
							Arch:     "x86_64",
							Platform: "linux",
							OSName:   "ubuntu1804",
						},
					},
					"Ubuntu 16.04": {
						linuxId:      "ubuntu",
						linuxVersion: "16.04",
						expectedSpec: &DownloadSpec{
							version:  &version,
							Arch:     "x86_64",
							Platform: "linux",
							OSName:   "ubuntu1604",
						},
					},
					"Old Ubuntu": {
						linuxId:      "ubuntu",
						linuxVersion: "14.04",
						expectedErr:  &UnsupportedSystemError{msg: "invalid ubuntu version 14 (min 16)"},
					},
					"Debian 10": {
						linuxId:      "debian",
						linuxVersion: "10",
						expectedSpec: &DownloadSpec{
							version:  &version,
							Arch:     "x86_64",
							Platform: "linux",
							OSName:   "debian10",
						},
					},
					"Debian 9.2": {
						linuxId:      "debian",
						linuxVersion: "9.2",
						expectedSpec: &DownloadSpec{
							version:  &version,
							Arch:     "x86_64",
							Platform: "linux",
							OSName:   "debian92",
						},
					},
					"Old Debian": {
						linuxId:      "debian",
						linuxVersion: "8.1",
						expectedErr:  &UnsupportedSystemError{msg: "invalid debian version 8 (min 9)"},
					},
					"Other Linux": {
						linuxId:      "fedora",
						linuxVersion: "17",
						expectedErr:  &UnsupportedSystemError{msg: "invalid linux version 'fedora'"},
					},
					"Invalid linux version": {
						linuxId:      "fedora",
						linuxVersion: "vvv111",
						expectedErr:  &UnsupportedSystemError{msg: "invalid version number vvv111"},
					},
				}
				for name, tc := range tests {
					Convey(name, func() {
						osrelease := fmt.Sprintf("ID=%s\nVERSION_ID=%s\n", tc.linuxId, tc.linuxVersion)
						// We are using a memory backed file system
						// and this will not affect a real file if it existed
						afs.WriteFile(etcOsReleaseFileName, []byte(osrelease), 0744)
						Convey("Then the returned spec is correct", func() {
							spec, err := MakeDownloadSpec(version)
							if tc.expectedErr != nil {
								So(err, ShouldBeError)
								So(err, ShouldResemble, tc.expectedErr)
								So(err, ShouldHaveSameTypeAs, tc.expectedErr)
							} else {
								So(err, ShouldBeNil)
							}

							if tc.expectedSpec != nil {
								So(spec, ShouldResemble, tc.expectedSpec)
							} else {
								So(spec, ShouldBeNil)
							}
						})
					})
				}

				Convey("When there is an error reading the os-release file", func() {
					// We are using a memory backed file system
					// and this will not affect a real file if it existed
					afs.Remove(etcOsReleaseFileName)

					Convey("Then an error is returned", func() {
						spec, err := MakeDownloadSpec(version)
						So(err, ShouldBeError)
						So(spec, ShouldBeNil)
					})
				})
			})

			Convey("And on a non supported platform", func() {
				goOS = "win32"
				Convey("Then an error is returned", func() {
					spec, err := MakeDownloadSpec(version)
					So(spec, ShouldBeNil)
					So(err, ShouldBeError)
					expectedError := &UnsupportedSystemError{msg: "OS " + goOS + " not supported"}
					So(err, ShouldResemble, expectedError)
					So(err, ShouldHaveSameTypeAs, expectedError)
				})
			})

			Reset(func() {
				goOS = originalGoOs
			})
		})

		Convey("When running on a non supported architecture", func() {
			goArch = "386"
			Convey("Then an error is returned", func() {
				spec, err := MakeDownloadSpec(version)
				So(spec, ShouldBeNil)
				So(err, ShouldBeError)
				expectedError := &UnsupportedSystemError{msg: "architecture " + goArch + " not supported"}
				So(err, ShouldResemble, expectedError)
				So(err, ShouldHaveSameTypeAs, expectedError)
			})
		})

		Reset(func() {
			goArch = originalGoArch
		})
	})
}
