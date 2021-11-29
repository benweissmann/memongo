package download

import (
	"errors"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNewConfig(t *testing.T) {
	var originalGetDownloadUrl = getDownloadUrl
	var originalGetEnv = getEnv
	var originalGoOs = goOS

	Convey("Given an invalid MongoDB version", t, func() {
		Convey("Without periods", func() {
			version := "version"
			Convey("Then an error is returned", func() {
				cfg, err := NewConfig(version)
				So(cfg, ShouldBeNil)
				So(err, ShouldBeError)
				expectedError := &UnsupportedMongoVersionError{
					version: version,
					msg:     "MongoDB version number must be in the form x.y.z",
				}
				So(err, ShouldResemble, expectedError)
				So(err, ShouldHaveSameTypeAs, expectedError)
			})
		})
		Convey("With less than 2 periods", func() {
			version := "1.2"
			Convey("Then an error is returned", func() {
				cfg, err := NewConfig(version)
				So(cfg, ShouldBeNil)
				So(err, ShouldBeError)
				expectedError := &UnsupportedMongoVersionError{
					version: version,
					msg:     "MongoDB version number must be in the form x.y.z",
				}
				So(err, ShouldResemble, expectedError)
				So(err, ShouldHaveSameTypeAs, expectedError)
			})
		})
		Convey("With more than 2 periods", func() {
			version := "2.1.0.a"
			Convey("Then an error is returned", func() {
				cfg, err := NewConfig(version)
				So(cfg, ShouldBeNil)
				So(err, ShouldBeError)
				expectedError := &UnsupportedMongoVersionError{
					version: version,
					msg:     "MongoDB version number must be in the form x.y.z",
				}
				So(err, ShouldResemble, expectedError)
				So(err, ShouldHaveSameTypeAs, expectedError)
			})
		})
		Convey("With an invalid major version", func() {
			version := "a.1.0"
			Convey("Then an error is returned", func() {
				cfg, err := NewConfig(version)
				So(cfg, ShouldBeNil)
				So(err, ShouldBeError)
				expectedError := &UnsupportedMongoVersionError{
					version: version,
					msg:     "could not parse major version",
				}
				So(err, ShouldResemble, expectedError)
				So(err, ShouldHaveSameTypeAs, expectedError)
			})
		})
		Convey("With an invalid minor version", func() {
			version := "4.minor.0"
			Convey("Then an error is returned", func() {
				cfg, err := NewConfig(version)
				So(cfg, ShouldBeNil)
				So(err, ShouldBeError)
				expectedError := &UnsupportedMongoVersionError{
					version: version,
					msg:     "could not parse minor version",
				}
				So(err, ShouldResemble, expectedError)
				So(err, ShouldHaveSameTypeAs, expectedError)
			})
		})
		Convey("With an invalid patch version", func() {
			version := "4.7.pp"
			Convey("Then an error is returned", func() {
				cfg, err := NewConfig(version)
				So(cfg, ShouldBeNil)
				So(err, ShouldBeError)
				expectedError := &UnsupportedMongoVersionError{
					version: version,
					msg:     "could not parse patch version",
				}
				So(err, ShouldResemble, expectedError)
				So(err, ShouldHaveSameTypeAs, expectedError)
			})
		})
		Convey("With a non-supported old version", func() {
			version := "4.2.15"
			Convey("Then an error is returned", func() {
				cfg, err := NewConfig(version)
				So(cfg, ShouldBeNil)
				So(err, ShouldBeError)
				expectedError := &UnsupportedMongoVersionError{
					version: version,
					msg:     "only version 4.4 and above are supported",
				}
				So(err, ShouldResemble, expectedError)
				So(err, ShouldHaveSameTypeAs, expectedError)
			})
		})
	})

	Convey("Given a valid MongoDB version", t, func() {
		version := "5.0.2"

		Convey("When the download url can be found", func() {

			Convey("And the url is valid", func() {
				filename := "mongodb-linux-x86_64-ubuntu2004-" + version + ".tgz"
				mongoUrl := "https://fastdl.mongodb.org/linux/" + filename

				getDownloadUrl = func(v Version) (string, error) {
					return mongoUrl, nil
				}
				Convey("And XDG_CACHE_HOME env var is set", func() {
					getEnv = func(key string) string {
						if key == "XDG_CACHE_HOME" {
							return "/cache/home"
						}
						return ""
					}
					Convey("Then NewConfig uses the XDG_CACHE_HOME env var to determine the cache path", func() {
						cfg, err := NewConfig(version)
						So(err, ShouldBeNil)
						So(cfg.mongoVersion.String(), ShouldEqual, version)
						So(cfg.mongoUrl, ShouldEqual, mongoUrl)
						So(cfg.cachePath, ShouldEqual, "/cache/home/dp-mongodb-in-memory/"+filename+"/mongod")
					})
				})
				Convey("And XDG_CACHE_HOME env var is not set", func() {
					userHome := "/usr/home"
					getEnv = func(key string) string {
						if key == "HOME" {
							return userHome
						}
						return ""
					}
					Convey("And running on OSX", func() {
						goOS = "darwin"
						Convey("Then NewConfig determines the right home cache path", func() {
							cfg, err := NewConfig(version)
							So(err, ShouldBeNil)
							So(cfg.mongoVersion.String(), ShouldEqual, version)
							So(cfg.mongoUrl, ShouldEqual, mongoUrl)
							So(cfg.cachePath, ShouldEqual, userHome+"/Library/Caches/dp-mongodb-in-memory/"+filename+"/mongod")
						})
					})
					Convey("And running on Linux", func() {
						goOS = "linux"
						Convey("Then NewConfig determines the right home cache path", func() {
							cfg, err := NewConfig(version)
							So(err, ShouldBeNil)
							So(cfg.mongoVersion.String(), ShouldEqual, version)
							So(cfg.mongoUrl, ShouldEqual, mongoUrl)
							So(cfg.cachePath, ShouldEqual, userHome+"/.cache/dp-mongodb-in-memory/"+filename+"/mongod")
						})
					})
					Convey("And running on Windows", func() {
						goOS = "win32"
						Convey("Then NewConfig errors", func() {
							cfg, err := NewConfig(version)
							So(cfg, ShouldBeNil)
							So(err, ShouldBeError)
							expectedError := &UnsupportedSystemError{msg: "OS 'win32'"}
							So(err, ShouldResemble, expectedError)
							So(err, ShouldHaveSameTypeAs, expectedError)
						})
					})
					Reset(func() {
						goOS = originalGoOs
					})
				})
			})

			Convey("And the url is invalid", func() {
				getDownloadUrl = func(v Version) (string, error) {
					return ":invalid", nil
				}
				Convey("Then NewConfig errors", func() {
					cfg, err := NewConfig(version)
					So(err, ShouldBeError)
					So(cfg, ShouldBeNil)
				})
			})

			Reset(func() {
				getDownloadUrl = originalGetDownloadUrl
				getEnv = originalGetEnv
			})

		})

		Convey("When an error occurs while determining the download url", func() {
			expectedError := errors.New("unsupported system")
			getDownloadUrl = func(v Version) (string, error) {
				return "", expectedError
			}

			Convey("Then NewConfig errors", func() {
				cfg, err := NewConfig(version)
				So(cfg, ShouldBeNil)
				So(err, ShouldEqual, expectedError)
			})

			Reset(func() {
				getDownloadUrl = originalGetDownloadUrl
			})
		})
	})

}
