/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements. See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership. The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License. You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package thrift

import (
	//"fmt"
	//"time"
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"github.com/valyala/fasthttp"
	"net/url"
	//"strconv"
)

// Default to using the shared http client. Library users are
// free to change this global client or specify one through
// THttpClientOptions.

//var DefaultHttpClient *fasthttp.Client = new(fasthttp.Client)

type THttpClient struct {
	client             *fasthttp.Client
	header             fasthttp.RequestHeader
	response           *fasthttp.Response
	url                *url.URL
	requestBuffer      *bytes.Buffer
	nsecConnectTimeout int64
	nsecReadTimeout    int64
	loop               int64
	respReader         *bytes.Reader
}

type THttpClientTransportFactory struct {
	options THttpClientOptions
	url     string
}

func (p *THttpClientTransportFactory) GetTransport(trans TTransport) (TTransport, error) {
	if trans != nil {
		t, ok := trans.(*THttpClient)
		if ok && t.url != nil {
			return NewTHttpClientWithOptions(t.url.String(), p.options)
		}
	}
	return NewTHttpClientWithOptions(p.url, p.options)
}

type THttpClientOptions struct {
	// If nil, DefaultHttpClient is used
	Client *fasthttp.Client
}

func NewTHttpClientTransportFactory(url string) *THttpClientTransportFactory {
	return NewTHttpClientTransportFactoryWithOptions(url, THttpClientOptions{})
}

func NewTHttpClientTransportFactoryWithOptions(url string, options THttpClientOptions) *THttpClientTransportFactory {
	return &THttpClientTransportFactory{url: url, options: options}
}

func NewTHttpClientWithOptions(urlstr string, options THttpClientOptions) (TTransport, error) {
	parsedURL, err := url.Parse(urlstr)
	if err != nil {
		return nil, err
	}
	buf := make([]byte, 0, 2048)
	client := options.Client
	if client == nil {
		client = new(fasthttp.Client)
	}
	Header := fasthttp.RequestHeader{}
	Header.SetContentType("application/x-thrift")
	Header.Add("Content-Type", "application/x-thrift")
    Header.SetMethod("POST")
    response := fasthttp.AcquireResponse()
	return &THttpClient{client: client, header: Header, response: response, url: parsedURL, requestBuffer: bytes.NewBuffer(buf), loop: 0, respReader: bytes.NewReader(buf)}, nil
}

func NewTHttpClient(urlstr string) (TTransport, error) {
	return NewTHttpClientWithOptions(urlstr, THttpClientOptions{})
}

// Set the HTTP Header for this specific Thrift Transport
// It is important that you first assert the TTransport as a THttpClient type
// like so:
//
// httpTrans := trans.(THttpClient)
// httpTrans.SetHeader("User-Agent","Thrift Client 1.0")
func (p *THttpClient) SetHeader(key string, value string) {
	if key != "" && value != "" {
		p.header.Add(key, value)
	}
}

// Get the HTTP Header represented by the supplied Header Key for this specific Thrift Transport
// It is important that you first assert the TTransport as a THttpClient type
// like so:
//
// httpTrans := trans.(THttpClient)
// hdrValue := httpTrans.GetHeader("User-Agent")
func (p *THttpClient) GetHeader(key string) string {
	return string(p.header.Peek(key))
}

// Deletes the HTTP Header given a Header Key for this specific Thrift Transport
// It is important that you first assert the TTransport as a THttpClient type
// like so:
//
// httpTrans := trans.(THttpClient)
// httpTrans.DelHeader("User-Agent")
func (p *THttpClient) DelHeader(key string) {
	p.header.Del(key)
}

func (p *THttpClient) Open() error {
	// do nothing
	return nil
}

func (p *THttpClient) IsOpen() bool {
	return p.response != nil || p.requestBuffer != nil
}

func (p *THttpClient) closeResponse() error {
	if p.response != nil {
	    io.Copy(ioutil.Discard, p.respReader)
	}
	return nil
}

func (p *THttpClient) Close() error {
	if p.requestBuffer != nil {
		p.requestBuffer.Reset()
	}
	return p.closeResponse()
}

func (p *THttpClient) Read(buf []byte) (int, error) {
	
	n, err := p.respReader.Read(buf)
	if n > 0 && (err == nil || err == io.EOF) {
		return n, nil
	}
	return n, NewTTransportExceptionFromError(err)
}

func (p *THttpClient) ReadByte() (c byte, err error) {
	return readByte(p.respReader)
}

func (p *THttpClient) Write(buf []byte) (int, error) {
	return p.requestBuffer.Write(buf)
}

func (p *THttpClient) WriteByte(c byte) error {
	return p.requestBuffer.WriteByte(c)
}

func (p *THttpClient) WriteString(s string) (n int, err error) {
	return p.requestBuffer.WriteString(s)
}

func (p *THttpClient) Flush() error {
	// Close any previous response body to avoid leaking connections.
	//if p.loop <= 2 {
	//    if p.IsOpen(){ p.closeResponse()}; p.loop += 1
	//}
	//:= time.Now()
	req := fasthttp.AcquireRequest()
	defer func(){
		fasthttp.ReleaseRequest(req)
	    fasthttp.ReleaseResponse(p.response)
	}()
	req.Header = p.header
    req.SetRequestURI(p.url.String())
    req.SetBody(p.requestBuffer.Bytes())
	err := p.client.Do(req, p.response)
	if err != nil {
		return NewTTransportExceptionFromError(err)
	}
	if p.response.StatusCode() != http.StatusOK {
		// Close the response to avoid leaking file descriptors. closeResponse does
		// more than just call Close(), so temporarily assign it and reuse the logic.
        p.respReader.Reset(p.response.Body())
        p.requestBuffer.Reset()
        return nil

		// TODO(pomack) log bad response
		///return NewTTransportException(UNKNOWN_TRANSPORT_EXCEPTION, "HTTP Response code: "+strconv.Itoa(resp.StatusCode()))
	}

    p.respReader.Reset(p.response.Body())
    p.requestBuffer.Reset()

	return nil
}

func (p *THttpClient) RemainingBytes() (num_bytes uint64) {
	len := p.respReader.Size()
	if len >= 0 {
		return uint64(len)
	}

	const maxSize = ^uint64(27)
	return maxSize // the thruth is, we just don't know unless framed is used
}

// Deprecated: Use NewTHttpClientTransportFactory instead.
func NewTHttpPostClientTransportFactory(url string) *THttpClientTransportFactory {
	return NewTHttpClientTransportFactoryWithOptions(url, THttpClientOptions{})
}

// Deprecated: Use NewTHttpClientTransportFactoryWithOptions instead.
func NewTHttpPostClientTransportFactoryWithOptions(url string, options THttpClientOptions) *THttpClientTransportFactory {
	return NewTHttpClientTransportFactoryWithOptions(url, options)
}

// Deprecated: Use NewTHttpClientWithOptions instead.
func NewTHttpPostClientWithOptions(urlstr string, options THttpClientOptions) (TTransport, error) {
	return NewTHttpClientWithOptions(urlstr, options)
}

// Deprecated: Use NewTHttpClient instead.
func NewTHttpPostClient(urlstr string) (TTransport, error) {
	return NewTHttpClientWithOptions(urlstr, THttpClientOptions{})
}
