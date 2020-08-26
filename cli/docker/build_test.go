package docker

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testReadCloser struct {
	data bytes.Buffer
}

func (r *testReadCloser) Put(data string) {
	r.data.WriteString(data)
}

func (r *testReadCloser) Clear() {
	r.data.Reset()
}

func (r *testReadCloser) Read(p []byte) (int, error) {
	return r.data.Read(p)
}

func (r *testReadCloser) Close() error {
	return nil
}

func TestPrintBuildOutput(t *testing.T) {
	t.Parallel()

	readerSize = 30
	assert := assert.New(t)

	r := testReadCloser{}
	outBuf := bytes.NewBuffer(nil)

	// put valid JSON data
	r.Clear()
	outBuf.Reset()
	r.Put(`{"stream":"I am stream"}`)

	assert.Nil(printBuildOutput(outBuf, &r))
	assert.Equal("I am stream", outBuf.String())

	// put long data
	r.Clear()
	outBuf.Reset()
	longString := strings.Repeat("a", readerSize+1)
	r.Put(fmt.Sprintf(`{"stream":"%s"}`, longString))

	assert.Nil(printBuildOutput(outBuf, &r))
	assert.Equal(longString, outBuf.String())
}
