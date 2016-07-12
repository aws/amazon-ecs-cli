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

package utils

import "time"

// Sleeper interface implements the sleep() method. Useful for testing Wait* methods.
type Sleeper interface {
	Sleep(time.Duration)
}

// timeSleeper implements sleeper interface by calling the time.Sleep method for sleeping.
type TimeSleeper struct{}

func (t *TimeSleeper) Sleep(d time.Duration) {
	time.Sleep(d)
}
