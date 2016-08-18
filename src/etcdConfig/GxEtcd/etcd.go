package GxEtcd

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	etcdClient "github.com/coreos/etcd/client"
)

type PreloadFN func(etcdClient.KeysAPI, Options) error

var (
	Preloads map[string]PreloadFN // 预加载配置列表
)

// etcd地址
var (
	DEFAULT_ETCD = "http://172.17.42.1:2379"
)

func init() {
	Preloads = make(map[string]PreloadFN)
	if env := os.Getenv("ETCD_HOST"); env != "" {
		DEFAULT_ETCD = env
	}
}

func Initialize(opts ...Option) {
	config := etcdClient.Config{
		Endpoints: []string{"http://127.0.0.1:2379"},
	}

	var options Options
	for _, o := range opts {
		o(&options)
	}

	if options.Timeout == 0 {
		options.Timeout = etcdClient.DefaultRequestTimeout
	}

	if options.Secure || options.TLSConfig != nil {
		tlsConfig := options.TLSConfig
		if tlsConfig == nil {
			tlsConfig = &tls.Config{
				InsecureSkipVerify: true,
			}
		}

		// for InsecureSkipVerify
		t := &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			Dial: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).Dial,
			TLSHandshakeTimeout: 10 * time.Second,
			TLSClientConfig:     tlsConfig,
		}

		runtime.SetFinalizer(&t, func(tr **http.Transport) {
			(*tr).CloseIdleConnections()
		})

		config.Transport = t

		// default secure address
		config.Endpoints = []string{"https://127.0.0.1:2379"}
	}

	var cAddrs []string

	for _, addr := range options.Addrs {
		if len(addr) == 0 {
			continue
		}

		if options.Secure {
			// replace http:// with https:// if its there
			addr = strings.Replace(addr, "http://", "https://", 1)

			// has the prefix? no... ok add it
			if !strings.HasPrefix(addr, "https://") {
				addr = "https://" + addr
			}
		}

		cAddrs = append(cAddrs, addr)
	}

	// if we got addrs then we'll update
	if len(cAddrs) > 0 {
		config.Endpoints = cAddrs
	}

	c, _ := etcdClient.New(config)
	client := etcdClient.NewKeysAPI(c)
	for k, fn := range Preloads {
		err := fn(client, options)
		if err != nil {
			fmt.Errorf(">>>> 预载入%s时出错: %v", k, err)
		} else {
			fmt.Printf(">>>> 预载入%s完成", k)
		}
	}
}
