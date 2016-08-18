package Conf

import (
	"errors"
	"fmt"
	"sync"

	etcd "github.com/coreos/etcd/client"
	"golang.org/x/net/context"
)

// 监听器
type EtcdWatcher struct {
	ctx     context.Context
	once    sync.Once
	stop    chan bool
	w       etcd.Watcher
	keyPath string
}

// 数据结构
type EtcdWatcherResp struct {
	Id     string
	Action string
	V      string
}

type ConfigWatcher interface {
	Delete(resp *EtcdWatcherResp) error
	Update(resp *EtcdWatcherResp) error
}

// 注意：recursive字段,要注意区分，原则是：所有父目录都不要有配置条目
// --true: keyPath目录下没有子目录，只有数据配置条数
// --false: keyPath目录下有子目录
func NewEtcdWatcher(kapi etcd.KeysAPI, keyPath string, recursive bool) (*EtcdWatcher, error) {
	var once sync.Once
	ctx, cancel := context.WithCancel(context.Background())
	stop := make(chan bool, 1)

	go func() {
		<-stop
		cancel()
	}()

	return &EtcdWatcher{
		ctx:     ctx,
		w:       kapi.Watcher(keyPath, &etcd.WatcherOptions{AfterIndex: 0, Recursive: recursive}),
		once:    once,
		stop:    stop,
		keyPath: keyPath,
	}, nil
}

func (ew *EtcdWatcher) Next(watcher ConfigWatcher) error {
	rsp, err := ew.w.Next(ew.ctx)
	if err != nil && ew.ctx.Err() != nil {
		return err
	}
	if rsp.Node.Dir {
		return errors.New("配置项竟然是一个目录")
	}
	var js string
	if _, err := fmt.Sscanf(rsp.Node.Key, (ew.getKeyPath() + "%s"), &js); err != nil {
		return errors.New("无法解析文本编号")
	}
	id := js[:len(js)-5]

	switch rsp.Action {
	case "delete":
		resp := &EtcdWatcherResp{
			Id:     id,
			Action: "delete",
		}
		return watcher.Delete(resp)
	case "create", "set", "update":
		resp := &EtcdWatcherResp{
			Id:     id,
			Action: rsp.Action,
			V:      rsp.Node.Value,
		}
		return watcher.Update(resp)
	default:
		return errors.New("未知的事件类型:" + rsp.Action)
	}
}

func (ew *EtcdWatcher) getKeyPath() string {
	return ew.keyPath
}

func (ew *EtcdWatcher) Stop() {
	ew.once.Do(func() {
		ew.stop <- true
	})
}
