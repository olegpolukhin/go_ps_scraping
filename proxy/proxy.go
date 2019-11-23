package proxy

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/spf13/viper"

	"golang.org/x/net/proxy"
)

// NewProxyTransport proxy transport
func NewProxyTransport() *http.Transport {
	useAuth := true
	if viper.GetString("PROXY_USER") == "" || viper.GetString("PROXY_PASSWORD") == "" {
		useAuth = false
	}

	var proxyAuth *proxy.Auth
	if useAuth {
		proxyAuth = &proxy.Auth{
			User:     viper.GetString("PROXY_USER"),
			Password: viper.GetString("PROXY_PASSWORD"),
		}
	}

	tr := &http.Transport{
		DialContext: func(_ context.Context, network, addr string) (net.Conn, error) {
			socksDialer, err := proxy.SOCKS5(
				"tcp",
				fmt.Sprintf("%s:%d",
					viper.GetString("PROXY_HOST"),
					viper.GetInt("PROXY_PORT"),
				),
				proxyAuth,
				proxy.Direct,
			)
			if err != nil {
				return nil, err
			}

			return socksDialer.Dial(network, addr)
		},
	}

	return tr
}
