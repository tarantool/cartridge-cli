package common

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func writeFile(file *os.File, content string) {
	if err := ioutil.WriteFile(file.Name(), []byte(content), 0644); err != nil {
		panic(fmt.Errorf("Failed to write file: %s", err))
	}
}

func getFileContentSinceOffset(file *os.File, offset int64) string {
	if _, err := file.Seek(offset, io.SeekStart); err != nil {
		panic(fmt.Errorf("Failed to seek: %s", err))
	}

	fileContent, err := ioutil.ReadAll(file)
	if err != nil {
		panic(fmt.Errorf("Failed to read file content: %s", err))
	}

	return string(fileContent)
}

func TestGetLastNLinesBegin(t *testing.T) {
	assert := assert.New(t)

	bufSize = 10

	var n int64
	var err error
	var longLine string

	// create tmp file
	f, err := ioutil.TempFile("", "Dockerfile")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(f.Name())

	// all lines w/o `\n` at the ent of file
	writeFile(f, "one\ntwo\nthree\nfour\nfive\nsix\nseven")
	n, err = GetLastNLinesBegin(f.Name(), 0)
	assert.Nil(err)
	assert.EqualValues(0, n)

	// all lines w/ `\n` at the ent of file
	writeFile(f, "one\ntwo\nthree\nfour\nfive\nsix\nseven\n")
	n, err = GetLastNLinesBegin(f.Name(), 0)
	assert.Nil(err)
	assert.EqualValues(0, n)

	// last 2 lines w/o `\n` at the ent of file
	writeFile(f, "one\ntwo\nthree\nfour\nfive\nsix\nseven")
	n, err = GetLastNLinesBegin(f.Name(), 2)
	assert.Nil(err)
	assert.Equal("six\nseven", getFileContentSinceOffset(f, n))

	// last 2 lines w/ `\n` at the ent of file
	writeFile(f, "one\ntwo\nthree\nfour\nfive\nsix\nseven\n")
	n, err = GetLastNLinesBegin(f.Name(), 2)
	assert.Nil(err)
	assert.Equal("six\nseven\n", getFileContentSinceOffset(f, n))

	// last 100 lines
	writeFile(f, "one\ntwo\nthree\nfour\nfive\nsix\nseven")
	n, err = GetLastNLinesBegin(f.Name(), 100)
	assert.Nil(err)
	assert.EqualValues(0, n)

	// last 2 lines w/ last line longer than buf size
	longLine = strings.Repeat("a", int(bufSize+1))
	writeFile(f, fmt.Sprintf("one\ntwo\nthree\nfour\nfive\nsix\n%s\n", longLine))
	n, err = GetLastNLinesBegin(f.Name(), 2)
	assert.Nil(err)
	assert.Equal(fmt.Sprintf("six\n%s\n", longLine), getFileContentSinceOffset(f, n))

	// last 100 lines w/ first line longer than buf size
	longLine = strings.Repeat("a", int(bufSize+1))
	writeFile(f, fmt.Sprintf("%s\ntwo\nthree\nfour\nfive\nsix\nseven\n", longLine))
	n, err = GetLastNLinesBegin(f.Name(), 0)
	assert.Nil(err)
	assert.EqualValues(0, n)

}
