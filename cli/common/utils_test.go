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

	// last 2 lines w/ n = -2
	writeFile(f, "one\ntwo\nthree\nfour\nfive\nsix\nseven\n")
	n, err = GetLastNLinesBegin(f.Name(), -2)
	assert.Nil(err)
	assert.Equal("six\nseven\n", getFileContentSinceOffset(f, n))

	// last 100 lines
	writeFile(f, "one\ntwo\nthree\nfour\nfive\nsix\nseven")
	n, err = GetLastNLinesBegin(f.Name(), 100)
	assert.Nil(err)
	assert.EqualValues(0, n)

	// last 100 lines w/ n = -100
	writeFile(f, "one\ntwo\nthree\nfour\nfive\nsix\nseven")
	n, err = GetLastNLinesBegin(f.Name(), -100)
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

func TestGetInstancesFromArgs(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)

	var err error
	var args []string
	var instances []string

	projectName := "myapp"

	// wrong format
	args = []string{"myapp.instance-1", "myapp.instance-2"}
	_, err = GetInstancesFromArgs(args, projectName)
	assert.EqualError(err, instanceIDSpecified)

	args = []string{"instance-1", "myapp.instance-2"}
	_, err = GetInstancesFromArgs(args, projectName)
	assert.EqualError(err, instanceIDSpecified)

	args = []string{"myapp"}
	_, err = GetInstancesFromArgs(args, projectName)
	assert.True(strings.Contains(err.Error(), appNameSpecifiedError))

	// duplicate instance name
	args = []string{"instance-1", "instance-1"}
	_, err = GetInstancesFromArgs(args, projectName)
	assert.True(strings.Contains(err.Error(), "Duplicate instance name specified: instance-1"))

	// instances are specified
	args = []string{"instance-1", "instance-2"}
	instances, err = GetInstancesFromArgs(args, projectName)
	assert.Nil(err)
	assert.Equal([]string{"instance-1", "instance-2"}, instances)

	// specified both app name and instance name
	args = []string{"instance-1", "myapp"}
	instances, err = GetInstancesFromArgs(args, projectName)
	assert.EqualError(err, appNameSpecifiedError)

	args = []string{"myapp", "instance-1"}
	instances, err = GetInstancesFromArgs(args, projectName)
	assert.EqualError(err, appNameSpecifiedError)
}

func TestCorrectDependencyParsing(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)

	rawDeps := []string{
		"dependency_01 > 1.2, <= 4",
		"dependency_02 < 7,>=1.5",
		"dependency_03==2.8 // comment here",
		"// comment",
		"	dependency_04   <= 5.2   ",
		"",
		"\tdependency_05>=2.4,<=5.1",
		"dependency_06 =15",
	}

	deps, err := ParseDependencies(rawDeps)

	assert.Equal(nil, err)
	assert.Equal(deps, PackDependencies{
		{
			Name: "dependency_01",
			Relations: []DepRelation{
				{
					Relation: ">",
					Version:  "1.2",
				},
				{
					Relation: "<=",
					Version:  "4",
				},
			}},
		{
			Name: "dependency_02",
			Relations: []DepRelation{
				{
					Relation: "<",
					Version:  "7",
				},
				{
					Relation: ">=",
					Version:  "1.5",
				},
			}},
		{
			Name: "dependency_03",
			Relations: []DepRelation{
				{
					Relation: "==",
					Version:  "2.8",
				},
			}},
		{
			Name: "dependency_04",
			Relations: []DepRelation{
				{
					Relation: "<=",
					Version:  "5.2",
				},
			},
		},
		{
			Name: "dependency_05",
			Relations: []DepRelation{
				{
					Relation: ">=",
					Version:  "2.4",
				},
				{
					Relation: "<=",
					Version:  "5.1",
				},
			}},
		{
			Name: "dependency_06",
			Relations: []DepRelation{
				{
					Relation: "=",
					Version:  "15",
				},
			},
		},
	})
}

func TestInvalidDependencyParsing(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)
	errMsg := "Error during parse dependencies file"
	rawDeps := []string{
		"dependency_01 > 1.2, <= 4",
		"dep01 >= 14, <= 25, > 14",
	}

	deps, err := ParseDependencies(rawDeps)
	assert.Contains(err.Error(), errMsg)
	assert.Equal(PackDependencies(nil), deps)

	rawDeps = []string{
		"broke broke broke broke broke broke broke",
	}

	deps, err = ParseDependencies(rawDeps)
	assert.Contains(err.Error(), errMsg)
	assert.Equal(PackDependencies(nil), deps)

	rawDeps = []string{
		"dependency , ,",
	}

	deps, err = ParseDependencies(rawDeps)

	assert.Contains(err.Error(), errMsg)
	assert.Equal(PackDependencies(nil), deps)

	rawDeps = []string{
		"dependency >= 3.2, < ",
	}

	deps, err = ParseDependencies(rawDeps)

	assert.Contains(err.Error(), errMsg)
	assert.Equal(PackDependencies(nil), deps)
}

func TestCorrectPreAndPostInstallScriptParsing(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)

	userScript := "pwd\n" +
                  "ls $HOME\n" +
                  "cat some_file\n"

	expectedScript := "/bin/sh -c 'pwd'\n" +
                      "/bin/sh -c 'ls $HOME'\n" +
                      "/bin/sh -c 'cat some_file'"

	parsedScript, err := ParseUserInstallScript(userScript)

	assert.Equal(nil, err)
	assert.Equal(expectedScript, parsedScript)

	userScript = "#!/bin/sh\n" +
                 "pwd\n" +
                 "ls $HOME\n" +
                 "cat some_file\n"

	parsedScript, err = ParseUserInstallScript(userScript)

	assert.Equal(nil, err)
	assert.Equal(expectedScript, parsedScript)
}

func TestInvalidPreAndPostInstallScriptParsing(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)
	errMsg := "Unknown command bad_command some_file"
	userScript := "echo hi\n" +
                  "bad_command some_file\n" +
                  "ln -s some_file created_link\n"

	parsedScript, err := ParseUserInstallScript(userScript)

	assert.Contains(err.Error(), errMsg)
	assert.Equal("", parsedScript)
}
