package project

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPath(t *testing.T) {
	assert := assert.New(t)

	var err error
	var path string
	var conf map[string]interface{}

	curDir, err := os.Getwd()
	assert.Nil(err)

	const specifiedPath = "specifiedPath"
	const defaultPath = "defaultPath"
	const sectionName = "sectionName"
	const otherSectionName = "otherSectionName"
	const sectionValue = "sectionValue"

	// path is specified
	path, err = getPath(nil, PathOpts{
		SpecifiedPath: specifiedPath,
		DefaultPath:   defaultPath,
	})
	assert.Nil(err)
	assert.Equal(specifiedPath, path)

	// path is specified, GetAbs
	path, err = getPath(nil, PathOpts{
		SpecifiedPath: specifiedPath,
		DefaultPath:   defaultPath,
		GetAbs:        true,
	})
	assert.Nil(err)
	assert.Equal(filepath.Join(curDir, specifiedPath), path)

	// path isn't specified
	path, err = getPath(nil, PathOpts{
		DefaultPath: defaultPath,
	})
	assert.Nil(err)
	assert.Equal(defaultPath, path)

	// path isn't specified, GetAbs
	path, err = getPath(nil, PathOpts{
		DefaultPath: defaultPath,
		GetAbs:      true,
	})
	assert.Nil(err)
	assert.Equal(filepath.Join(curDir, defaultPath), path)

	// path isn't specified, defaultPath is empty, GetAbs
	path, err = getPath(nil, PathOpts{
		DefaultPath: "",
		GetAbs:      true,
	})
	assert.Nil(err)
	assert.Equal("", path)

	// specified conf, but no section
	conf = map[string]interface{}{
		sectionName: sectionValue,
	}
	path, err = getPath(conf, PathOpts{
		DefaultPath: defaultPath,
	})
	assert.Nil(err)
	assert.Equal(defaultPath, path)

	// specified section, but no conf
	path, err = getPath(nil, PathOpts{
		DefaultPath:     defaultPath,
		ConfSectionName: sectionName,
	})
	assert.Nil(err)
	assert.Equal(defaultPath, path)

	// specified section not present in conf
	conf = map[string]interface{}{
		otherSectionName: sectionValue,
	}
	path, err = getPath(conf, PathOpts{
		DefaultPath:     defaultPath,
		ConfSectionName: sectionName,
	})
	assert.Nil(err)
	assert.Equal(defaultPath, path)

	// specified section present in conf
	conf = map[string]interface{}{
		sectionName: sectionValue,
	}
	path, err = getPath(conf, PathOpts{
		DefaultPath:     defaultPath,
		ConfSectionName: sectionName,
	})
	assert.Nil(err)
	assert.Equal(sectionValue, path)

	// specified section present in conf, GetAbs
	conf = map[string]interface{}{
		sectionName: sectionValue,
	}
	path, err = getPath(conf, PathOpts{
		DefaultPath:     defaultPath,
		ConfSectionName: sectionName,
		GetAbs:          true,
	})
	assert.Nil(err)
	assert.Equal(filepath.Join(curDir, sectionValue), path)

	// specified section present in conf with no string value
	conf = map[string]interface{}{
		sectionName: true,
	}
	path, err = getPath(conf, PathOpts{
		DefaultPath:     defaultPath,
		ConfSectionName: sectionName,
	})
	assert.True(strings.Contains(err.Error(), "config value should be string"))
}
