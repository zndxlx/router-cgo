package main

//	log "github.com/thinkboy/log4go"

//	"github.com/garyburd/redigo/redis"

type ClientStatus struct {
    UserId int64
    Status int16
}

const (
    Connected    int16 = 1
    Disconnected int16 = 2
)

//var RedisPool chan []redis.Conn

//func InitRedis(redisAddr string) (err error) {
//	if len(RedisPool) == 0 {

//	}
//}

//func GetRedis() redis.Conn {

//}
