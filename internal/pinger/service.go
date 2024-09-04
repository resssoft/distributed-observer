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

//TODO: ping history, count of ping, result history, trigger by good and bad result (+antispam -> notice for change status)

const queueLimit = 10000

type Data struct {
	ItemsGroup []ItemsGroup `json:"items_group"`
	dispatcher *mediator.Dispatcher
	logger     *logger.Logger
	settings   services.Settings
	queue      chan Item
}

func New(dispatcher *mediator.Dispatcher, logger *logger.Logger, settings services.Settings) *Data {
	logger = logger.With("service", "pinger")
	return &Data{
		dispatcher: dispatcher,
		logger:     logger,
		settings:   settings,
		queue:      make(chan Item, queueLimit),
		ItemsGroup: make([]ItemsGroup, 0),
	}
}

func (d *Data) Append(t time.Duration, items ...Item) []ItemsGroup {
	d.ItemsGroup = append(d.ItemsGroup, ItemsGroup{
		Timeout: t,
		Items:   items,
	})
	return d.ItemsGroup
}

func (d *Data) Start(ctx context.Context) {
	d.logger.Info(ctx, "Start Pinger")
	d.Append(time.Minute*25,
		//Item{Url: "http://127.0.0.2/"}.CheckStatus(),
		Item{Url: "188.21.21.21"}.CheckPing(
			d.settings.GetValueSeconds("OBSERVER_PINGER_PING_TIMEOUT_SEC", 5),
			d.settings.GetValueInt("OBSERVER_PINGER_PING_REPEAT", 3)),
		Item{Url: "https://example.com/"}.CheckStatus(),
		Item{Url: "https://no-exist-domain-243524523452345234524524.com/"}.CheckStatus(),
	)

	d.Append(time.Second*55,
		Item{Url: "127.0.0.1"}.CheckPing(
			d.settings.GetValueSeconds("OBSERVER_PINGER_PING_TIMEOUT_SEC", 5),
			d.settings.GetValueInt("OBSERVER_PINGER_PING_REPEAT", 3)))
	for i := 0; i < runtime.NumCPU(); i++ {
		go d.Receiver(ctx)
	}
	go d.Sender(ctx)
}

func (d *Data) Sender(ctx context.Context) {
	for _, ig := range d.ItemsGroup {
		igCopy := ig
		go func() {
			for {
				select {
				case <-time.After(igCopy.Timeout): //d.settings.AfterSeconds("OBSERVER_PINGER_SENDER_TIMEOUT_SEC", 15):
					println("")
					d.logger.Info(ctx, "send by timeout")
					d.Send(ctx, igCopy.Items)
				}
			}
		}()
	}
}

func (d *Data) Send(ctx context.Context, items []Item) {
	for _, item := range items {
		d.logger.Debug(ctx, "sending item", "item", item)
		d.queue <- item
	}
}

func (d *Data) Receiver(ctx context.Context) {
	for item := range d.queue {
		host := getHost(item.Url)
		d.logger.Info(ctx, "receiving item", "host", host)
		//println("check item", item.url, host)
		if host != "" {
			if item.Result.Ping != nil {
				state, err := d.ping(host,
					item.Result.Ping.Repeat,
					item.Result.Ping.Timeout)
				d.logger.Info(ctx, fmt.Sprintf("Received [%s] ping for url [%s] result %s",
					defaults.Str(item.Name, item.Url),
					item.Url,
					fmt.Sprintf("%v err: %v", state, err),
				))
			} else {
				state, err := d.web(ctx, item)
				d.logger.Info(ctx, fmt.Sprintf("Received [%s] web for url [%s] result %s",
					defaults.Str(item.Name, item.Url),
					item.Url,
					fmt.Sprintf("%v err: %v", state, err),
				))
			}
		} else {
			d.logger.Info(ctx, fmt.Sprintf("EMPTY HOST [%s] is empty", item.Url), "item", item)
		}
	}
}

func (d *Data) ping(address string, repeat int, timeout time.Duration) (bool, error) {
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

func (d *Data) web(ctx context.Context, item Item) (bool, string) {
	client := &http.Client{}
	if item.Request.Proxy != nil {
		proxyURL, err := url.Parse(item.Request.Proxy.Host)
		if err != nil {
			return false, "proxy url invalid: " + err.Error()
		}
		if item.Request.Proxy.User != "" {
			//proxyURL.Host = address
			proxyURL.User = url.UserPassword(item.Request.Proxy.User, item.Request.Proxy.Pass)
		}
		proxy := http.ProxyURL(proxyURL)
		transport := &http.Transport{Proxy: proxy}
		client.Transport = transport
		//clientWithProxy, err := transport.DialContext(ctx, "tcp", item.Request.Proxy.Host)
		//if nil != err {
		//	log.Fatalf("Dial: %v", err)
		//}
		//client = &clientWithProxy
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
	if item.Result.Status.Code != 0 && resp.StatusCode == item.Result.Status.Code {
		return true, ""
	}
	if item.Result.Status.Min != 0 && item.Result.Status.Max != 0 &&
		resp.StatusCode >= item.Result.Status.Min && resp.StatusCode <= item.Result.Status.Max {
		return true, ""
	}
	if item.Result.Body != "" && string(webBody) == item.Result.Body {
		return true, ""
	}
	return false, ""
}
