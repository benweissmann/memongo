// Copyright 2021 Tryvium Travels LTD
// Copyright 2019-2020 Ben Weissmann
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
