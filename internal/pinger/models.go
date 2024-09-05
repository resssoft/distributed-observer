package pinger

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	StatusSuccess = "success"
	StatusFailure = "failure"
	StatusCustom  = "custom"
)

type ItemsGroup struct {
	Timeout time.Duration `json:"timeout"`
	Items   []Item        `json:"items"`
}

type Item struct {
	Id      interface{} `json:"id"`
	Name    string      `json:"name"`
	Order   int         `json:"order"`
	Request Request     `json:"request"`
	Status  Status      `json:"status"`
}

type Status struct {
	Name          string    `json:"name"`
	EventsCount   int       `json:"events_count"`
	LastEventDate time.Time `json:"last_event_date"`
	LastCode      int       `json:"last_code"`
}

type Trigger struct {
	Antispam     *time.Duration `json:"antispam"`
	SkipBy       int            `json:"skip_by"`
	OnSuccessful *Request       `json:"on_successful"`
	OnFail       *Request       `json:"on_fail"`
	Always       *Request       `json:"always"`
}

type History struct { // TODO: implement
	ItemId      interface{}           `json:"item_id"`
	ItemName    string                `json:"item_name"`
	EventDate   time.Time             `json:"event_date"`
	LastRequest Request               `json:"last_request"`
	Requests    map[time.Time]Request `json:"requests"`
}

type Request struct {
	Method   string              `json:"method"`
	Url      string              `json:"url"`
	Body     string              `json:"body"`
	Header   map[string][]string `json:"header"`
	Proxy    *Proxy              `json:"proxy"`
	Ping     string              `json:"address"`
	Repeat   int                 `json:"repeat"`
	Timeout  time.Duration       `json:"timeout"`
	Response Response            `json:"response"`
	Trigger  *Trigger            `json:"trigger"`
}

type Response struct {
	Status   ItemResultStatus `json:"status"`
	Body     *ResponseBody    `json:"body"`
	SaveBody bool             `json:"save_body"`
}

type ItemResultStatus struct {
	Code int   `json:"code"`
	Min  int   `json:"min"`
	Max  int   `json:"max"`
	List []int `json:"list"`
}

type ResponseBody struct {
	Full    string `json:"full"`
	Contain string `json:"contain"`
	Regex   string `json:"regex"`
	Grep    *Grep  `json:"grep"`
}

type ResponseResult struct {
	Successful bool   `json:"successful"`
	StatusCode int    `json:"status_code"`
	Body       string `json:"body"`
	Error      string `json:"error"`
}

func (rr ResponseResult) WithErr(format string, err error) ResponseResult {
	if err != nil {
		rr.Error = fmt.Sprintf(format, err.Error())
	} else {
		rr.Error = format
	}
	return rr
}

func (rr ResponseResult) SetErr(e string) ResponseResult {
	rr.Error = e
	return rr
}

type Grep struct {
	Xpath    string `json:"xpath"`
	JsonPath string `json:"json_path"`
}

type Proxy struct {
	Host string `json:"host"`
	Port string `json:"port"`
	User string `json:"user"`
	Pass string `json:"pass"`
	Key  string `json:"key"`
}

func PingItem(address string, duration time.Duration, repeat int) Item {
	newItem := Item{
		Request: Request{
			Ping:    address,
			Timeout: duration,
			Repeat:  repeat,
		},
	}
	return newItem
}

func CheckStatusItem(url string) Item {
	newItem := Item{
		Request: Request{
			Url: url,
			Response: Response{
				Status: ItemResultStatus{
					Min: 200,
					Max: 299,
				},
			},
		},
	}
	return newItem
}

func (i Item) CheckFullBody(body string) Item {
	i.Request.Response.Body.Full = body
	return i
}

func (i Item) buildRequest() (*http.Request, error) {
	var body io.Reader
	var err error
	req := &http.Request{}

	_, err = url.Parse(i.Request.Url)
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
	req, err = http.NewRequest(i.Request.Method, i.Request.Url, body)
	return req, nil
}
