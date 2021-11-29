package download

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/spf13/afero"
)

func TestGetMongoDB(t *testing.T) {
	const (
		validMongodTarball   = "/mongodb-test.tgz"
		invalidMongodTarball = "/random.tgz"
		notTarball           = "/test"
		corrupted            = "/corrupted"
	)

	// Use a memory backed filesystem (no persistence)
	afs = afero.Afero{Fs: afero.NewMemMapFs()}

	tmpCache, _ := afs.TempDir("", "")

	Convey("Having set up a mocked server", t, func() {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			f, err := os.Open("testdata" + r.URL.Path)
			if err != nil {
				http.NotFound(w, r)
				return
			}
			defer f.Close()
			io.Copy(w, f)
		}))

		cfg := new(Config)
		cfg.cachePath = path.Join(tmpCache, "mongod")

		Convey("When the mongod exec file is not in cache", func() {
			afs.Remove(cfg.cachePath)
			Convey("And the requested url exists", func() {
				cfg.mongoUrl = ts.URL + validMongodTarball
				Convey("And the appropriate key was used to sign the package", func() {
					getMongoPublicKey = func(v Version) (afero.File, error) {
						return os.Open("testdata/key-correct.asc")
					}
					Convey("Then it downloads the tarball and stores the exec file in cache", func() {
						startTime := time.Now()

						err := GetMongoDB(*cfg)
						So(err, ShouldBeNil)

						stat, err := afs.Stat(cfg.cachePath)
						So(err, ShouldBeNil)
						So(stat.Size(), ShouldBeGreaterThan, 0)
						So(stat.Mode()&0100, ShouldNotBeZeroValue)
						So(stat.ModTime(), ShouldHappenBetween, startTime, time.Now())
					})
				})
				Convey("And the wrong key was used to sign the package", func() {
					getMongoPublicKey = func(v Version) (afero.File, error) {
						return os.Open("testdata/key-incorrect.asc")
					}
					Convey("Then an error is returned", func() {
						err := GetMongoDB(*cfg)
						So(err, ShouldBeError)
						So(err.Error(), ShouldEqual, "signature verification failed")
					})
				})
			})
			Convey("And the requested url can not be found", func() {
				cfg.mongoUrl = ts.URL + "/invalid"
				Convey("Then an error is returned", func() {
					err := GetMongoDB(*cfg)
					So(err, ShouldBeError)
					So(err.Error(), ShouldEqual, "invalid status code 404")
				})
			})
			Convey("And the requested file's checksum can not be verified", func() {
				cfg.mongoUrl = ts.URL + corrupted
				Convey("Then an error is returned", func() {
					err := GetMongoDB(*cfg)
					So(err, ShouldBeError)
					So(err.Error(), ShouldEqual, "checksum verification failed")
				})
			})
			Convey("And the requested url is not a tarball", func() {
				cfg.mongoUrl = ts.URL + notTarball
				Convey("Then an error is returned", func() {
					err := GetMongoDB(*cfg)
					So(err, ShouldBeError)
				})
			})
			Convey("And the requested url is a tarball not containing a mongod file", func() {
				cfg.mongoUrl = ts.URL + invalidMongodTarball
				Convey("Then an error is returned", func() {
					err := GetMongoDB(*cfg)
					So(err, ShouldBeError)
				})
			})
		})

		Convey("When the mongod exec file is found in cache", func() {
			afs.Create(cfg.cachePath)

			Convey("Then it uses the file in cache and it does not download it again", func() {
				cfg.mongoUrl = ts.URL + "/should-not-be-called"

				err := GetMongoDB(*cfg)
				So(err, ShouldBeNil)
			})
		})

		Reset(func() {
			ts.Close()
			afs.Remove(cfg.cachePath)
		})

	})
}
