package config

import (
	"log"
	"os"
)

// GetSolrURL obtiene la URL de SolR desde variables de entorno
func GetSolrURL() string {
	url := os.Getenv("SOLR_URL")
	if url == "" {
		url = "http://localhost:8983/solr/schedules"
	}
	log.Printf("🔍 SolR URL: %s", url)
	return url
}
