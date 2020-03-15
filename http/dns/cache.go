package dns

import (
	"github.com/ReneKroon/ttlcache"
	"time"
)

const defaultTtl = 43200 * time.Second

var cache *ttlcache.Cache

func init() {
	cache = ttlcache.NewCache()
	cache.SetTTL(defaultTtl)
	cache.SkipTtlExtensionOnHit(true)
}

func fromCache(host string) string {
	ip, ok := cache.Get(host)
	if !ok {
		return ""
	}
	return ip.(string)
}

func toCache(host string, ip string, ttl time.Duration) {
	cache.SetWithTTL(host, ip, ttl)
}

func Shutdown() {
	cache.Close()
}
