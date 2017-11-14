package proxy

import (
	"math/rand"
	"sync"
	"time"

	"github.com/go-playground/log"
	"github.com/parnurzeal/gorequest"
	"github.com/yang-f/ProxyPool/models"
	"github.com/yang-f/ProxyPool/storage"
	"github.com/yang-f/ProxyPool/util"
)

type Proxy struct {
	Storage *storage.Storage
}

func NewProxy(conf *util.Config) *Proxy {
	return &Proxy{Storage: storage.NewStorage(conf)}
}

// CheckProxy .
func (p *Proxy) CheckProxy(ip *models.IP) {
	if p.CheckIP(ip) {
		p.ProxyAdd(ip)
	}
}

// CheckIP is to check the ip work or not
func (p *Proxy) CheckIP(ip *models.IP) bool {
	pollURL := "http://httpbin.org/get"
	resp, _, errs := gorequest.New().Proxy("http://" + ip.Data).Get(pollURL).End()
	if errs != nil {
		return false
	}
	if resp.StatusCode == 200 {
		return true
	}
	return false
}

// CheckProxyDB to check the ip in DB
func (p *Proxy) CheckProxyDB() {
	x := p.Storage.Count()
	log.Println("Before check, DB has:", x, "records.")
	ips, err := p.Storage.GetAll()
	if err != nil {
		log.Println(err.Error())
		return
	}

	workersCount := 10
	workers := make(chan int, workersCount)
	for i := 0; i < workersCount; i++ {
		workers <- 1
	}

	var wg sync.WaitGroup
	for _, v := range ips {
		wg.Add(1)
		<-workers
		go func(v *models.IP) {
			defer func() {
				workers <- 1
				wg.Done()
			}()
			if !p.CheckIP(v) {
				p.ProxyDel(v)
			}
		}(v)
	}
	wg.Wait()
	x = p.Storage.Count()
	log.Println("After check, DB has:", x, "records.")
}

// ProxyRandom .
func (p *Proxy) ProxyRandom() (ip *models.IP) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	ips, _ := p.Storage.GetAll()
	x := len(ips)

	return ips[r.Intn(x)]
}

// ProxyFind .
func (p *Proxy) ProxyFind(value string) (ip *models.IP) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	ips, _ := p.Storage.FindAll(value)
	x := len(ips)

	return ips[r.Intn(x)]
}

// ProxyAdd .
func (p *Proxy) ProxyAdd(ip *models.IP) {
	_, err := p.Storage.GetOne(ip.Data)
	if err != nil {
		p.Storage.Create(ip)
	}
}

// ProxyDel .
func (p *Proxy) ProxyDel(ip *models.IP) {
	if err := p.Storage.Delete(ip); err != nil {
		log.Println(err.Error())
	}
}
