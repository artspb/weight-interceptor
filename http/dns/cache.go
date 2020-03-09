package dns

import (
	"github.com/ReneKroon/ttlcache"
	"net"
	"time"
)

const defaultTtl = 43200 * time.Second

var cache *ttlcache.Cache

func init() {
	cache = ttlcache.NewCache()
	cache.SetTTL(defaultTtl)
	cache.SkipTtlExtensionOnHit(true)
}

func fromCache(host string) net.IP {
	ip, ok := cache.Get(host)
	if !ok {
		return nil
	}
	return ip.(net.IP)
}

func toCache(host string, ip net.IP, ttl time.Duration) {
	cache.SetWithTTL(host, ip, ttl)
}

func Shutdown() {
	cache.Close()
}
