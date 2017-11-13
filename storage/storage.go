package storage

import (
	"encoding/json"
	"github.com/fzzy/radix/redis"
	"github.com/pkg/errors"
	"github.com/yang-f/ProxyPool/models"
	"github.com/yang-f/ProxyPool/util"
	"time"
)

// Storage struct is used for storeing persistent data of alerts
type Storage struct {
	Config *util.Config
}

// NewStorage creates and returns new Storage instance
func NewStorage(conf *util.Config) *Storage {
	return &Storage{Config: conf}
}

// GetDBSession returns a new connection from the pool
func (s *Storage) GetDBSession() *redis.Client {
	client, _ := redis.DialTimeout("tcp", s.Config.Redis.Addr, time.Duration(10)*time.Second)
	return client
}

// Create insert new item
func (s *Storage) Create(ip *models.IP) error {
	ses := s.GetDBSession()
	defer ses.Close()
	value, err := json.Marshal(ip)
	if err != nil {
		return err
	}
	r := ses.Cmd("HMSET", s.Config.Redis.Key, ip.Data, string(value))
	if r.Err != nil {
		return r.Err
	}
	return nil
}

// GetOne Finds and returns one data from storage
func (s *Storage) GetOne(value string) (*models.IP, error) {
	ses := s.GetDBSession()
	defer ses.Close()
	ls, err := ses.Cmd("HGET", s.Config.Redis.Key, value).List()
	if len(ls) == 0 {
		return nil, errors.New("no ip found.")
	}
	t := models.NewIP()
	err = json.Unmarshal([]byte(ls[0]), &t)
	if err != nil {
		return nil, err
	}
	return t, nil
}

// Count all collections
func (s *Storage) Count() int {
	ses := s.GetDBSession()
	defer ses.Close()
	r := ses.Cmd("HLEN", s.Config.Redis.Key)
	num, err := r.Int()
	if err != nil {
		num = 0
	}
	return num
}

// Delete .
func (s *Storage) Delete(ip *models.IP) error {
	ses := s.GetDBSession()
	defer ses.Close()
	r := ses.Cmd("HDEL", s.Config.Redis.Key, ip.Data)
	if r.Err != nil {
		return r.Err
	}
	return nil
}

// Update .
func (s *Storage) Update(ip *models.IP) error {
	ses := s.GetDBSession()
	defer ses.Close()
	value, err := json.Marshal(ip)
	if err != nil {
		return err
	}
	r := ses.Cmd("HMSET", s.Config.Redis.Key, ip.Data, string(value))
	if r.Err != nil {
		return r.Err
	}
	return nil
}

// GetAll .
func (s *Storage) GetAll() ([]*models.IP, error) {
	ses := s.GetDBSession()
	defer ses.Close()
	var ips []*models.IP
	ls, err := ses.Cmd("HVALS", s.Config.Redis.Key).List()
	if err != nil {
		return nil, err
	}
	ips = []*models.IP{}
	for _, ipData := range ls {
		ip := models.NewIP()
		err = json.Unmarshal([]byte(ipData), &ip)
		if err != nil {
			return nil, err
		}
		ips = append(ips, ip)
	}
	return ips, nil
}

func (s *Storage) FindAll(value string) ([]*models.IP, error) {
	ses := s.GetDBSession()
	defer ses.Close()
	var ips []*models.IP
	ls, err := ses.Cmd("HVALS", s.Config.Redis.Key).List()
	if err != nil {
		return nil, err
	}
	ips = []*models.IP{}
	for _, ipData := range ls {
		ip := models.NewIP()
		err = json.Unmarshal([]byte(ipData), &ip)
		if err != nil {
			return nil, err
		}
		if ip.Type == value {
			ips = append(ips, ip)
		}
	}
	return ips, nil
}
