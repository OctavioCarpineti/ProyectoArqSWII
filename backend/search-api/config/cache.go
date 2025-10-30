package config

import (
	"log"
	"os"
)

// GetMemcachedHost obtiene el host de Memcached desde variables de entorno
func GetMemcachedHost() string {
	host := os.Getenv("MEMCACHED_HOST")
	if host == "" {
		host = "localhost:11211"
	}
	log.Printf("📦 Memcached host: %s", host)
	return host
}
