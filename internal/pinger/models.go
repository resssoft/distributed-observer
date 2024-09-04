package pinger

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"time"
)

type Item struct {
	Name    string     `json:"name"`
	Url     string     `json:"url"`
	Request Request    `json:"request"`
	Result  ItemResult `json:"result"`
}

type ItemsGroup struct {
	Timeout time.Duration `json:"timeout"`
	Items   []Item        `json:"items"`
}

type ItemTrigger struct {
	Request Request       `json:"request"`
	Timeout time.Duration `json:"timeout"`
	Items   []Item        `json:"items"`
}

type ItemResult struct {
	Ping   *ItemPingOptions `json:"ping"`
	Status ItemResultStatus `json:"status"`
	Body   string           `json:"body"`
}

type ItemPingOptions struct {
	Address string        `json:"address"`
	Timeout time.Duration `json:"timeout"`
	Repeat  int           `json:"repeat"`
}

type ItemResultStatus struct {
	Code int   `json:"code"`
	Min  int   `json:"min"`
	Max  int   `json:"max"`
	List []int `json:"list"`
}

type Request struct {
	Method string              `json:"method"`
	URL    string              `json:"url"`
	Body   string              `json:"body"`
	Header map[string][]string `json:"header"`
	Proxy  *Proxy              `json:"proxy"`
}

type Proxy struct {
	Host string `json:"host"`
	Port string `json:"port"`
	User string `json:"user"`
	Pass string `json:"pass"`
	Key  string `json:"key"`
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
