package goline

import (
	"net/http"
	"io"
)

var client = &http.Client{}

func (p *LINE) getNewRequest(method string, url string, body io.Reader) (res *http.Response, err error) {
	req, _ := http.NewRequest(method, url, body)
	req.Header.Add("User-Agent",USER_AGENT)
	req.Header.Add("X-Line-Application",p.AppName)
	req.Header.Add("X-Line-Carrier",CARRIER)
	req.Header.Add("X-Line-Access",p.AuthToken)
	req.Header.Add("Connection","Keep-Alive")
	res, erro := client.Do(req)
	return res, erro
}

func (p *LINE) PostContent(url string, body io.Reader) (res *http.Response, err error) {
	res, erro := p.getNewRequest("POST", url, body)
	return res, erro
}

func (p *LINE) GetContent(url string, body io.Reader) (res *http.Response, err error) {
	res, erro := p.getNewRequest("GET", url, body)
	return res, erro
}

func (p *LINE) DeleteContent(url string, body io.Reader) (res *http.Response, err error) {
	res, erro := p.getNewRequest("DELETE", url, body)
	return res, erro
}

func (p *LINE) PutContent(url string, body io.Reader) (res *http.Response, err error) {
	res, erro := p.getNewRequest("PUT", url, body)
	return res, erro
}