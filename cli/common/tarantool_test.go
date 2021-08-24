package common

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTarantoolVersion(t *testing.T) {
	assert := assert.New(t)

	var err error

	dir, err := ioutil.TempDir(os.TempDir(), "temp")
	assert.Equal(err, nil)
	defer os.RemoveAll(dir)

	expectedTarantoolVersions := []string{
		"2.10.0-beta1-0-g7da4b1438",
		"2.10.0-beta1",
		"2.10.0",

		"2.8.1-0-ge2a1ec0c2-r399",
		"2.8.1-0-ge2a1ec0c2-r399-macos",

		"2.8.2-r420",
		"2.8.2-r420-macos",

		"2.10.0-beta1-r420",
		"2.10.0-beta1-r420-macos",

		"2.10.2-149-g1575f3c07-dev",
		"3.0.0-alpha1-14-gxxxxxxxxx-dev",
		"3.0.0-entrypoint-17-gxxxxxxxxx-dev",
		"3.1.2-5-gxxxxxxxxx-dev",

		"3.0.0-alpha1",
		"3.0.0-alpha2",
		"3.0.0-beta1",
		"3.0.0-beta2",
		"3.0.0-rc1",
		"3.0.0-rc2",
	}

	for _, version := range expectedTarantoolVersions {
		content := []byte(fmt.Sprintf("#!/bin/sh\necho %s", version))
		err = ioutil.WriteFile(filepath.Join(dir, "tarantool"), content, 0777)
		assert.Nil(err)

		tarantoolVersion, err := GetTarantoolVersion(dir)
		assert.Nil(err)
		assert.Equal(tarantoolVersion, version)
	}
}
