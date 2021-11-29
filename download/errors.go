package download

// UnsupportedSystemError is used to indicate that we do not support
// automatic selection of the right MongoDB binary for your system
type UnsupportedSystemError struct {
	msg string
}

func (err *UnsupportedSystemError) Error() string {
	return "unsupported system: " + err.msg
}

// UnsupportedMongoVersionError is used to indicate we do not know
// how to download the given version of MongoDB
type UnsupportedMongoVersionError struct {
	version string
	msg     string
}

func (err *UnsupportedMongoVersionError) Error() string {
	return "unsupported MongoDB version \"" + err.version + "\": " + err.msg
}
