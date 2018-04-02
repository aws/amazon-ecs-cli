package project

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckComposeVersionForOneFile(t *testing.T) {
	testVersion := "3"
	composeFileString := `version: '` + testVersion + `'
services:
  wordpress:
    image: wordpress
    ports: ["80:80"]
    mem_reservation: 500000000
  mysql:
    image: mysql`

	// set up compose file
	tmpfile, err := ioutil.TempFile("", "test")
	if err != nil {
		t.Fatal("Unexpected error in creating test file", err)
	}
	defer os.Remove(tmpfile.Name())

	tmpfile.Write([]byte(composeFileString))
	tmpfile.Close()

	// setup project and parse
	project := setupTestProject(t)
	project.ecsContext.ComposeFiles = append(project.ecsContext.ComposeFiles, tmpfile.Name())
	foundVersion, _ := project.checkComposeVersion()

	assert.Equal(t, testVersion, foundVersion, "Found compose version does not match expected.")
}

func TestCheckComposeVersionForMultipleFiles(t *testing.T) {
	testVersion := "3"
	firstFileString := `version: '` + testVersion + `'
services:
  wordpress:
    image: wordpress
    ports: ["80:80"]
    mem_reservation: 500000000
  mysql:
    image: mysql`

	secondFileString := `version: '` + testVersion + `'
services:
  redis:
    image: redis
    ports: ["90:90"]`

	// initialize temp files
	tmpfile1, err1 := ioutil.TempFile("", "test")
	if err1 != nil {
		t.Fatal("Unexpected error in creating test file", err1)
	}
	defer os.Remove(tmpfile1.Name())

	tmpfile2, err2 := ioutil.TempFile("", "test")
	if err2 != nil {
		t.Fatal("Unexpected error in creating test file", err2)
	}
	defer os.Remove(tmpfile2.Name())

	// write compose contents to file
	tmpfile1.Write([]byte(firstFileString))
	tmpfile1.Close()

	tmpfile2.Write([]byte(secondFileString))
	tmpfile2.Close()

	// setup project and check file version(s)
	project := setupTestProject(t)
	project.ecsContext.ComposeFiles = append(project.ecsContext.ComposeFiles, tmpfile1.Name())
	project.ecsContext.ComposeFiles = append(project.ecsContext.ComposeFiles, tmpfile2.Name())
	foundVersion, _ := project.checkComposeVersion()

	assert.Equal(t, testVersion, foundVersion, "Found compose version does not match expected.")
}

func TestThrowErrorWhenComposeVersionsConflict(t *testing.T) {
	version3FileString := `version: '3'
services:
  wordpress:
    image: wordpress
    ports: ["80:80"]
    mem_reservation: 500000000
  mysql:
    image: mysql`

	version2FileString := `version: '2'
services:
  redis:
    image: redis
    ports: ["90:90"]`

	// initialize temp files
	tmpfile1, err1 := ioutil.TempFile("", "test")
	if err1 != nil {
		t.Fatal("Unexpected error in creating test file", err1)
	}
	defer os.Remove(tmpfile1.Name())

	tmpfile2, err2 := ioutil.TempFile("", "test")
	if err2 != nil {
		t.Fatal("Unexpected error in creating test file", err2)
	}
	defer os.Remove(tmpfile2.Name())

	// write compose contents to file
	tmpfile1.Write([]byte(version3FileString))
	tmpfile1.Close()
	tmpfile2.Write([]byte(version2FileString))
	tmpfile2.Close()

	// setup project and check that error is thrown for mismatches file versions
	project := setupTestProject(t)
	project.ecsContext.ComposeFiles = append(project.ecsContext.ComposeFiles, tmpfile1.Name())
	project.ecsContext.ComposeFiles = append(project.ecsContext.ComposeFiles, tmpfile2.Name())
	_, error := project.checkComposeVersion()

	assert.Error(t, error)
}
