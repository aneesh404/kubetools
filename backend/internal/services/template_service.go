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
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type TemplateService struct {
	client     *mongo.Client
	collection *mongo.Collection
	mu         sync.RWMutex
	templates  []models.TemplateDefinition
}

func NewTemplateService(ctx context.Context, cfg config.Config) (*TemplateService, error) {
	service := &TemplateService{templates: defaultTemplates()}

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

	collection := client.Database(cfg.MongoDatabase).Collection(cfg.MongoTemplateColl)
	service.client = client
	service.collection = collection

	indexCtx, indexCancel := context.WithTimeout(ctx, 5*time.Second)
	defer indexCancel()
	_, _ = collection.Indexes().CreateOne(indexCtx, mongo.IndexModel{
		Keys:    bson.D{{Key: "id", Value: 1}},
		Options: options.Index().SetUnique(true),
	})

	if err := service.seedDefaultsIfEmpty(ctx); err != nil {
		return service, fmt.Errorf("seed templates: %w", err)
	}

	return service, nil
}

func (s *TemplateService) Close(ctx context.Context) error {
	if s == nil || s.client == nil {
		return nil
	}
	return s.client.Disconnect(ctx)
}

func (s *TemplateService) List(ctx context.Context) ([]models.TemplateDefinition, error) {
	if s.collection == nil {
		s.mu.RLock()
		defer s.mu.RUnlock()
		return cloneTemplateList(s.templates), nil
	}

	cursor, err := s.collection.Find(ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "title", Value: 1}}))
	if err != nil {
		return nil, fmt.Errorf("list templates: %w", err)
	}
	defer cursor.Close(ctx)

	out := make([]models.TemplateDefinition, 0, 16)
	for cursor.Next(ctx) {
		var item models.TemplateDefinition
		if err := cursor.Decode(&item); err != nil {
			return nil, fmt.Errorf("decode template: %w", err)
		}
		out = append(out, item)
	}
	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("template cursor: %w", err)
	}

	if len(out) == 0 {
		return cloneTemplateList(defaultTemplates()), nil
	}

	return out, nil
}

func (s *TemplateService) Upsert(ctx context.Context, template models.TemplateDefinition) error {
	template.ID = strings.TrimSpace(template.ID)
	if template.ID == "" {
		return fmt.Errorf("template id is required")
	}

	if s.collection == nil {
		s.mu.Lock()
		defer s.mu.Unlock()
		s.templates = upsertTemplateInMemory(s.templates, template)
		return nil
	}

	_, err := s.collection.UpdateOne(
		ctx,
		bson.M{"id": template.ID},
		bson.M{"$set": template},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		return fmt.Errorf("upsert template: %w", err)
	}
	return nil
}

func (s *TemplateService) seedDefaultsIfEmpty(ctx context.Context) error {
	if s.collection == nil {
		return nil
	}

	count, err := s.collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("count templates: %w", err)
	}
	if count > 0 {
		return nil
	}

	defaults := defaultTemplates()
	docs := make([]any, 0, len(defaults))
	for _, template := range defaults {
		docs = append(docs, template)
	}

	if len(docs) == 0 {
		return nil
	}
	_, err = s.collection.InsertMany(ctx, docs)
	if err != nil {
		return fmt.Errorf("insert default templates: %w", err)
	}
	return nil
}

func cloneTemplateList(in []models.TemplateDefinition) []models.TemplateDefinition {
	out := make([]models.TemplateDefinition, len(in))
	copy(out, in)
	return out
}

func upsertTemplateInMemory(list []models.TemplateDefinition, template models.TemplateDefinition) []models.TemplateDefinition {
	for i := range list {
		if list[i].ID == template.ID {
			list[i] = template
			return list
		}
	}
	return append(list, template)
}

func defaultTemplates() []models.TemplateDefinition {
	return []models.TemplateDefinition{
		{
			ID:         "deployment",
			Title:      "Deployment",
			APIVersion: "apps/v1",
			Kind:       "Deployment",
			Note:       "Progressive delivery for stateless workloads with rollout controls.",
			DefaultFields: []models.FieldDefinition{
				{Path: "metadata.name", Value: "web-app", Description: "Unique deployment name."},
				{Path: "metadata.namespace", Value: "default", Description: "Target namespace."},
				{Path: "spec.replicas", Value: "3", Type: "number", Description: "Desired replica count."},
				{Path: "spec.selector.matchLabels.app", Value: "web-app", Description: "Pod label selector."},
				{Path: "spec.template.spec.containers[0].name", Value: "app", Description: "Container name."},
				{Path: "spec.template.spec.containers[0].image", Value: "nginx:1.27", Description: "Container image."},
			},
			OptionalFields: []models.FieldDefinition{
				{Path: "spec.strategy.type", Description: "Deployment strategy type."},
				{Path: "spec.template.spec.containers[0].ports[0].containerPort", Type: "number", Description: "Exposed container port."},
				{Path: "spec.template.spec.imagePullSecrets[0].name", Description: "Image pull secret."},
			},
		},
		{
			ID:         "statefulset",
			Title:      "StatefulSet",
			APIVersion: "apps/v1",
			Kind:       "StatefulSet",
			Note:       "Stable identity and storage for stateful workloads.",
			DefaultFields: []models.FieldDefinition{
				{Path: "metadata.name", Value: "db", Description: "StatefulSet name."},
				{Path: "metadata.namespace", Value: "default", Description: "Target namespace."},
				{Path: "spec.serviceName", Value: "db-headless", Description: "Headless service name."},
				{Path: "spec.replicas", Value: "2", Type: "number", Description: "Replica count."},
				{Path: "spec.selector.matchLabels.app", Value: "db", Description: "Selector labels matching pod template labels."},
				{Path: "spec.template.metadata.labels.app", Value: "db", Description: "Pod template labels for selector matching."},
				{Path: "spec.template.spec.containers[0].name", Value: "postgres", Description: "Container name."},
				{Path: "spec.template.spec.containers[0].image", Value: "postgres:17", Description: "Container image."},
			},
			OptionalFields: []models.FieldDefinition{
				{Path: "spec.volumeClaimTemplates[0].metadata.name", Description: "PVC template name."},
				{Path: "spec.volumeClaimTemplates[0].spec.resources.requests.storage", Description: "Per-pod requested storage."},
				{Path: "spec.persistentVolumeClaimRetentionPolicy.whenDeleted", Description: "PVC retention policy."},
				{Path: "spec.persistentVolumeClaimRetentionPolicy.whenScaled", Description: "PVC retention when scaling down."},
			},
		},
		{
			ID:         "pvc",
			Title:      "PVC",
			APIVersion: "v1",
			Kind:       "PersistentVolumeClaim",
			Note:       "Declarative persistent storage request with class and size.",
			DefaultFields: []models.FieldDefinition{
				{Path: "metadata.name", Value: "app-data", Description: "PVC name."},
				{Path: "metadata.namespace", Value: "default", Description: "Target namespace."},
				{Path: "spec.accessModes[0]", Value: "ReadWriteOnce", Description: "Access mode."},
				{Path: "spec.storageClassName", Value: "standard", Description: "StorageClass name."},
				{Path: "spec.resources.requests.storage", Value: "20Gi", Description: "Requested size."},
			},
			OptionalFields: []models.FieldDefinition{
				{Path: "spec.volumeMode", Description: "Filesystem or Block mode."},
				{Path: "spec.volumeName", Description: "Bind to an existing pre-provisioned volume."},
				{Path: "spec.selector.matchLabels.tier", Description: "PV selector label."},
			},
		},
		{
			ID:         "volumesnapshot",
			Title:      "VolumeSnapshot",
			APIVersion: "snapshot.storage.k8s.io/v1",
			Kind:       "VolumeSnapshot",
			Note:       "Point-in-time snapshots for backup and restore.",
			DefaultFields: []models.FieldDefinition{
				{Path: "metadata.name", Value: "db-snapshot-001", Description: "Snapshot name."},
				{Path: "metadata.namespace", Value: "default", Description: "Target namespace."},
				{Path: "spec.volumeSnapshotClassName", Value: "csi-hostpath-snapclass", Description: "Snapshot class."},
				{Path: "spec.source.persistentVolumeClaimName", Value: "db-data", Description: "Source PVC."},
			},
			OptionalFields: []models.FieldDefinition{
				{Path: "metadata.labels.backup", Description: "Backup retention label."},
				{Path: "metadata.annotations.purpose", Description: "Snapshot purpose annotation."},
				{Path: "spec.source.volumeSnapshotContentName", Description: "Pre-provisioned snapshot content."},
			},
		},
		{
			ID:         "cronjob",
			Title:      "CronJob",
			APIVersion: "batch/v1",
			Kind:       "CronJob",
			Note:       "Scheduled Kubernetes jobs with retry controls.",
			DefaultFields: []models.FieldDefinition{
				{Path: "metadata.name", Value: "nightly-report", Description: "CronJob name."},
				{Path: "metadata.namespace", Value: "default", Description: "Target namespace."},
				{Path: "spec.schedule", Value: "0 2 * * *", Description: "Cron schedule expression."},
				{Path: "spec.jobTemplate.spec.template.spec.containers[0].name", Value: "runner", Description: "Container name."},
				{Path: "spec.jobTemplate.spec.template.spec.containers[0].image", Value: "alpine:3.21", Description: "Container image."},
				{Path: "spec.jobTemplate.spec.template.spec.restartPolicy", Value: "OnFailure", Description: "Restart behavior."},
			},
			OptionalFields: []models.FieldDefinition{
				{Path: "spec.concurrencyPolicy", Description: "Concurrency handling."},
				{Path: "spec.successfulJobsHistoryLimit", Type: "number", Description: "Successful job history length."},
				{Path: "spec.failedJobsHistoryLimit", Type: "number", Description: "Failed job history length."},
			},
		},
	}
}
