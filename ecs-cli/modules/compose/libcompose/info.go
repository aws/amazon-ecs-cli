// This file is derived from Docker's Libcompose project, Copyright 2015 Docker, Inc.
// The original code may be found :
// https://github.com/docker/libcompose/blob/master/project/info.go
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Modifications are Copyright 2015 Amazon.com, Inc. or its affiliates. Licensed under the Apache License 2.0
package libcompose

import (
	"bytes"
	"io"
	"text/tabwriter"
)

func (infos InfoSet) String() string {
	//no error checking, none of this should fail
	buffer := bytes.NewBuffer(make([]byte, 0, 1024))
	tabwriter := tabwriter.NewWriter(buffer, 4, 4, 2, ' ', 0)

	first := true
	for _, info := range infos {
		if first {
			writeLine(tabwriter, true, info)
		}
		first = false
		writeLine(tabwriter, false, info)
	}

	tabwriter.Flush()
	return buffer.String()
}

func writeLine(writer io.Writer, key bool, info Info) {
	first := true
	for _, part := range info {
		if !first {
			writer.Write([]byte{'\t'})
		}
		first = false
		if key {
			writer.Write([]byte(part.Key))
		} else {
			writer.Write([]byte(part.Value))
		}
	}

	writer.Write([]byte{'\n'})
}
