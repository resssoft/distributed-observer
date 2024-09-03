package pinger

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"runtime"
	"time"

	pinger "github.com/go-ping/ping"

	"observer/internal/domain/services"
	"observer/internal/logger"
	"observer/pkg/defaults"
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
	logger = logger.With("service", "pinger")
	return &Data{
		dispatcher: dispatcher,
		logger:     logger,
		settings:   settings,
		queue:      make(chan Item, queueLimit),
		Items:      items,
	}
}

func (d *Data) Start(ctx context.Context) {
	d.logger.Info(ctx, "Start Pinger")
	d.Items = append(d.Items,
		Item{url: "127.0.0.1"}.CheckPing(
			d.settings.GetValueSeconds("OBSERVER_PINGER_PING_TIMEOUT_SEC", 5),
			d.settings.GetValueInt("OBSERVER_PINGER_PING_REPEAT", 3)),
		//Item{url: "http://127.0.0.2/"}.CheckStatus(),
		Item{url: "188.21.21.21"}.CheckPing(
			d.settings.GetValueSeconds("OBSERVER_PINGER_PING_TIMEOUT_SEC", 5),
			d.settings.GetValueInt("OBSERVER_PINGER_PING_REPEAT", 3)),
		Item{url: "https://example.com/"}.CheckStatus(),
		Item{url: "https://no-exist-domain-243524523452345234524524.com/"}.CheckStatus(),
	)
	d.Send(ctx)
	for i := 0; i < runtime.NumCPU(); i++ {
		go d.Receiver(ctx)
	}
	go d.Sender(ctx)
}

func (d *Data) Sender(ctx context.Context) {
	//TODO: group for different timeout
	for {
		select {
		case <-d.settings.AfterSeconds("OBSERVER_PINGER_SENDER_TIMEOUT_SEC", 15):
			println("")
			d.logger.Info(ctx, "send by timeout")
			d.Send(ctx)
		}
	}
}

func (d *Data) Send(ctx context.Context) {
	for _, item := range d.Items {
		d.logger.Debug(ctx, "sending item", "item", item)
		d.queue <- item
	}
}

func (d *Data) Receiver(ctx context.Context) {
	for item := range d.queue {
		host := getHost(item.url)
		d.logger.Info(ctx, "receiving item", "host", host)
		//println("check item", item.url, host)
		if host != "" {
			if item.result.ping != nil {
				//println("!!!!!!!check ping", item.url, host)
				state, err := d.ping(host,
					item.result.ping.repeat,
					item.result.ping.timeout)
				d.logger.Info(ctx, fmt.Sprintf("Received [%s] ping for url [%s] result %s",
					defaults.Str(item.name, item.url),
					item.url,
					fmt.Sprintf("%v err: %v", state, err),
				))
			} else {
				//println("!!!!!!!check web", item.url, host)
				state, err := d.web(item)
				d.logger.Info(ctx, fmt.Sprintf("Received [%s] ping for url [%s] result %s",
					defaults.Str(item.name, item.url),
					item.url,
					fmt.Sprintf("%v err: %v", state, err),
				))
			}
		} else {
			d.logger.Info(ctx, fmt.Sprintf("EMPTY HOST [%s] is empty", item.url), "item", item)
		}
	}
}

func (d Data) ping(address string, repeat int, timeout time.Duration) (bool, error) {
	pinger, err := pinger.NewPinger(address)
	if err != nil {
		return false, err
	}
	pinger.Count = repeat
	pinger.Timeout = timeout
	err = pinger.Run()
	if err != nil {
		return false, err
	}
	stats := pinger.Statistics() // get send/receive/rtt stats
	if stats == nil {
		return false, nil
	}
	if stats.PacketsRecv != repeat {
		return false, nil
	}
	//d.logger.Info(context.Background(), "ping address", "address", address, "stats", pinger.Statistics())
	return true, err
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
	if item.proxy != "" {
		proxyURL, _ := url.Parse(item.proxy)
		proxy := http.ProxyURL(proxyURL)
		transport := &http.Transport{Proxy: proxy}
		client.Transport = transport
	}
	request, err := item.buildRequest()
	if err != nil {
		return false, "build request err: " + err.Error()
	}
	resp, err := client.Do(request)
	if err != nil {
		return false, "request err: " + err.Error()
	}
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
