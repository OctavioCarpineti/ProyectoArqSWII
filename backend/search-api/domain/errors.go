package domain

import "errors"

var (
	// Errores de búsqueda
	ErrInvalidSearchQuery = errors.New("invalid search query")
	ErrSolrNotAvailable   = errors.New("solr search engine not available")
	ErrCacheError         = errors.New("cache error")

	// Errores de indexación
	ErrIndexingFailed   = errors.New("failed to index document in solr")
	ErrDocumentNotFound = errors.New("document not found in solr")

	// Errores de RabbitMQ
	ErrRabbitMQConnection = errors.New("rabbitmq connection error")
	ErrMessageProcessing  = errors.New("message processing error")
)
