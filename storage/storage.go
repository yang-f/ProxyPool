package storage

import (
	"encoding/json"
	"github.com/fzzy/radix/redis"
	"github.com/go-playground/log"
	"github.com/yang-f/ProxyPool/models"
	"github.com/yang-f/ProxyPool/util"
	"time"
)

// Config 全局配置文件
var Config = util.NewConfig()

// GlobalMgoSession 全局连接Session

// Storage struct is used for storeing persistent data of alerts
type Storage struct {
}

// NewStorage creates and returns new Storage instance
func NewStorage() *Storage {
	return &Storage{}
}

// GetDBSession returns a new connection from the pool
func (s *Storage) GetDBSession() *redis.Client {
	client, _ := redis.DialTimeout("tcp", Config.Redis.Addr, time.Duration(10)*time.Second)
	return client
}

// Create insert new item
func (s *Storage) Create(ip *models.IP) error {
	log.Printf("Create:%v", ip)
	ses := s.GetDBSession()
	defer ses.Close()
	value, err := json.Marshal(ip)
	if err != nil {
		return err
	}
	r := ses.Cmd("HMSET", Config.Redis.Key, ip.Data, string(value))
	if r.Err != nil {
		return r.Err
	}
	return nil
}

// GetOne Finds and returns one data from storage
func (s *Storage) GetOne(value string) (*models.IP, error) {
	ses := s.GetDBSession()
	defer ses.Close()
	ls, err := ses.Cmd("mget", value).List()
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
	r := ses.Cmd("HLEN", Config.Redis.Key)
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
	r := ses.Cmd("HDEL", Config.Redis.Key, ip.Data)
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
	r := ses.Cmd("HMSET", Config.Redis.Key, ip.Data, string(value))
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
	ls, err := ses.Cmd("HVALS", Config.Redis.Key).List()
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

//func (s *Storage) FindAll(value string) ([]*models.IP, error) {
//	ses := s.GetDBSession()
//	defer ses.Close()
//	var ips []*models.IP
//	err := ses.DB(s.database).C(s.table).Find(bson.M{"type": bson.M{"$regex": value, "$options": "$i"}}).All(&ips)
//	if err != nil {
//		return nil, err
//	}
//	return ips, nil
//}
