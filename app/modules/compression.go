package modules

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
	"strings"
)

func Gzip(input string) (output string, err error) {
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	_, err = w.Write([]byte(input))
	if err != nil {
		return "", err
	}

	err = w.Close()
	if err != nil {
		return "", err
	}

	return b.String(), nil
}

func Gunzip(input string) (output string, err error) {
	r := strings.NewReader(input)
	gread, err := gzip.NewReader(r)

	if err != nil {
		return "", err
	}

	defer gread.Close()
	bytes, err := ioutil.ReadAll(gread)

	if err != nil {
		return "", err
	}

	return string(bytes), nil
}
