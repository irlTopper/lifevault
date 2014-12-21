package cache

import (
	"strings"

	"github.com/revel/revel"
	redisCache "github.com/revel/revel/cache"
)

var RedisPersist redisCache.RedisCache

func init() {
	revel.OnAppStart(func() {
		hosts := strings.Split(revel.Config.StringDefault("redis.persist.hosts", ""), ",")
		if len(hosts) == 0 {
			panic("The persistent redis server has not been specified - add 'redis.persist.hosts' to app.conf")
		}
		if len(hosts) > 1 {
			panic("Redis currently only supports one host!")
		}
		password := revel.Config.StringDefault("redis.persist.password", "")
		//defaultExpiration := time.Hour // The default for the default is one hour.
		RedisPersist = redisCache.NewRedisCache(hosts[0], password, redisCache.FOREVER)
	})
}
