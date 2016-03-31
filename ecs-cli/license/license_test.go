// Copyright 2015 Amazon.com, Inc. or its affiliates. All Rights Reserved.
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
	directories, _ := ioutil.ReadDir("./../vendor/src")
	if len(directories) != 1 {
		t.Errorf("Should have exactly 1 directory under vendor/src. Found [%s] directories", len(directories))
	}
	dir := directories[0]
	if !dir.IsDir() {
		t.Error("Expected only contents of vendor/src to be a directory")
	}
	if "github.com" != dir.Name() {
		t.Errorf("directory name : Expected=github.com, Found=[%s]", dir.Name())
	}
}
