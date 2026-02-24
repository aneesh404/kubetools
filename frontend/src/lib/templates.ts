import type { TemplateDefinition } from "../types/crd";

export const defaultTemplates: TemplateDefinition[] = [
  {
    id: "deployment",
    title: "Deployment",
    apiVersion: "apps/v1",
    kind: "Deployment",
    note: "Progressive delivery for stateless workloads with full rollout controls.",
    defaultFields: [
      {
        path: "metadata.name",
        value: "web-app",
        description: "Unique name for this deployment object."
      },
      {
        path: "metadata.namespace",
        value: "default",
        description: "Namespace where this deployment is created."
      },
      {
        path: "spec.replicas",
        value: "3",
        type: "number",
        description: "Desired number of running replicas."
      },
      {
        path: "spec.selector.matchLabels.app",
        value: "web-app",
        description: "Label selector used to match pod template labels."
      },
      {
        path: "spec.template.spec.containers[0].name",
        value: "app",
        description: "Primary container name."
      },
      {
        path: "spec.template.spec.containers[0].image",
        value: "nginx:1.27",
        description: "Container image with tag to deploy."
      }
    ],
    optionalFields: [
      {
        path: "spec.strategy.type",
        description: "RollingUpdate or Recreate strategy."
      },
      {
        path: "spec.template.spec.containers[0].ports[0].containerPort",
        type: "number",
        description: "Port exposed by the container."
      },
      {
        path: "spec.template.spec.imagePullSecrets[0].name",
        description: "Image pull secret for private registries."
      }
    ]
  },
  {
    id: "statefulset",
    title: "StatefulSet",
    apiVersion: "apps/v1",
    kind: "StatefulSet",
    note: "Stable identity and storage for ordered stateful workloads.",
    defaultFields: [
      {
        path: "metadata.name",
        value: "db",
        description: "StatefulSet name."
      },
      {
        path: "metadata.namespace",
        value: "default",
        description: "Namespace of this StatefulSet."
      },
      {
        path: "spec.serviceName",
        value: "db-headless",
        description: "Headless service backing stable pod identities."
      },
      {
        path: "spec.replicas",
        value: "2",
        type: "number",
        description: "Desired number of pod replicas."
      },
      {
        path: "spec.template.spec.containers[0].name",
        value: "postgres",
        description: "Container name in pod template."
      },
      {
        path: "spec.template.spec.containers[0].image",
        value: "postgres:17",
        description: "Image used by each replica pod."
      }
    ],
    optionalFields: [
      {
        path: "spec.volumeClaimTemplates[0].metadata.name",
        description: "Per-pod PVC template name."
      },
      {
        path: "spec.volumeClaimTemplates[0].spec.resources.requests.storage",
        description: "Storage capacity requested per replica volume."
      },
      {
        path: "spec.persistentVolumeClaimRetentionPolicy.whenDeleted",
        description: "How PVCs are retained when deleting StatefulSet."
      }
    ]
  },
  {
    id: "pvc",
    title: "PVC",
    apiVersion: "v1",
    kind: "PersistentVolumeClaim",
    note: "Declarative persistent storage request with class and size.",
    defaultFields: [
      {
        path: "metadata.name",
        value: "app-data",
        description: "PersistentVolumeClaim name."
      },
      {
        path: "metadata.namespace",
        value: "default",
        description: "Namespace containing this PVC."
      },
      {
        path: "spec.accessModes[0]",
        value: "ReadWriteOnce",
        description: "Access mode for mounting this volume."
      },
      {
        path: "spec.storageClassName",
        value: "standard",
        description: "StorageClass used for provisioning."
      },
      {
        path: "spec.resources.requests.storage",
        value: "20Gi",
        description: "Requested storage size."
      }
    ],
    optionalFields: [
      {
        path: "spec.volumeMode",
        description: "Filesystem or Block volume mode."
      },
      {
        path: "spec.dataSource.name",
        description: "Restore from snapshot/PVC data source."
      },
      {
        path: "spec.selector.matchLabels.tier",
        description: "Label selector for specific PV binding."
      }
    ]
  },
  {
    id: "volumesnapshot",
    title: "VolumeSnapshot",
    apiVersion: "snapshot.storage.k8s.io/v1",
    kind: "VolumeSnapshot",
    note: "Point-in-time PVC snapshots for backup, cloning, and restore.",
    defaultFields: [
      {
        path: "metadata.name",
        value: "db-snapshot-001",
        description: "VolumeSnapshot resource name."
      },
      {
        path: "metadata.namespace",
        value: "default",
        description: "Namespace where snapshot metadata is stored."
      },
      {
        path: "spec.volumeSnapshotClassName",
        value: "csi-hostpath-snapclass",
        description: "Snapshot class controlling driver behavior."
      },
      {
        path: "spec.source.persistentVolumeClaimName",
        value: "db-data",
        description: "Source PVC name to snapshot."
      }
    ],
    optionalFields: [
      {
        path: "metadata.labels.backup",
        description: "Label for retention and backup automation."
      },
      {
        path: "metadata.annotations.purpose",
        description: "Human-readable purpose annotation."
      },
      {
        path: "spec.source.volumeSnapshotContentName",
        description: "Reference an existing snapshot content object."
      }
    ]
  },
  {
    id: "cronjob",
    title: "CronJob",
    apiVersion: "batch/v1",
    kind: "CronJob",
    note: "Reliable scheduled jobs with native retry and history retention.",
    defaultFields: [
      {
        path: "metadata.name",
        value: "nightly-report",
        description: "CronJob object name."
      },
      {
        path: "metadata.namespace",
        value: "default",
        description: "Namespace where CronJob will run."
      },
      {
        path: "spec.schedule",
        value: "0 2 * * *",
        description: "Cron expression for job execution."
      },
      {
        path: "spec.jobTemplate.spec.template.spec.containers[0].name",
        value: "runner",
        description: "Container name used in jobs."
      },
      {
        path: "spec.jobTemplate.spec.template.spec.containers[0].image",
        value: "alpine:3.21",
        description: "Image used for each scheduled run."
      },
      {
        path: "spec.jobTemplate.spec.template.spec.restartPolicy",
        value: "OnFailure",
        description: "Pod restart behavior when job fails."
      }
    ],
    optionalFields: [
      {
        path: "spec.concurrencyPolicy",
        description: "Allow, Forbid, or Replace concurrent runs."
      },
      {
        path: "spec.successfulJobsHistoryLimit",
        type: "number",
        description: "How many successful jobs to retain."
      },
      {
        path: "spec.failedJobsHistoryLimit",
        type: "number",
        description: "How many failed jobs to retain."
      }
    ]
  }
];
