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
		archiveName += "osx-ssl-" + spec.Arch + "-" + spec.Version + ".tgz"
	}

	return fmt.Sprintf(
		"https://fastdl.mongodb.org/%s/%s",
		spec.Platform,
		archiveName,
	)
}
