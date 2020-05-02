package goline

import (
    "io/ioutil"
    "os"
    "io"
)

/* files */

func (p *LINE) SaveFile(path string, data []byte) (bool) {
	err := ioutil.WriteFile(path, data, 0644)
	if err!=nil {
		return false
	}
	return true
}

func (p *LINE) DeleteFile(path string) (bool) {
	err := os.Remove(path)
	if err!=nil {
		return false
	}
	return true
}

func (p *LINE) DownloadFile(url string) (string) {
	tmpfile, err := ioutil.TempFile("/tmp","Line-GOBOT-")
	if err!=nil {
		return "Failed."
	}
	res, err := p.GetContent(url, nil)
	if err!=nil {
		return "Failed."
	}
	if _, err := io.Copy(tmpfile, res.Body); err!=nil {
		return "Failed."
	}
	return tmpfile.Name()
}