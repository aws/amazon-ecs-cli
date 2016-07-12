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

import "errors"

type noopCache struct{}

// NewFSCache returns a new cache backed by the filesystem. The 'name' value
// should be constant in order to access the same data between uses.
func NewNoopCache() Cache {
	return noopCache{}
}

func (self noopCache) Put(key string, val interface{}) error {
	return nil
}

func (self noopCache) Get(key string, i interface{}) error {
	return errors.New("noop cache")
}
