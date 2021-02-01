package static

import (
	"io/ioutil"
	"net/http"

	"github.com/apex/log"
)

// GetStaticFileContent open file in generated static filesystem
func GetStaticFileContent(fs http.FileSystem, filename string) (string, error) {
	file, err := fs.Open(filename)
	if err != nil {
		log.Errorf("Failed to open static file: %s", err)
		return "", err
	}

	content, err := ioutil.ReadAll(file)
	if err != nil {
		log.Errorf("Failed to get static file content: %s", err)
		return "", err
	}

	defer file.Close()

	return string(content), nil
}
