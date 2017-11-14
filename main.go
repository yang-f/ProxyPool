package main

import (
	"runtime"
	"sync"
	"time"

	"github.com/go-playground/log"
	"github.com/go-playground/log/handlers/console"
	"github.com/yang-f/ProxyPool/getter"
	"github.com/yang-f/ProxyPool/models"
	"github.com/yang-f/ProxyPool/proxy"
	"github.com/yang-f/ProxyPool/util"
)

func main() {

	cLog := console.New()

	log.RegisterHandler(cLog, log.AllLevels...)

	runtime.GOMAXPROCS(runtime.NumCPU())
	ipChan := make(chan *models.IP, 2000)
	proxy := proxy.NewProxy(util.NewConfig())

	// Check the IPs in channel
	for i := 0; i < 30; i++ {
		go func() {
			for {
				proxy.CheckProxy(<-ipChan)
			}
		}()
	}

	// Start getters to scraper IP and put it in channel
	for {
		x := proxy.Storage.Count()
		log.Printf("Chan: %v, IP: %v\n", len(ipChan), x)
		go func() {
			proxy.CheckProxyDB()
		}()
		if len(ipChan) < 100 {
			go run(ipChan)
		}
		time.Sleep(30 * time.Minute)
	}
}

func run(ipChan chan<- *models.IP) {
	var wg sync.WaitGroup
	funs := []func() []*models.IP{
		getter.Data5u,
		getter.IP66,
		//getter.KDL,
		getter.GBJ,
		//getter.Xici,
		//getter.XDL,
		getter.IP181,
		//getter.YDL,		//失效的采集脚本，用作系统容错实验
		getter.PLP,
	}
	for _, f := range funs {
		wg.Add(1)
		go func(f func() []*models.IP) {
			temp := f()
			for _, v := range temp {
				ipChan <- v
			}
			wg.Done()
		}(f)
	}
	wg.Wait()
	log.Println("All getters finished.")
}
