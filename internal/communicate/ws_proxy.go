package communicate

import (
	"context"
	"crypto/tls"
	"github.com/gorilla/websocket"
	"golang.org/x/net/proxy"
	"net"
	"net/http"
	"net/url"
)

func (c *Communicate) applyWebSocketProxyIfSet(dialer *websocket.Dialer) {
	if c.opt.HttpProxy != "" {
		proxyUrl, _ := url.Parse(c.opt.HttpProxy)
		dialer.Proxy = http.ProxyURL(proxyUrl)
	}
	if c.opt.Socket5Proxy != "" {
		auth := &proxy.Auth{User: c.opt.Socket5ProxyUser, Password: c.opt.Socket5ProxyPass}
		socket5ProxyDialer, _ := proxy.SOCKS5("tcp", c.opt.Socket5Proxy, auth, proxy.Direct)
		dialContext := func(ctx context.Context, network, address string) (net.Conn, error) {
			return socket5ProxyDialer.Dial(network, address)
		}
		dialer.NetDialContext = dialContext
	}
	if c.opt.IgnoreSSL {
		dialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
}
