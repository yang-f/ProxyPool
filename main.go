package main

import (
	"runtime"

	"github.com/go-playground/log"
	"github.com/go-playground/log/handlers/console"
	"github.com/yang-f/ProxyPool/manager"
	"github.com/yang-f/ProxyPool/util"
)

func main() {

	cLog := console.New()

	log.RegisterHandler(cLog, log.AllLevels...)
	runtime.GOMAXPROCS(runtime.NumCPU())

	manager, err := manager.NewProxyManager(util.NewConfig())
	if err != nil {
		log.Fatal(err)
		return
	}
	go manager.Run()
}
