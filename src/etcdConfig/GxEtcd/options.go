package GxEtcd

import (
	"crypto/tls"
	"encoding/json"
	"time"

	"golang.org/x/net/context"
)

type Option func(*Options)

type Options struct {
	Addrs     []string
	Timeout   time.Duration
	Secure    bool
	TLSConfig *tls.Config

	// Other options for implementations of the interface
	// can be stored in a context
	Context context.Context
}

// Addrs is the registry addresses to use
func Addrs(addrs ...string) Option {
	return func(o *Options) {
		o.Addrs = addrs
	}
}

func Timeout(t time.Duration) Option {
	return func(o *Options) {
		o.Timeout = t
	}
}

// Secure communication with the registry
func Secure(b bool) Option {
	return func(o *Options) {
		o.Secure = b
	}
}

// Specify TLS Config
func TLSConfig(t *tls.Config) Option {
	return func(o *Options) {
		o.TLSConfig = t
	}
}

func Decode(ds string, s interface{}) error {
	return json.Unmarshal([]byte(ds), &s)
}
