// Copyright 2015-2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.
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

package logger

import "github.com/awslabs/amazon-ecr-credential-helper/ecr-login/config"

func SetupLogger() {
	config.SetupLoggerWithConfig(loggerConfig())
}

func loggerConfig() string {
	return `
	<seelog type="asyncloop" minlevel="info">
		<outputs>
			<console formatid="colored"/>
		</outputs>
		<formats>
			<format id="colored" format="%EscM(34)%LEVEL%EscM(39) %Msg%n%EscM(0)" />
		</formats>
	</seelog>
`
}
