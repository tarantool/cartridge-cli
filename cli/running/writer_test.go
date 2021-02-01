package running

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func getLogsBytes(logs []string) []byte {
	return []byte(strings.Join(logs, "\n"))
}

func getLogsNoPrefix(logs []string) string {
	resLines := make([]string, len(logs))
	for i := range logs {
		resLines[i] = fmt.Sprintf("%s", logs[i])
	}

	return strings.Join(resLines, "\n")
}

func getLogsWithPrefix(prefix string, logs []string) string {
	resLines := make([]string, len(logs))
	for i := range logs {
		resLines[i] = fmt.Sprintf("%s | %s", prefix, logs[i])
	}

	return strings.Join(resLines, "\n")
}

func TestWrite(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)

	out := bytes.NewBuffer(nil)
	var logs []string
	var logBytes []byte
	var n int
	var err error

	id := "my-id"

	writer := newColorizedWriter(id)
	writer.out = out

	// multiline
	out.Reset()
	logs = []string{
		"Some long",
		"multiline",
		"log",
	}
	logBytes = getLogsBytes(logs)
	n, err = writer.Write(logBytes)
	assert.Nil(err)
	assert.Equal(len(logBytes), n)
	assert.Equal(getLogsWithPrefix(id, logs), out.String())

	// one line (w/o \n)
	out.Reset()
	logs = []string{
		"Some one-line log line",
	}
	logBytes = getLogsBytes(logs)
	n, err = writer.Write(logBytes)
	assert.Nil(err)
	assert.Equal(len(logBytes), n)
	assert.Equal(getLogsWithPrefix(id, logs), out.String())
}

func TestNoPrefixWrite(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)
	out := bytes.NewBuffer(nil)

	writer := newDummyWriter()
	writer.out = out

	out.Reset()
	logs := []string{
		"Line_01",
		"Multiline02!",
		"Log.Line03",
	}

	logBytes := getLogsBytes(logs)
	n, err := writer.Write(logBytes)
	assert.Nil(err)
	assert.Equal(len(logBytes), n)
	assert.Equal(getLogsNoPrefix(logs), out.String())

	out.Reset()
	logs = []string{
		"One line log without prefix.",
	}

	logBytes = getLogsBytes(logs)
	n, err = writer.Write(logBytes)
	assert.Nil(err)
	assert.Equal(len(logBytes), n)
	assert.Equal(getLogsNoPrefix(logs), out.String())
}
