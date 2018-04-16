package project

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseV3WithOneFile(t *testing.T) {
	// set up file
	numOfServices := 2
	composeFileString := `version: '3'
services:
  wordpress:
    image: wordpress
    ports: ["80:80"]
  mysql:
    image: mysql`

	tmpfile, err := ioutil.TempFile("", "test")
	if err != nil {
		t.Fatal("Unexpected error in creating test file: ", err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(composeFileString)); err != nil {
		t.Fatal("Unexpected error writing to test file: ", err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal("Unexpected error closing test file: ", err)
	}

	// add files to projects
	project := setupTestProject(t)
	project.ecsContext.ComposeFiles = append(project.ecsContext.ComposeFiles, tmpfile.Name())

	// assert number of container configs matches expected number of services
	// TODO: assert containerConfig content after conversion implemented
	conConfigResult, err := project.parseV3()
	if err != nil {
		t.Fatal("Unexpected error parsing file: ", err)
	}
	assert.Equal(t, numOfServices, len(*conConfigResult))
}

func TestParseV3WithMultileFiles(t *testing.T) {
	// set up files
	numOfServices := 3
	fileString1 := `version: '3'
services:
  wordpress:
    image: wordpress
    ports: ["80:80"]
  mysql:
    image: mysql`

	fileString2 := `version: '3'
services:
  redis:
    image: redis
    ports: ["90:90"]`

	// initialize temp files
	tmpfile1, err1 := ioutil.TempFile("", "test")
	if err1 != nil {
		t.Fatal("Unexpected error in creating test file: ", err1)
	}
	defer os.Remove(tmpfile1.Name())

	tmpfile2, err2 := ioutil.TempFile("", "test")
	if err2 != nil {
		t.Fatal("Unexpected error in creating test file: ", err2)
	}
	defer os.Remove(tmpfile2.Name())

	// write compose contents to files
	if _, err := tmpfile1.Write([]byte(fileString1)); err != nil {
		t.Fatal("Unexpected error writing to test file: ", err)
	}
	if err := tmpfile1.Close(); err != nil {
		t.Fatal("Unexpected error closing test file: ", err)
	}

	if _, err := tmpfile2.Write([]byte(fileString2)); err != nil {
		t.Fatal("Unexpected error writing to test file: ", err)
	}
	if err := tmpfile2.Close(); err != nil {
		t.Fatal("Unexpected error closing test file: ", err)
	}
	// add files to projects
	project := setupTestProject(t)
	project.ecsContext.ComposeFiles = append(project.ecsContext.ComposeFiles, tmpfile1.Name())
	project.ecsContext.ComposeFiles = append(project.ecsContext.ComposeFiles, tmpfile2.Name())

	// assert number of container configs matches expected number of services
	// TODO: assert containerConfig content after conversion implemented
	conConfigResult, err := project.parseV3()
	if err != nil {
		t.Fatal("Unexpected error parsing file: ", err)
	}
	assert.Equal(t, numOfServices, len(*conConfigResult))
}
