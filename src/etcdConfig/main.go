package main

import (
	"fmt"

	confTest "etcdConfig/Conf/global"
	gxetcd "etcdConfig/GxEtcd"
)

func main() {
	gxetcd.Initialize(gxetcd.Addrs(gxetcd.DEFAULT_ETCD))
	id := "1"
	conf, ok := confTest.Get(id)
	if ok {
		fmt.Printf("邮件内容:%v", conf)
	} else {
		fmt.Errorf("邮件模板id找不到:%s", id)
	}
}
