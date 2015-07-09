package utility

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func DownloadFromURL(url string, folder string, fileName string) (*string, error) {

	if fileName == "" {
		tokens := strings.Split(url, "/")
		fileName = tokens[len(tokens)-1]
	}
	path := filepath.FromSlash(folder + "/" + fileName)

	output, err := os.Create(path)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error while creating %v: %v", path, err))
	}
	defer output.Close()

	response, err := http.Get(url)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error while downloading %v: %v", url, err))
	}
	defer response.Body.Close()

	_, err = io.Copy(output, response.Body)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error while downloading %v: %v", url, err))
	}

	return &path, nil
}
