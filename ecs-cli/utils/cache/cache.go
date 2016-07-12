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

// Package cache provides a simple interface for a key-value cache. It also
// provides a default implementation that stores files on disk in the user's
// '.cache' directory.
// The default implementation should only be used within this program since it
// hardcodes program-specific information
package cache

type Cache interface {
	Put(key string, value interface{}) error
	Get(key string, i interface{}) error
}
