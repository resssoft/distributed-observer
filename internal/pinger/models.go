package pinger

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
)

type Item struct {
	name    string
	url     string
	Request Request
	proxy   string //TODO: struct with creds
	result  ItemResult
}

type ItemResult struct {
	ping   bool
	status ItemResultStatus
	body   string
}

type ItemResultStatus struct {
	code int
	min  int
	max  int
	list []int
}

type Request struct {
	Method string
	URL    string
	Body   string
	Header map[string][]string
	proxy  Proxy
}

type Proxy struct {
	Host  string
	Port  string
	Login string
	Pass  string
	Key   string
}

func (i Item) CheckPing() Item {
	i.result = ItemResult{
		ping: true,
	}
	return i
}

func (i Item) CheckStatus() Item {
	i.result = ItemResult{
		status: ItemResultStatus{
			min: 200,
			max: 299,
		},
	}
	return i
}

func (i Item) CheckBody(body string) Item {
	i.result = ItemResult{
		body: body,
	}
	return i
}

func (i Item) buildRequest() (*http.Request, error) {
	var body io.Reader
	var err error
	req := &http.Request{}

	_, err = url.Parse(i.url)
	if err != nil {
		return nil, err
	}

	if i.Request.Method == "" {
		i.Request.Method = "GET"
	}

	if i.Request.Body != "" {
		data := []byte(i.Request.Body)
		body = bytes.NewReader(data)
	}

	//req = &http.Request{
	//	Method:           i.Request.Method,
	//	URL:              nil,
	//	Proto:            "",
	//	ProtoMajor:       0,
	//	ProtoMinor:       0,
	//	Header:           i.Request.Header,
	//	Body:             nil,
	//	GetBody:          nil,
	//	ContentLength:    0,
	//	TransferEncoding: nil,
	//	Close:            false,
	//	Host:             "",
	//	Form:             nil,
	//	PostForm:         nil,
	//	MultipartForm:    nil,
	//	Trailer:          nil,
	//	RemoteAddr:       "",
	//	RequestURI:       "",
	//	TLS:              nil,
	//	Cancel:           nil,
	//	Response:         nil,
	//}
	req, err = http.NewRequest(i.Request.Method, i.url, body)
	return req, nil
}
