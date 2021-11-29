package download

import (
	"context"
	"net/url"
	"os"
	"path"

	"github.com/ONSdigital/log.go/v2/log"
)

//folderName is the name of the folder we will be saving mongod in the cache path
const folderName = "dp-mongodb-in-memory"

// getDownloadUrl returns the mongodb download url for a given version
var getDownloadUrl = func(v Version) (string, error) {
	spec, err := MakeDownloadSpec(v)
	if err != nil {
		return "", err
	}

	return spec.GetDownloadURL()
}

// getEnv returns the value of an environment variable
var getEnv = func(key string) string {
	return os.Getenv(key)
}

// Config keeps the configuration values for downloading and storing the Mongo binary
type Config struct {
	// The MongoDB version we are using
	mongoVersion Version
	// The URL where the required mongodb tarball can be downloaded from
	mongoUrl string
	// The path where the mongod executable can be found if previously downloaded
	cachePath string
}

// NewConfig creates the config values for the given version.
// It will identify the appropriate mongodb artifact
// and the cache path based on the current OS
func NewConfig(mongoVersionStr string) (*Config, error) {
	version, versionErr := NewVersion(mongoVersionStr)
	if versionErr != nil {
		return nil, versionErr
	}

	downloadUrl, err := getDownloadUrl(*version)
	if err != nil {
		return nil, err
	}

	cachePath, err := buildBinCachePath(downloadUrl)
	if err != nil {
		return nil, err
	}

	return &Config{
		mongoVersion: *version,
		mongoUrl:     downloadUrl,
		cachePath:    cachePath,
	}, nil
}

// buildBinCachePath returns the full path to where the mongod binary should be located.
func buildBinCachePath(downloadUrl string) (string, error) {
	cacheHome, err := defaultBaseCachePath()
	if err != nil {
		log.Error(context.Background(), "cache directory not found", err)
		return "", err
	}

	urlParsed, err := url.Parse(downloadUrl)
	if err != nil {
		log.Error(context.Background(), "error parsing url", err, log.Data{"url": downloadUrl})
		return "", err
	}

	dirname := path.Base(urlParsed.Path)

	return path.Join(cacheHome, folderName, dirname, "mongod"), nil
}

// defaultBaseCachePath finds the OS cache path.
//
// Returns the value of XDG_CACHE_HOME environment variable if any.
// Otherwise it uses the default Mac or Linux home caches
func defaultBaseCachePath() (string, error) {
	var cacheHome = getEnv("XDG_CACHE_HOME")

	if cacheHome == "" {
		switch goOS {
		case "darwin":
			cacheHome = path.Join(getEnv("HOME"), "Library", "Caches")
		case "linux":
			cacheHome = path.Join(getEnv("HOME"), ".cache")
		default:
			return "", &UnsupportedSystemError{msg: "OS '" + goOS + "'"}
		}
	}
	return cacheHome, nil
}

// MongoPath returns the path to the mongod executable file
func (cfg *Config) MongoPath() string {
	return cfg.cachePath
}

// mongoSignatureUrl returns the url for the public signature file.
func (cfg *Config) mongoSignatureUrl() string {
	return cfg.mongoUrl + ".sig"
}

// mongoChecksumUrl returns the url for the SHA256 file
func (cfg *Config) mongoChecksumUrl() string {
	return cfg.mongoUrl + ".sha256"
}
