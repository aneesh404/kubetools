package services

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/aneeshchawla/kubetools/backend/internal/config"
	"github.com/aneeshchawla/kubetools/backend/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type ManifestService struct {
	client     *mongo.Client
	collection *mongo.Collection
	mu         sync.RWMutex
	memory     []models.ManifestRecord
}

func NewManifestService(ctx context.Context, cfg config.Config) (*ManifestService, error) {
	service := &ManifestService{
		memory: make([]models.ManifestRecord, 0, 64),
	}

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		return service, fmt.Errorf("connect mongodb: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := client.Ping(pingCtx, readpref.Primary()); err != nil {
		_ = client.Disconnect(context.Background())
		return service, fmt.Errorf("ping mongodb: %w", err)
	}

	collection := client.Database(cfg.MongoDatabase).Collection(cfg.MongoManifestColl)

	indexCtx, indexCancel := context.WithTimeout(ctx, 5*time.Second)
	defer indexCancel()
	_, _ = collection.Indexes().CreateOne(indexCtx, mongo.IndexModel{
		Keys: bson.D{{Key: "createdAt", Value: -1}},
	})

	service.client = client
	service.collection = collection

	return service, nil
}

func (s *ManifestService) Close(ctx context.Context) error {
	if s == nil || s.client == nil {
		return nil
	}
	return s.client.Disconnect(ctx)
}

func (s *ManifestService) SaveManifest(ctx context.Context, req models.SaveManifestRequest) (models.ManifestRecord, error) {
	if strings.TrimSpace(req.YAML) == "" {
		return models.ManifestRecord{}, fmt.Errorf("yaml is required")
	}

	now := time.Now().UTC()
	record := models.ManifestRecord{
		ID:         primitive.NewObjectID().Hex(),
		Title:      fallback(req.Title, "Manifest"),
		Resource:   strings.TrimSpace(req.Resource),
		APIVersion: strings.TrimSpace(req.APIVersion),
		Kind:       strings.TrimSpace(req.Kind),
		YAML:       req.YAML,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if s.collection == nil {
		s.mu.Lock()
		defer s.mu.Unlock()
		s.memory = append([]models.ManifestRecord{record}, s.memory...)
		if len(s.memory) > 200 {
			s.memory = s.memory[:200]
		}
		return record, nil
	}

	if _, err := s.collection.InsertOne(ctx, record); err != nil {
		return models.ManifestRecord{}, fmt.Errorf("insert manifest: %w", err)
	}
	return record, nil
}

func (s *ManifestService) ListManifests(ctx context.Context, query string, limit int64) ([]models.ManifestRecord, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	if s.collection == nil {
		s.mu.RLock()
		defer s.mu.RUnlock()

		if query == "" {
			if int64(len(s.memory)) <= limit {
				return append([]models.ManifestRecord(nil), s.memory...), nil
			}
			return append([]models.ManifestRecord(nil), s.memory[:limit]...), nil
		}

		lowerQuery := strings.ToLower(strings.TrimSpace(query))
		out := make([]models.ManifestRecord, 0, limit)
		for _, item := range s.memory {
			if matchesManifestQuery(item, lowerQuery) {
				out = append(out, item)
			}
			if int64(len(out)) >= limit {
				break
			}
		}

		return out, nil
	}

	filter := bson.M{}
	trimmed := strings.TrimSpace(query)
	if trimmed != "" {
		regex := bson.M{"$regex": trimmed, "$options": "i"}
		filter = bson.M{
			"$or": []bson.M{
				{"title": regex},
				{"resource": regex},
				{"kind": regex},
				{"apiVersion": regex},
				{"yaml": regex},
			},
		}
	}

	cursor, err := s.collection.Find(
		ctx,
		filter,
		options.Find().
			SetSort(bson.D{{Key: "createdAt", Value: -1}}).
			SetLimit(limit),
	)
	if err != nil {
		return nil, fmt.Errorf("list manifests: %w", err)
	}
	defer cursor.Close(ctx)

	out := make([]models.ManifestRecord, 0)
	for cursor.Next(ctx) {
		var item models.ManifestRecord
		if err := cursor.Decode(&item); err != nil {
			return nil, fmt.Errorf("decode manifest: %w", err)
		}
		out = append(out, item)
	}
	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("manifest cursor: %w", err)
	}

	return out, nil
}

func fallback(value string, defaultValue string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return defaultValue
	}
	return trimmed
}

func matchesManifestQuery(item models.ManifestRecord, lowerQuery string) bool {
	return strings.Contains(strings.ToLower(item.Title), lowerQuery) ||
		strings.Contains(strings.ToLower(item.Resource), lowerQuery) ||
		strings.Contains(strings.ToLower(item.Kind), lowerQuery) ||
		strings.Contains(strings.ToLower(item.APIVersion), lowerQuery) ||
		strings.Contains(strings.ToLower(item.YAML), lowerQuery)
}
