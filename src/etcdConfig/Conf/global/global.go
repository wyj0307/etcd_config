package global

//  测试模块

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	. "etcdConfig/Conf"
	gxetcd "etcdConfig/GxEtcd"
	etcd "github.com/coreos/etcd/client"
	"golang.org/x/net/context"
)

const (
	glKeyPath = "/config/wyj_test/"
)

type I18N struct {
	ALL_ALL string `json:"all_all"`
	ZH_CN   string `json:"zh_cn"`
}

type SDMailTemplate struct {
	MailID    string           `json:"mail_id"`
	Title     I18N             `json:"title_i18n"`
	Content   I18N             `json:"text_i18n"`
	ValidTime int32            `json:"valid"`
	Resources map[string]int32 `json:"resources"`
}

type GlobalConf struct {
	_client  etcd.KeysAPI
	_opts    gxetcd.Options
	_watcher *EtcdWatcher
	_mu      sync.RWMutex
	_cached  map[string]SDMailTemplate
}

var globalConf GlobalConf

// delete
func (conf *GlobalConf) Delete(resp *EtcdWatcherResp) error {
	conf._mu.Lock()
	delete(conf._cached, resp.Id)
	conf._mu.Unlock()
	fmt.Println("删除[%s]配置数据", glKeyPath)
	return nil
}

// update
func (conf *GlobalConf) Update(resp *EtcdWatcherResp) error {
	var value SDMailTemplate
	if gxetcd.Decode(resp.V, &value) == nil {
		conf._mu.Lock()
		conf._cached[resp.Id] = value
		conf._mu.Unlock()
		fmt.Printf("更新[%s]配置数据：%v", glKeyPath, resp.V)
		return nil
	} else {
		return errors.New(fmt.Sprintf("更新[%s]配置数据出错：%v", glKeyPath, resp.V))
	}
}

// ------------------------------------------------------------------------------------------------

func init() {
	gxetcd.Preloads["global_config"] = globalConf.Preload
	globalConf._cached = make(map[string]SDMailTemplate)
}

// 预加载数据函数
func (conf *GlobalConf) Preload(kapi etcd.KeysAPI, opts gxetcd.Options) error {
	conf._client = kapi
	conf._opts = opts

	ctx, cancel := context.WithTimeout(context.Background(), opts.Timeout)
	defer cancel()

	rsp, err := conf._client.Get(ctx, glKeyPath, &etcd.GetOptions{Recursive: true, Sort: true})
	if err != nil && !strings.HasPrefix(err.Error(), "100: Key not found") {
		return err
	}
	if rsp != nil {
		for _, node := range rsp.Node.Nodes {
			mail := SDMailTemplate{}
			gxetcd.Decode(node.Value, &mail)
			var js string
			if _, err := fmt.Sscanf(node.Key, (glKeyPath + "%s"), &js); err != nil {
				return errors.New(fmt.Sprintf("无法获得[%s]的配置", glKeyPath))
			}
			js = js[:len(js)-5]
			conf._cached[js] = mail
		}
		fmt.Printf(">> 载入[%s]配置共%d条\n", glKeyPath, len(conf._cached))
	}

	conf._watcher, _ = NewEtcdWatcher(kapi, glKeyPath, true)
	go func() {
		for {
			err := conf._watcher.Next(&globalConf)
			if err != nil {
				fmt.Errorf("监听[%s]配置时出错: %v", glKeyPath, err)
				continue
			}
			time.Sleep(time.Second)
		}
	}()

	return nil
}

// 判断邮件配置数据是否存在
func Exists(tpl_id string) bool {
	globalConf._mu.RLock()
	defer globalConf._mu.RUnlock()

	_, exists := globalConf._cached[tpl_id]
	return exists
}

// 获取邮件配置数据
func Get(tpl_id string) (SDMailTemplate, bool) {
	globalConf._mu.RLock()
	defer globalConf._mu.RUnlock()

	v, ok := globalConf._cached[tpl_id]
	return v, ok
}
