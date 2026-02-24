package services

import "github.com/aneeshchawla/kubetools/backend/internal/models"

type TemplateService struct {
	templates []models.TemplateDefinition
}

func NewTemplateService() *TemplateService {
	return &TemplateService{templates: defaultTemplates()}
}

func (s *TemplateService) List() []models.TemplateDefinition {
	out := make([]models.TemplateDefinition, len(s.templates))
	copy(out, s.templates)
	return out
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
				{Path: "spec.template.spec.containers[0].name", Value: "postgres", Description: "Container name."},
				{Path: "spec.template.spec.containers[0].image", Value: "postgres:17", Description: "Container image."},
			},
			OptionalFields: []models.FieldDefinition{
				{Path: "spec.volumeClaimTemplates[0].metadata.name", Description: "PVC template name."},
				{Path: "spec.volumeClaimTemplates[0].spec.resources.requests.storage", Description: "Per-pod requested storage."},
				{Path: "spec.persistentVolumeClaimRetentionPolicy.whenDeleted", Description: "PVC retention policy."},
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
				{Path: "spec.dataSource.name", Description: "Snapshot/PVC data source."},
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
