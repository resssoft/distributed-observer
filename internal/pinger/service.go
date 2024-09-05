package pinger

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"sync"
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
	history    History
	mutex      *sync.Mutex
}

func New(dispatcher *mediator.Dispatcher, logger *logger.Logger, settings services.Settings) *Data {
	logger = logger.With("service", "pinger")
	return &Data{
		dispatcher: dispatcher,
		logger:     logger,
		settings:   settings,
		queue:      make(chan Item, queueLimit),
		ItemsGroup: make([]ItemsGroup, 0),
		history: History{
			Requests: make(map[time.Time]Request),
		},
		mutex: &sync.Mutex{},
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
		PingItem("188.21.21.21", d.settings.GetValueSeconds("OBSERVER_PINGER_PING_TIMEOUT_SEC", 5), d.settings.GetValueInt("OBSERVER_PINGER_PING_REPEAT", 3)), //some not pinged
		CheckStatusItem("https://google.com/"),                                   //domain exist, server state successful
		CheckStatusItem("https://dima.com/"),                                     //domain exist, server state fail
		CheckStatusItem("https://no-exist-domain-243524523452345234524524.com/"), //domain not exist, server state fail
	)

	d.Append(time.Second*55,
		PingItem("127.0.0.2", d.settings.GetValueSeconds("OBSERVER_PINGER_PING_TIMEOUT_SEC", 5), d.settings.GetValueInt("OBSERVER_PINGER_PING_REPEAT", 3)), //some pinged
	)
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
		d.logger.Info(ctx, "receiving item", "Name", item.Name)
		if item.Request.Ping != "" {
			state, err := d.ping(item.Request.Ping,
				item.Request.Repeat,
				item.Request.Timeout)
			d.logger.Info(ctx, fmt.Sprintf("Received [%s] ping for url [%s] result %s",
				defaults.Str(item.Name, item.Request.Url),
				item.Request.Url,
				fmt.Sprintf("%v err: %v", state, err),
			))
			continue
		}
		host := getHost(item.Request.Url)
		if host != "" {
			result := d.web(ctx, item)
			d.logger.Info(ctx, fmt.Sprintf("Received [%s] web for url [%s] result %s",
				defaults.Str(item.Name, item.Request.Url),
				item.Request.Url,
				fmt.Sprintf("%v err: %v", result.StatusCode, result.Error),
			))
			continue
		} else {
			d.logger.Info(ctx, fmt.Sprintf("EMPTY HOST [%s] is empty", item.Request.Url), "item", item)
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

func (d *Data) web(ctx context.Context, item Item) ResponseResult {
	result := ResponseResult{}
	client := &http.Client{}
	if item.Request.Proxy != nil {
		proxyURL, err := url.Parse(item.Request.Proxy.Host)
		if err != nil {
			return result.WithErr("Parse proxy url err: %s", err)
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
		return result.WithErr("build request err: %s", err)
	}
	resp, err := client.Do(request)
	if err != nil {
		return result.WithErr("request err: %s", err)
	}
	if resp == nil {
		return result.SetErr("empty response")
	}
	if item.Request.Response.Status.Code != 0 && resp.StatusCode == item.Request.Response.Status.Code {
		return result
	}
	if item.Request.Response.Status.Min != 0 && item.Request.Response.Status.Max != 0 &&
		resp.StatusCode >= item.Request.Response.Status.Min && resp.StatusCode <= item.Request.Response.Status.Max {
		return result
	}
	result.StatusCode = resp.StatusCode
	webBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return result.WithErr("read body err: %s", err)
	}
	result.Body = string(webBody)
	if item.Request.Response.Body != nil {
		if item.Request.Response.Body.Full == result.Body {
			return result
		}
		if strings.Contains(result.Body, item.Request.Response.Body.Contain) {
			return result
		}
		if item.Request.Response.Body.Grep != nil {
			//TODO: grep
			return result
		}
	}
	return result
}
