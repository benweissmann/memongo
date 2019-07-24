package mongobin

// UnsupportedSystemError is used to indicate that memongo does not support
// automatic selection of the right MongoDB binary for your system
type UnsupportedSystemError struct {
	msg string
}

func (err *UnsupportedSystemError) Error() string {
	return "memongo does not support automatic downloading on your system: " + err.msg
}

// UnsupportedMongoVersionError is used to indicate the memongo doesn't know
// how to download the given version of MongoDB
type UnsupportedMongoVersionError struct {
	version string
	msg     string
}

func (err *UnsupportedMongoVersionError) Error() string {
	return "memongo does not support MongoDB version \"" + err.version + "\": " + err.msg
}
