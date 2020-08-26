package repair

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/pmezard/go-difflib/difflib"
	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/context"
	"github.com/tarantool/cartridge-cli/cli/project"
)

func getAppInstanceNames(ctx *context.Ctx) ([]string, error) {
	if err := project.SetSystemRunningPaths(ctx); err != nil {
		return nil, fmt.Errorf("Failed to get default paths: %s", err)
	}

	if fileInfo, err := os.Stat(ctx.Running.DataDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("Data directory %s doesn't exist", ctx.Running.DataDir)
	} else if err != nil {
		return nil, fmt.Errorf("Failed to use specified data directory: %s", err)
	} else if !fileInfo.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", ctx.Running.DataDir)
	}

	workDirs, err := ioutil.ReadDir(ctx.Running.DataDir)
	if err != nil {
		return nil, fmt.Errorf("Failed to list the data directory: %s", err)
	}

	appWorkDirsPrefix := fmt.Sprintf("%s.", ctx.Project.Name)
	instanceNames := make([]string, 0)
	for _, workDir := range workDirs {
		workDirName := workDir.Name()
		if strings.HasPrefix(workDirName, appWorkDirsPrefix) {
			instanceName := strings.SplitN(workDirName, ".", 2)[1]
			if instanceName != "" {
				instanceNames = append(instanceNames, instanceName)
			}
		}
	}

	if len(instanceNames) == 0 {
		return nil, fmt.Errorf("No instance working directories found in %s", ctx.Running.DataDir)
	}

	return instanceNames, nil
}

func getBackupPath(path string) string {
	return fmt.Sprintf("%s.bak", path)
}

func createFileBackup(path string) (string, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("Failed to use specified path: %s", err)
	}

	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("Failed to open file: %s", err)
	}

	backupPath := getBackupPath(path)
	backupFile, err := os.OpenFile(backupPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, fileInfo.Mode())
	if err != nil {
		return "", fmt.Errorf("Failed to open backup file: %s", err)
	}

	if _, err := io.Copy(backupFile, file); err != nil {
		return "", fmt.Errorf("Failed to copy file content: %s", err)
	}

	return backupPath, nil
}

func getDiffLines(confBefore []byte, confAfter []byte, from string, to string) ([]string, error) {
	diff := difflib.UnifiedDiff{
		A:        difflib.SplitLines(string(confBefore)),
		B:        difflib.SplitLines(string(confAfter)),
		FromFile: from,
		ToFile:   to,
		Context:  5,
	}

	diffString, err := difflib.GetUnifiedDiffString(diff)
	if err != nil {
		return nil, err
	}

	// colorize log lines
	logLines := strings.Split(strings.TrimSpace(diffString), "\n")
	if len(logLines) == 1 && logLines[0] == "" {
		logLines = nil
	}

	for i := range logLines {
		if strings.HasPrefix(logLines[i], "-") {
			logLines[i] = common.ColorRed.Sprintf(logLines[i])
		} else if strings.HasPrefix(logLines[i], "+") {
			logLines[i] = common.ColorGreen.Sprintf(logLines[i])
		}
	}

	return logLines, nil
}
