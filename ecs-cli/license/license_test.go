// Copyright 2015-2016 Amazon.com, Inc. or its affiliates. All Rights Reserved.
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
	directories, _ := ioutil.ReadDir("./../vendor")
	if len(directories) != 2 {
		t.Errorf("Should have exactly 2 directories under vendor/. Found [%s] directories", len(directories))
	}
	var found bool
	for _, dir := range directories {
		if !dir.IsDir() {
			t.Errorf("Expected contents of vendor/ to be directories, but %s is not", dir)
		}
		if "github.com" == dir.Name() {
			found = true
		}
	}

	if !found {
		t.Errorf("Expected one directory to be name=github.com, Found=[%v]", directories)
	}
}
