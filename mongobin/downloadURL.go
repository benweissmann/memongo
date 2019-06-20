package mongobin

import "fmt"



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
