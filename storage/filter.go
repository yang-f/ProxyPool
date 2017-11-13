package storage

import (
	"math/rand"
	"sync"
	"time"

	"github.com/go-playground/log"
	"github.com/parnurzeal/gorequest"
	"github.com/yang-f/ProxyPool/models"
	"github.com/yang-f/ProxyPool/util"
)

type Filter struct {
	Storage *Storage
}

func NewFilter(conf *util.Config) *Filter {
	return &Filter{Storage: NewStorage(conf)}
}

// CheckProxy .
func (f *Filter) CheckProxy(ip *models.IP) {
	if f.CheckIP(ip) {
		f.ProxyAdd(ip)
	}
}

// CheckIP is to check the ip work or not
func (f *Filter) CheckIP(ip *models.IP) bool {
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
func (f *Filter) CheckProxyDB() {
	x := f.Storage.Count()
	log.Println("Before check, DB has:", x, "records.")
	ips, err := f.Storage.GetAll()
	if err != nil {
		log.Println(err.Error())
		return
	}
	var wg sync.WaitGroup
	for _, v := range ips {
		wg.Add(1)
		go func(v *models.IP) {
			if !f.CheckIP(v) {
				f.ProxyDel(v)
			}
			wg.Done()
		}(v)
	}
	wg.Wait()
	x = f.Storage.Count()
	log.Println("After check, DB has:", x, "records.")
}

// ProxyRandom .
func (f *Filter) ProxyRandom() (ip *models.IP) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	ips, _ := f.Storage.GetAll()
	x := len(ips)

	return ips[r.Intn(x)]
}

// ProxyFind .
func (f *Filter) ProxyFind(value string) (ip *models.IP) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	ips, _ := f.Storage.FindAll(value)
	x := len(ips)

	return ips[r.Intn(x)]
}

// ProxyAdd .
func (f *Filter) ProxyAdd(ip *models.IP) {
	_, err := f.Storage.GetOne(ip.Data)
	if err != nil {
		f.Storage.Create(ip)
	}
}

// ProxyDel .
func (f *Filter) ProxyDel(ip *models.IP) {
	if err := f.Storage.Delete(ip); err != nil {
		log.Println(err.Error())
	}
}
