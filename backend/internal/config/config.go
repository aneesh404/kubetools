package config

import (
	"os"
	"strings"
)

type Config struct {
	Host              string
	Port              string
	CORSOrigins       []string
	MongoURI          string
	MongoDatabase     string
	MongoManifestColl string
}

func Load() Config {
	host := getenv("HOST", "0.0.0.0")
	port := getenv("PORT", "8080")
	originsRaw := getenv("CORS_ORIGINS", "http://localhost:5173")
	mongoURI := getenv("MONGODB_URI", "mongodb://localhost:27017")
	mongoDatabase := getenv("MONGODB_DATABASE", "kubebuilder")
	mongoManifestColl := getenv("MONGODB_MANIFEST_COLLECTION", "manifests")
	origins := strings.Split(originsRaw, ",")
	for i := range origins {
		origins[i] = strings.TrimSpace(origins[i])
	}

	return Config{
		Host:              host,
		Port:              port,
		CORSOrigins:       origins,
		MongoURI:          mongoURI,
		MongoDatabase:     mongoDatabase,
		MongoManifestColl: mongoManifestColl,
	}
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
