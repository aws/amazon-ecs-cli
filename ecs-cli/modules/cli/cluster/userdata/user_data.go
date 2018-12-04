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
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/mail"
	"net/textproto"
	"strings"
)

// UserDataBuilder contains functionality to create user data scripts for Container Instances
type UserDataBuilder interface {
	AddFile(fileName string) error
	Build() (string, error)
}

// Builder implements UserDataBuilder
type Builder struct {
	writer      *multipart.Writer
	clusterName string
	userdata    *bytes.Buffer
}

// NewBuilder creates a Builder object for a given clusterName
func NewBuilder(clusterName string) UserDataBuilder {
	buf := new(bytes.Buffer)
	writer := multipart.NewWriter(buf)

	builder := &Builder{
		writer:      writer,
		clusterName: clusterName,
		userdata:    buf,
	}

	return builder
}

// AddFile adds new userdata from a file
func (b *Builder) AddFile(fileName string) error {
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return err
	}
	extraUserData := string(data)

	if ok, headers, body := isMultipart(extraUserData); ok { // extraUserData is multipart
		if err = b.processExistingMultipart(headers, body); err != nil {
			return err
		}
	} else { // extraUserData is not already multipart
		if err = b.writeExtraUserDataMimePart(extraUserData); err != nil {
			return err
		}
	}
	return nil
}

// Build the userdata for the given cluster
// Build() is not idempotent and can only be called once
func (b *Builder) Build() (string, error) {
	// add user data for joining the ECS Cluster
	if err := b.writeClusterUserDataMimePart(); err != nil {
		return "", err
	}
	if err := b.writer.Close(); err != nil {
		return "", err
	}
	header := fmt.Sprintf("Content-Type: multipart/mixed; boundary=\"%s\"\nMIME-Version: 1.0\n\n", b.writer.Boundary())
	archive := append([]byte(header), b.userdata.Bytes()...)
	return unixifyLineEndings(string(archive)), nil
}

func (b *Builder) writePart(header textproto.MIMEHeader, body []byte) error {
	newPart, err := b.writer.CreatePart(header)
	if err != nil {
		return err
	}
	if _, err = newPart.Write(body); err != nil {
		return err
	}
	return nil
}

// unpacks an existing multipart archive and writes it using `writer`
func (b *Builder) processExistingMultipart(headers map[string]string, body io.Reader) error {
	partReader := multipart.NewReader(body, headers["boundary"])
	for {
		part, err := partReader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		partBody, err := ioutil.ReadAll(part)
		if err != nil {
			return err
		}
		if err = b.writePart(part.Header, partBody); err != nil {
			return err
		}
		if err = part.Close(); err != nil {
			return err
		}
	}
	return nil
}

// Determines if the given string is already a multipart archive
// If it is, then it returns true, the multipart archive headers, and an io.Reader
// which can read the body of the multipart archive
func isMultipart(data string) (bool, map[string]string, io.Reader) {
	msg, err := mail.ReadMessage(strings.NewReader(data))
	if err != nil {
		return false, nil, nil
	}

	mediaType, headers, err := mime.ParseMediaType(msg.Header.Get("Content-Type"))
	if err != nil {
		return false, nil, nil
	}

	if strings.HasPrefix(mediaType, "multipart/") {
		return true, headers, msg.Body
	}
	return false, nil, nil
}

func (b *Builder) getClusterUserData() string {
	joinClusterUserData := `
#!/bin/bash
echo ECS_CLUSTER=%s >> /etc/ecs/ecs.config
`
	return fmt.Sprintf(joinClusterUserData, b.clusterName)
}

// writes the user data script to join the ecs cluster to a multipart archive
func (b *Builder) writeClusterUserDataMimePart() error {
	header := make(textproto.MIMEHeader)
	header.Add("Content-Type", "text/text/x-shellscript; charset=\"utf-8\"")
	header.Add("MIME-Version", "1.0")

	return b.writePart(header, []byte(b.getClusterUserData()))
}

// takes user inputted user data and writes it as one part in the mime multipart archive
// `extraUserData` is any user data passed in by the user which is not already a multipart archive
func (b *Builder) writeExtraUserDataMimePart(extraUserData string) error {
	header := make(textproto.MIMEHeader)
	// Setting the content type as text/plain is safe because Cloud Init will read its contents to determine its type
	header.Add("Content-Type", "text/text/plain; charset=\"utf-8\"")
	header.Add("MIME-Version", "1.0")

	return b.writePart(header, []byte(extraUserData))
}

// replaces all "\r\n" with "\n"
func unixifyLineEndings(s string) string {
	return strings.Replace(s, "\r\n", "\n", -1)
}
