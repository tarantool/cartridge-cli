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

type returnValueParseTarantoolVersion struct {
	version TarantoolVersion
	err     error
}

func TestParseTarantoolVersion(t *testing.T) {
	assert := assert.New(t)

	testCases := make(map[string]returnValueParseTarantoolVersion)

	testCases["1.10.11-0-g543e2a1ec0"] = returnValueParseTarantoolVersion{
		TarantoolVersion{
			Major:                 1,
			Minor:                 10,
			Patch:                 11,
			TagSuffix:             "",
			CommitsSinceTag:       0,
			CommitHashId:          "g543e2a1ec0",
			EnterpriseSDKRevision: "",
			EnterpriseIsOnMacOS:   false,
			IsDevelopmentBuild:    false,
		},
		nil,
	}

	testCases["2.8.1-0-ge2a1ec0c2"] = returnValueParseTarantoolVersion{
		TarantoolVersion{
			Major:                 2,
			Minor:                 8,
			Patch:                 1,
			TagSuffix:             "",
			CommitsSinceTag:       0,
			CommitHashId:          "ge2a1ec0c2",
			EnterpriseSDKRevision: "",
			EnterpriseIsOnMacOS:   false,
			IsDevelopmentBuild:    false,
		},
		nil,
	}

	testCases["2.10.0-alpha1-0-g4b387d14a"] = returnValueParseTarantoolVersion{
		TarantoolVersion{
			Major:                 2,
			Minor:                 10,
			Patch:                 0,
			TagSuffix:             "alpha1",
			CommitsSinceTag:       0,
			CommitHashId:          "g4b387d14a",
			EnterpriseSDKRevision: "",
			EnterpriseIsOnMacOS:   false,
			IsDevelopmentBuild:    false,
		},
		nil,
	}

	testCases["2.10.0-beta1-0-g7da4b1438"] = returnValueParseTarantoolVersion{
		TarantoolVersion{
			Major:                 2,
			Minor:                 10,
			Patch:                 0,
			TagSuffix:             "beta1",
			CommitsSinceTag:       0,
			CommitHashId:          "g7da4b1438",
			EnterpriseSDKRevision: "",
			EnterpriseIsOnMacOS:   false,
			IsDevelopmentBuild:    false,
		},
		nil,
	}

	testCases["2.10.0-rc3-0-gc2438eeb1"] = returnValueParseTarantoolVersion{
		TarantoolVersion{
			Major:                 2,
			Minor:                 10,
			Patch:                 0,
			TagSuffix:             "rc3",
			CommitsSinceTag:       0,
			CommitHashId:          "gc2438eeb1",
			EnterpriseSDKRevision: "",
			EnterpriseIsOnMacOS:   false,
			IsDevelopmentBuild:    false,
		},
		nil,
	}

	testCases["2.10.1-entrypoint-0-gc2438eeb1"] = returnValueParseTarantoolVersion{
		TarantoolVersion{
			Major:                 2,
			Minor:                 10,
			Patch:                 1,
			TagSuffix:             "entrypoint",
			CommitsSinceTag:       0,
			CommitHashId:          "gc2438eeb1",
			EnterpriseSDKRevision: "",
			EnterpriseIsOnMacOS:   false,
			IsDevelopmentBuild:    false,
		},
		nil,
	}

	testCases["2.8.1-121-g94ad1c2ee"] = returnValueParseTarantoolVersion{
		TarantoolVersion{
			Major:                 2,
			Minor:                 8,
			Patch:                 1,
			TagSuffix:             "",
			CommitsSinceTag:       121,
			CommitHashId:          "g94ad1c2ee",
			EnterpriseSDKRevision: "",
			EnterpriseIsOnMacOS:   false,
			IsDevelopmentBuild:    false,
		},
		nil,
	}

	testCases["2.8.1-0-ge2a1ec0c2-r399"] = returnValueParseTarantoolVersion{
		TarantoolVersion{
			Major:                 2,
			Minor:                 8,
			Patch:                 1,
			TagSuffix:             "",
			CommitsSinceTag:       0,
			CommitHashId:          "ge2a1ec0c2",
			EnterpriseSDKRevision: "r399",
			EnterpriseIsOnMacOS:   false,
			IsDevelopmentBuild:    false,
		},
		nil,
	}

	testCases["2.8.1-0-ge2a1ec0c2-r399-macos"] = returnValueParseTarantoolVersion{
		TarantoolVersion{
			Major:                 2,
			Minor:                 8,
			Patch:                 1,
			TagSuffix:             "",
			CommitsSinceTag:       0,
			CommitHashId:          "ge2a1ec0c2",
			EnterpriseSDKRevision: "r399",
			EnterpriseIsOnMacOS:   true,
			IsDevelopmentBuild:    false,
		},
		nil,
	}

	testCases["2.10.1-23-g0c2e2a1ec-dev"] = returnValueParseTarantoolVersion{
		TarantoolVersion{
			Major:                 2,
			Minor:                 10,
			Patch:                 1,
			TagSuffix:             "",
			CommitsSinceTag:       23,
			CommitHashId:          "g0c2e2a1ec",
			EnterpriseSDKRevision: "",
			EnterpriseIsOnMacOS:   false,
			IsDevelopmentBuild:    true,
		},
		nil,
	}

	testCases["2.10.0-beta1-3-g4b17da438-dev"] = returnValueParseTarantoolVersion{
		TarantoolVersion{
			Major:                 2,
			Minor:                 10,
			Patch:                 0,
			TagSuffix:             "beta1",
			CommitsSinceTag:       3,
			CommitHashId:          "g4b17da438",
			EnterpriseSDKRevision: "",
			EnterpriseIsOnMacOS:   false,
			IsDevelopmentBuild:    true,
		},
		nil,
	}

	testCases["2.8.2"] = returnValueParseTarantoolVersion{
		TarantoolVersion{},
		fmt.Errorf("Failed to parse Tarantool version: format is not valid"),
	}

	testCases["2.10.0-beta1"] = returnValueParseTarantoolVersion{
		TarantoolVersion{},
		fmt.Errorf("Failed to parse Tarantool version: format is not valid"),
	}

	testCases["2.8.1-ge2a1ec0c2-0"] = returnValueParseTarantoolVersion{
		TarantoolVersion{},
		fmt.Errorf("Failed to parse Tarantool version: format is not valid"),
	}

	testCases["2.10.0-beta1-0"] = returnValueParseTarantoolVersion{
		TarantoolVersion{},
		fmt.Errorf("Failed to parse Tarantool version: format is not valid"),
	}

	testCases["2.8.2-r420"] = returnValueParseTarantoolVersion{
		TarantoolVersion{},
		fmt.Errorf("Failed to parse Tarantool version: format is not valid"),
	}

	testCases["2.8.2-r420-macos"] = returnValueParseTarantoolVersion{
		TarantoolVersion{},
		fmt.Errorf("Failed to parse Tarantool version: format is not valid"),
	}

	testCases["2.7.1.2-0-ge2a1ec0c2"] = returnValueParseTarantoolVersion{
		TarantoolVersion{},
		fmt.Errorf("Failed to parse Tarantool version: format is not valid"),
	}

	testCases["2.8.1-0-ge2a1ec0c2-macos"] = returnValueParseTarantoolVersion{
		TarantoolVersion{},
		fmt.Errorf("Failed to parse Tarantool version: format is not valid"),
	}

	for input, output := range testCases {
		version, err := ParseTarantoolVersion(input)

		if output.err == nil {
			assert.Nil(err)
			assert.Equal(output.version, version)
		} else {
			assert.Equal(output.err, err)
		}
	}
}
