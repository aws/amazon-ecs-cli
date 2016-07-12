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

package utils

import (
	"errors"

	log "github.com/Sirupsen/logrus"
)

var ErrUnsupported error = errors.New("UnsupportedOperation")

// LogError logs the error with the given message at ERROR level
func LogError(err error, msg string) {
	logErrorWithFields(err, msg, nil)
}

// logErrorWithFields logs the error with the given message and custom values at ERROR level
func logErrorWithFields(err error, msg string, additionalFields log.Fields) {
	fields := log.Fields{}
	if err != nil {
		fields["error"] = err
	}
	for key, value := range additionalFields {
		fields[key] = value
	}
	log.WithFields(fields).Error(msg)
}
