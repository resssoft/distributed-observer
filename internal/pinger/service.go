package pinger

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"runtime"
	"sync"
	"time"

	"observer/internal/domain/services"
	"observer/internal/logger"
	"observer/pkg/defaults"
	"observer/pkg/go-fastping"
	"observer/pkg/mediator"
)

const queueLimit = 10000

type Data struct {
	Items      []Item `json:"items"`
	dispatcher *mediator.Dispatcher
	logger     *logger.Logger
	settings   services.Settings
	queue      chan Item
}

func New(dispatcher *mediator.Dispatcher, logger *logger.Logger, settings services.Settings) *Data {
	items := make([]Item, 0)
	items = append(items,
		Item{url: "127.0.0.1"}.CheckPing(),
		Item{url: "http://127.0.0.1/"}.CheckStatus(),
		Item{url: "188.123.321.21"}.CheckPing(),
		Item{url: "http://nstarchitects.am/"}.CheckStatus(),
	)
	return &Data{
		dispatcher: dispatcher,
		logger:     logger,
		settings:   settings,
		queue:      make(chan Item, queueLimit),
		Items:      items,
	}
}

func (d *Data) Start(ctx context.Context) {
	println("start pinger")
	d.Send(ctx)
	for i := 0; i < runtime.NumCPU(); i++ {
		//go d.Sender(ctx)
		go d.Receiver(ctx)
	}
}

func (d *Data) Sender(ctx context.Context) {
	println("start Sender")
	for {
		select {
		case <-d.settings.AfterSeconds("OBSERVER_PINGER_SENDER_TIMEOUT_SEC", 15):
			d.logger.Debug(ctx, "send")
			d.Send(ctx)
		}
	}
}

func (d *Data) Send(ctx context.Context) {
	for _, item := range d.Items {
		d.logger.Info(ctx, "sending item", "item", item)
		d.queue <- item
	}
}

func (d *Data) Receiver(ctx context.Context) {
	for item := range d.queue {
		host := getHost(item.url)
		println("check item", item.url, host)
		if host != "" {
			if item.result.ping {
				println("check ping", item.url, host)
				state, err := d.ping(host)
				d.logger.Info(ctx, fmt.Sprintf("Received [%s] ping for url [%s] result %s",
					defaults.Str(item.name, item.url),
					item.url,
					fmt.Sprintf("%v err: %v", state, err),
				))
			} else {
				println("check web", item.url, host)
				state, err := d.web(item)
				d.logger.Info(ctx, fmt.Sprintf("Received [%s] ping for url [%s] result %s",
					defaults.Str(item.name, item.url),
					item.url,
					fmt.Sprintf("%v err: %v", state, err),
				))
			}
		}
	}
}

func (d Data) ping(address string) (bool, error) {
	println("ping address 1", address)
	state := false
	var err error
	var wg sync.WaitGroup
	wg.Add(1)
	println("ping address 2", address)
	p := fastping.NewPinger()
	p.MaxRTT = d.settings.GetValueSeconds("OBSERVER_PINGER_PING_TIMEOUT_SEC", 7)
	err = p.AddIP(address)
	if err != nil {
		println("ping address 2", address)
		return false, err
	}
	p.OnRecv = func(addr *net.IPAddr, rtt time.Duration) {
		println("ping address 3", address)
		state = true
	}
	p.OnIdle = func() {
		println("ping address 4", address)
		state = false
		defer wg.Done()
	}
	println("ping address 5", address)
	err = p.Run()
	println("ping address 6", address)
	wg.Wait()
	println("ping address 7", address)
	return state, err
}

func getHost(address string) string {
	if address == "" {
		return ""
	}
	parseIp := net.ParseIP(address)
	if parseIp != nil {
		return parseIp.String()
	}
	urlItem, err := url.Parse(address)
	if err != nil {
		return ""
	}
	return urlItem.Host
}

func (d Data) web(item Item) (bool, string) {
	client := &http.Client{}
	println("web address 1")
	if item.proxy != "" {
		proxyURL, _ := url.Parse(item.proxy)
		proxy := http.ProxyURL(proxyURL)
		transport := &http.Transport{Proxy: proxy}
		client.Transport = transport
	}
	println("web address 2")
	request, err := item.buildRequest()
	if err != nil {
		return false, "build request err: " + err.Error()
	}
	resp, err := client.Do(request)
	if err != nil {
		return false, "request err: " + err.Error()
	}
	println("web address 3")
	if resp == nil {
		return false, "empty response"
	}
	webBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, "read body err"
	}
	if item.result.status.code != 0 && resp.StatusCode == item.result.status.code {
		return true, ""
	}
	if item.result.status.min != 0 && item.result.status.max != 0 &&
		resp.StatusCode >= item.result.status.min && resp.StatusCode <= item.result.status.max {
		return true, ""
	}
	if item.result.body != "" && string(webBody) == item.result.body {
		return true, ""
	}
	return false, ""
}
