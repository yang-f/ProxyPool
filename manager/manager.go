package manager

import (
	"math/rand"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/go-playground/log"
	"github.com/parnurzeal/gorequest"
	"github.com/pkg/errors"
	"github.com/yang-f/ProxyPool/getter"
	"github.com/yang-f/ProxyPool/models"
	"github.com/yang-f/ProxyPool/storage"
	"github.com/yang-f/ProxyPool/util"
)

type ProxyManager struct {
	Storage *storage.Storage
	Cookie  string
	IP      *models.IP
}

func NewProxyManager(conf *util.Config) (*ProxyManager, error) {
	storage, err := storage.NewStorage(conf)
	if err != nil {
		return nil, err
	}
	return &ProxyManager{
		Storage: storage,
		Cookie:  "",
	}, nil
}

// CheckProxy .
func (m *ProxyManager) CheckProxy(ip *models.IP) {
	if m.CheckIP(ip) {
		m.ProxyAdd(ip)
	}
}

// CheckIP is to check the ip work or not
func (m *ProxyManager) CheckIP(ip *models.IP) bool {
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
func (m *ProxyManager) CheckProxyDB() error {
	x := m.Storage.Count()
	log.Println("Before check, DB has:", x, "records.")
	ips, err := m.Storage.GetAll()
	if err != nil {
		log.Println(err.Error())
		return err
	}
	for _, v := range ips {
		if !m.CheckIP(v) {
			err := m.ProxyDel(v)
			if err != nil {
				return err
			}
		}
	}
	x = m.Storage.Count()
	log.Println("After check, DB has:", x, "records.")
	return nil
}

// ProxyRandom .
func (m *ProxyManager) ProxyRandom() (ip *models.IP, err error) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	ips, err := m.Storage.GetAll()
	if err != nil {
		return nil, err
	}
	x := len(ips)
	if x == 0 {
		return nil, errors.Errorf("no ip found")
	}
	return ips[r.Intn(x)], nil
}

// ProxyFind .
func (m *ProxyManager) ProxyFind(value string) (ip *models.IP, err error) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	ips, err := m.Storage.FindAll(value)
	if err != nil {
		return nil, err
	}
	x := len(ips)
	if x == 0 {
		return nil, errors.Errorf("no ip found")
	}
	return ips[r.Intn(x)], nil
}

// ProxyAdd .
func (m *ProxyManager) ProxyAdd(ip *models.IP) {
	_, err := m.Storage.GetOne(ip.Data)
	if err != nil {
		m.Storage.Create(ip)
	}
}

// ProxyDel .
func (m *ProxyManager) ProxyDel(ip *models.IP) error {
	if err := m.Storage.Delete(ip); err != nil {
		return err
	}
	return nil
}

func (m *ProxyManager) SetIP() error {
	ip, err := m.ProxyRandom()
	if err != nil {
		return errors.Errorf("no ip found")
	}
	m.IP = ip
	return nil
}

func (m *ProxyManager) SetCookie() error {
	targetURL := "http://www.target.com"
	ip, err := m.ProxyRandom()
	if err != nil {
		return errors.Errorf("no ip found")
	}
	m.IP = ip
	url, err := url.Parse("http://" + ip.Data)
	if err != nil {
		return err
	}
	proxy := http.ProxyURL(url)
	client := &http.Client{
		Timeout: time.Second * 30,
		Transport: &http.Transport{
			Proxy: proxy,
		},
	}

	req, err := http.NewRequest(http.MethodGet, targetURL, nil)
	if err != nil {
		return err
	}

	req.Header.Add("Cache-Control", "max-age=0")
	req.Header.Add("Host", "www.target.com")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36")
	httpResp, err := client.Do(req)
	if err != nil {
		return err
	}
	if len(httpResp.Header["Set-Cookie"]) > 0 {
		log.Printf("set-cookie:%s", httpResp.Header["Set-Cookie"][0])
		m.Cookie = httpResp.Header["Set-Cookie"][0]
		return nil
	}
	return errors.Errorf("no cookie")

}

func (m *ProxyManager) Run() {
	ipChan := make(chan *models.IP, 2000)
	for i := 0; i < 30; i++ {
		go func() {
			for {
				m.ProxyAdd(<-ipChan)
			}
		}()
	}
	count := 0
	for {
		count++
		go m.SetIP()
		if count%40000 == 0 {
			x := m.Storage.Count()
			log.Printf("Chan: %v, IP: %v\n", len(ipChan), x)
			go m.SetCookie()
			if m.Storage.Count() < 500 {
				go runGetter(ipChan)
			}
			count = 0
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func runGetter(ipChan chan<- *models.IP) {
	var wg sync.WaitGroup
	funs := []func() []*models.IP{
		getter.Zhima,
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
