package pinger

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"time"
)

type Item struct {
	Name    string
	Url     string
	Request Request
	Result  ItemResult
}

type ItemsGroup struct {
	Timeout time.Duration
	Items   []Item
}

type ItemTrigger struct {
	Request Request
	Timeout time.Duration
	Items   []Item
}

type ItemResult struct {
	Ping   *ItemPingOptions
	Status ItemResultStatus
	Body   string
}

type ItemPingOptions struct {
	Address string
	Timeout time.Duration
	Repeat  int
}

type ItemResultStatus struct {
	Code int
	Min  int
	Max  int
	List []int
}

type Request struct {
	Method string
	URL    string
	Body   string
	Header map[string][]string
	Proxy  *Proxy
}

type Proxy struct {
	Host string
	Port string
	User string
	Pass string
	Key  string
}

func (i Item) CheckPing(duration time.Duration, repeat int) Item {
	i.Result = ItemResult{
		Ping: &ItemPingOptions{
			Address: i.Url,
			Timeout: duration,
			Repeat:  repeat,
		},
	}
	return i
}

func (i Item) CheckStatus() Item {
	i.Result = ItemResult{
		Status: ItemResultStatus{
			Min: 200,
			Max: 299,
		},
	}
	return i
}

func (i Item) CheckBody(body string) Item {
	i.Result = ItemResult{
		Body: body,
	}
	return i
}

func (i Item) buildRequest() (*http.Request, error) {
	var body io.Reader
	var err error
	req := &http.Request{}

	_, err = url.Parse(i.Url)
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
	req, err = http.NewRequest(i.Request.Method, i.Url, body)
	return req, nil
}
