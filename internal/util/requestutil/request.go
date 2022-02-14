package requestutil

import (
	"github.com/wallacehenriquesilva/slack-assets-bot/internal/util/fileutil"
	"io/fs"
	"io/ioutil"
	"net/http"
)

func DownloadFile(url string, headers map[string]string, fileExtension string) (string, error) {
	request, err := getWithHeaders(url, headers)
	if err != nil {
		return "", err
	}

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	file, err := bodyToFile(body, fileExtension)
	if err != nil {
		return "", err
	}

	return file, nil
}

func getWithHeaders(url string, headers map[string]string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	for k, v := range headers {
		req.Header.Add(k, v)
	}

	return req, err
}

func bodyToFile(body []byte, fileExtension string) (string, error) {
	file, err := fileutil.NewFile(fileExtension)
	if err != nil {
		return "", err
	}

	err = ioutil.WriteFile(file.Name(), body, fs.ModeTemporary)
	if err != nil {
		return "", err
	}

	return file.Name(), nil
}
