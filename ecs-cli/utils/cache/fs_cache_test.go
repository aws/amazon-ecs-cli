// Copyright 2015-2016 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package cache

import (
	"os"
	"testing"
)

func TestCacheCreatesDir(t *testing.T) {
	created := make(chan string)
	osMkdirAll = func(path string, perms os.FileMode) error {
		if perms != 0700 {
			t.Errorf("directory created with more open perms than expected; %v", perms)
		}
		go func() { created <- path }()
		return nil
	}
	NewFSCache("foo")
	dir := <-created

	expected, _ := cacheDir("foo")
	if dir != expected {
		t.Errorf("expected %v, got %v", expected, dir)
	}
}
