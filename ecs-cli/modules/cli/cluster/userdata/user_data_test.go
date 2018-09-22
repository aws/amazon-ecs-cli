// Copyright 2018 Amazon.com, Inc. or its affiliates. All Rights Reserved.
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

package userdata

import (
	"bytes"
	"io/ioutil"
	"mime/multipart"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testClusterName = "cluster"
	testBoundary    = "========multipart-boundary=="
)

func newBuilderInTest(buf *bytes.Buffer, writer *multipart.Writer) *Builder {
	builder := &Builder{
		writer:      writer,
		clusterName: testClusterName,
		userdata:    buf,
	}

	return builder
}

var existingMultipartArchive = `Content-Type: multipart/mixed; boundary="b19714718818a0e648a502570ed78486f7358d7d3a4d42c3716e81102b56"
MIME-Version: 1.0

--b19714718818a0e648a502570ed78486f7358d7d3a4d42c3716e81102b56
Content-Type: text/text/plain; charset="utf-8"
Mime-Version: 1.0

#!/bin/bash
echo "Clyde spends the next year lonely and confused" >> $HOME/chapter3
echo "He knows Pudding wasn't right for him, but he doesn't know how to move on" >> $HOME/chapter3
echo "Meanwhile, Pudding and Tum Tum have the honeymoon of a lifetime" >> $HOME/chapter3
echo "They spend 3 months travelling the world, visiting DogseyLand, the Temple of Doge in Japan, and much more" >> $HOME/chapter3
echo "Tum Tum experiments with 10 different haircuts/stylings" >> $HOME/chapter3

--b19714718818a0e648a502570ed78486f7358d7d3a4d42c3716e81102b56
Content-Type: text/text/x-shellscript; charset="utf-8"
Mime-Version: 1.0


#!/bin/bash
aws s3 cp s3://my-scripts/setupenv $HOME/setupenv
source $HOME/setupenv


--b19714718818a0e648a502570ed78486f7358d7d3a4d42c3716e81102b56--`

var extraUserDataShellScript = `#!/bin/bash
echo "Quickly, the honeymoon bliss wears off" >> $HOME/chapter4
echo "The once happy couple, now fights over small things, like who gets to chew which toy" >> $HOME/chapter4
echo "Pudding drags her bed to away from Tum Tum's, and hides a pile of toys underneath it" >> $HOME/chapter4`

var extraUserDataCloudConfig = `#cloud-config

# Install additional packages on first boot
packages:
 - pwgen
 - pastebinit
 - [libpython2.7, 2.7.3-0ubuntu3.1]
 - gpg`

var expectedMimeMultipart = `Content-Type: multipart/mixed; boundary="========multipart-boundary=="
MIME-Version: 1.0

--========multipart-boundary==
Content-Type: text/text/plain; charset="utf-8"
Mime-Version: 1.0

#!/bin/bash
echo "Clyde spends the next year lonely and confused" >> $HOME/chapter3
echo "He knows Pudding wasn't right for him, but he doesn't know how to move on" >> $HOME/chapter3
echo "Meanwhile, Pudding and Tum Tum have the honeymoon of a lifetime" >> $HOME/chapter3
echo "They spend 3 months travelling the world, visiting DogseyLand, the Temple of Doge in Japan, and much more" >> $HOME/chapter3
echo "Tum Tum experiments with 10 different haircuts/stylings" >> $HOME/chapter3

--========multipart-boundary==
Content-Type: text/text/x-shellscript; charset="utf-8"
Mime-Version: 1.0


#!/bin/bash
aws s3 cp s3://my-scripts/setupenv $HOME/setupenv
source $HOME/setupenv


--========multipart-boundary==
Content-Type: text/text/plain; charset="utf-8"
Mime-Version: 1.0

#!/bin/bash
echo "Quickly, the honeymoon bliss wears off" >> $HOME/chapter4
echo "The once happy couple, now fights over small things, like who gets to chew which toy" >> $HOME/chapter4
echo "Pudding drags her bed to away from Tum Tum's, and hides a pile of toys underneath it" >> $HOME/chapter4
--========multipart-boundary==
Content-Type: text/text/plain; charset="utf-8"
Mime-Version: 1.0

#cloud-config

# Install additional packages on first boot
packages:
 - pwgen
 - pastebinit
 - [libpython2.7, 2.7.3-0ubuntu3.1]
 - gpg
--========multipart-boundary==
Content-Type: text/text/x-shellscript; charset="utf-8"
Mime-Version: 1.0


#!/bin/bash
echo ECS_CLUSTER=cluster >> /etc/ecs/ecs.config

--========multipart-boundary==--
`

func TestBuildUserDataWithExtraData(t *testing.T) {

	multipartFilePath := writeTempFile(t, "existingMultipartArchive", existingMultipartArchive)
	defer os.Remove(multipartFilePath)

	shellScriptFilePath := writeTempFile(t, "extraUserDataShellScript", extraUserDataShellScript)
	defer os.Remove(shellScriptFilePath)

	cloudConfigFilePath := writeTempFile(t, "extraUserDataCloudConfig", extraUserDataCloudConfig)
	defer os.Remove(cloudConfigFilePath)

	buf := new(bytes.Buffer)
	writer := multipart.NewWriter(buf)
	// set the boundary between parts so that output is deterministic
	writer.SetBoundary(testBoundary)
	builder := newBuilderInTest(buf, writer)

	err := builder.AddFile(multipartFilePath)
	assert.NoError(t, err, "Unexpected error calling AddFile()")
	err = builder.AddFile(shellScriptFilePath)
	assert.NoError(t, err, "Unexpected error calling AddFile()")
	err = builder.AddFile(cloudConfigFilePath)
	assert.NoError(t, err, "Unexpected error calling AddFile()")

	actual, err := builder.Build()
	assert.NoError(t, err, "Unexpected error calling Build()")
	expected := unixifyLineEndings(expectedMimeMultipart)
	assert.Equal(t, expected, actual, "Expected resulting mime multipart archive to match")
}

func TestBuildUserDataNoExtraData(t *testing.T) {
	var expectedUserData = `Content-Type: multipart/mixed; boundary="========multipart-boundary=="
MIME-Version: 1.0

--========multipart-boundary==
Content-Type: text/text/x-shellscript; charset="utf-8"
Mime-Version: 1.0


#!/bin/bash
echo ECS_CLUSTER=cluster >> /etc/ecs/ecs.config

--========multipart-boundary==--
`

	buf := new(bytes.Buffer)
	writer := multipart.NewWriter(buf)
	// set the boundary between parts so that output is deterministic
	writer.SetBoundary(testBoundary)
	builder := newBuilderInTest(buf, writer)

	actual, err := builder.Build()
	assert.NoError(t, err, "Unexpected error calling Build()")
	expected := unixifyLineEndings(expectedUserData)
	assert.Equal(t, expected, actual, "Expected resulting mime multipart archive to match")
}

func writeTempFile(t *testing.T, name, content string) string {
	tmpfile, err := ioutil.TempFile("", name)
	assert.NoError(t, err, "Could not create tempfile")

	_, err = tmpfile.Write([]byte(content))
	assert.NoError(t, err, "Could not write data to ecs-params tempfile")

	err = tmpfile.Close()
	assert.NoError(t, err, "Could not close tempfile")

	return tmpfile.Name()
}
