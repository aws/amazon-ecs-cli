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
		t.Fatal("Unexpected error in creating test file: ", err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(composeFileString)); err != nil {
		t.Fatal("Unexpected error writing to test file: ", err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal("Unexpected error closing test file: ", err)
	}

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
		t.Fatal("Unexpected error in creating test file: ", err1)
	}
	defer os.Remove(tmpfile1.Name())

	tmpfile2, err2 := ioutil.TempFile("", "test")
	if err2 != nil {
		t.Fatal("Unexpected error in creating test file: ", err2)
	}
	defer os.Remove(tmpfile2.Name())

	// write compose contents to files
	if _, err := tmpfile1.Write([]byte(firstFileString)); err != nil {
		t.Fatal("Unexpected error writing to test file: ", err)
	}
	if err := tmpfile1.Close(); err != nil {
		t.Fatal("Unexpected error closing test file: ", err)
	}

	if _, err := tmpfile2.Write([]byte(secondFileString)); err != nil {
		t.Fatal("Unexpected error writing to test file: ", err)
	}
	if err := tmpfile2.Close(); err != nil {
		t.Fatal("Unexpected error closing test file: ", err)
	}

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
		t.Fatal("Unexpected error in creating test file: ", err1)
	}
	defer os.Remove(tmpfile1.Name())

	tmpfile2, err2 := ioutil.TempFile("", "test")
	if err2 != nil {
		t.Fatal("Unexpected error in creating test file: ", err2)
	}
	defer os.Remove(tmpfile2.Name())

	// write compose contents to files
	if _, err := tmpfile1.Write([]byte(version3FileString)); err != nil {
		t.Fatal("Unexpected error writing to test file: ", err)
	}
	if err := tmpfile1.Close(); err != nil {
		t.Fatal("Unexpected error closing test file: ", err)
	}

	if _, err := tmpfile2.Write([]byte(version2FileString)); err != nil {
		t.Fatal("Unexpected error writing to test file: ", err)
	}
	if err := tmpfile2.Close(); err != nil {
		t.Fatal("Unexpected error closing test file: ", err)
	}

	// setup project and check that error is thrown for mismatches file versions
	project := setupTestProject(t)
	project.ecsContext.ComposeFiles = append(project.ecsContext.ComposeFiles, tmpfile1.Name())
	project.ecsContext.ComposeFiles = append(project.ecsContext.ComposeFiles, tmpfile2.Name())
	_, error := project.checkComposeVersion()

	assert.Error(t, error)
}

func TestCheckComposeVersionWhenEmpty(t *testing.T) {
	testVersion := ""
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
		t.Fatal("Unexpected error in creating test file: ", err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(composeFileString)); err != nil {
		t.Fatal("Unexpected error writing to test file: ", err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal("Unexpected error closing test file: ", err)
	}

	// setup project and parse
	project := setupTestProject(t)
	project.ecsContext.ComposeFiles = append(project.ecsContext.ComposeFiles, tmpfile.Name())
	foundVersion, _ := project.checkComposeVersion()

	assert.Equal(t, testVersion, foundVersion, "Found compose version does not match expected.")
}

func TestCheckComposeVersionWhenMissing(t *testing.T) {
	testVersion := ""
	composeFileString := `wordpress:
  image: wordpress
  ports: ["80:80"]
  mem_reservation: 500000000
mysql:
  image: mysql`

	// set up compose file
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

	// setup project and parse
	project := setupTestProject(t)
	project.ecsContext.ComposeFiles = append(project.ecsContext.ComposeFiles, tmpfile.Name())
	foundVersion, _ := project.checkComposeVersion()

	assert.Equal(t, testVersion, foundVersion, "Found compose version does not match expected.")
}

func TestThrowErrorWhenVersionInDifferentFormats(t *testing.T) {
	justMajorVersionFileString := `version: '2'
services:
  wordpress:
    image: wordpress
    ports: ["80:80"]
    mem_reservation: 500000000
  mysql:
    image: mysql`

	withMinorVersionFileString := `version: '2.0'
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
	if _, err := tmpfile1.Write([]byte(justMajorVersionFileString)); err != nil {
		t.Fatal("Unexpected error writing to test file: ", err)
	}
	if err := tmpfile1.Close(); err != nil {
		t.Fatal("Unexpected error closing test file: ", err)
	}

	if _, err := tmpfile2.Write([]byte(withMinorVersionFileString)); err != nil {
		t.Fatal("Unexpected error writing to test file: ", err)
	}
	if err := tmpfile2.Close(); err != nil {
		t.Fatal("Unexpected error closing test file: ", err)
	}

	// setup project and check that error is thrown for mismatches file versions
	project := setupTestProject(t)
	project.ecsContext.ComposeFiles = append(project.ecsContext.ComposeFiles, tmpfile1.Name())
	project.ecsContext.ComposeFiles = append(project.ecsContext.ComposeFiles, tmpfile2.Name())
	_, error := project.checkComposeVersion()

	assert.Error(t, error)
}
