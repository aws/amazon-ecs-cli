// Copyright 2015-2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//	http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package license

import (
	"io/ioutil"
	"testing"
)

func TestVendorDirectoryStructure(t *testing.T) {
	expectedDirectories := map[string]bool{"github.com": true, "golang.org": true, "gopkg.in": true, "google.golang.org": true}
	expectedFiles := map[string]bool{"modules.txt": true}
	expectedEntryCount := len(expectedDirectories) + len(expectedFiles) // for modules.txt

	files, _ := ioutil.ReadDir("./../../../../vendor")
	if len(files) != expectedEntryCount {
		t.Errorf("Should have exactly 4 directories and 1 modules.txt under vendor/. Found [%d] directories", len(files))
	}
	foundDirCount := 0
	for _, file := range files {
		if !file.IsDir() {
			if file.Name() != "modules.txt" { // we have a modules.txt from `go mod vendor`
				t.Errorf("Expected contents of vendor/ to be directories, but %s is not", file)
			}
		} else if expectedDirectories[file.Name()] {
			foundDirCount += 1
		}
	}

	if foundDirCount != len(expectedDirectories) {
		t.Errorf("Expected [%d] directories, Found=[%v] count=[%d]", expectedEntryCount, files, foundDirCount)
	}
}
