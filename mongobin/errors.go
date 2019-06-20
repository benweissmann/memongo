package mongobin

type UnsupportedSystemError struct {
	msg string
}

func (err *UnsupportedSystemError) Error() string {
	return "memongo does not support automatic downloading on your system: " + err.msg
}

type UnsupportedMongoVersionError struct {
	version string
	msg     string
}

func (err *UnsupportedMongoVersionError) Error() string {
	return "memongo does not support MongoDB version \"" + err.version + "\": " + err.msg
}
