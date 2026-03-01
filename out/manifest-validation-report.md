Summary: 25/25 passed, 0 failed, kubectl infra-blocked in 25 cases.
# Manifest Dry-Run Loop Report
Generated at: 2026-03-01T16:16:51.590Z
API base: http://localhost:8081/api/v1
Templates validated: 6

## MadaraInstanceType (karnot.xyz/v1alpha1)
Source: CRD:Schema: MadaraInstanceType
Variants: 5
- base: PASS | yq=ok | kubeconform=ok | kubectl=infra-blocked
- optional-1: PASS | yq=ok | kubeconform=ok | kubectl=infra-blocked
- optional-4: PASS | yq=ok | kubeconform=ok | kubectl=infra-blocked
- mutated-defaults: PASS | yq=ok | kubeconform=ok | kubectl=infra-blocked
- dense: PASS | yq=ok | kubeconform=ok | kubectl=infra-blocked

## Deployment (apps/v1)
Source: default-template
Variants: 4
- base: PASS | yq=ok | kubeconform=ok | kubectl=infra-blocked
- optional-1: PASS | yq=ok | kubeconform=ok | kubectl=infra-blocked
- optional-4: PASS | yq=ok | kubeconform=ok | kubectl=infra-blocked
- mutated-defaults: PASS | yq=ok | kubeconform=ok | kubectl=infra-blocked

## StatefulSet (apps/v1)
Source: default-template
Variants: 4
- base: PASS | yq=ok | kubeconform=ok | kubectl=infra-blocked
- optional-1: PASS | yq=ok | kubeconform=ok | kubectl=infra-blocked
- optional-4: PASS | yq=ok | kubeconform=ok | kubectl=infra-blocked
- mutated-defaults: PASS | yq=ok | kubeconform=ok | kubectl=infra-blocked

## PersistentVolumeClaim (v1)
Source: default-template
Variants: 4
- base: PASS | yq=ok | kubeconform=ok | kubectl=infra-blocked
- optional-1: PASS | yq=ok | kubeconform=ok | kubectl=infra-blocked
- optional-4: PASS | yq=ok | kubeconform=ok | kubectl=infra-blocked
- mutated-defaults: PASS | yq=ok | kubeconform=ok | kubectl=infra-blocked

## VolumeSnapshot (snapshot.storage.k8s.io/v1)
Source: default-template
Variants: 4
- base: PASS | yq=ok | kubeconform=ok | kubectl=infra-blocked
- optional-1: PASS | yq=ok | kubeconform=ok | kubectl=infra-blocked
- optional-4: PASS | yq=ok | kubeconform=ok | kubectl=infra-blocked
- mutated-defaults: PASS | yq=ok | kubeconform=ok | kubectl=infra-blocked

## CronJob (batch/v1)
Source: default-template
Variants: 4
- base: PASS | yq=ok | kubeconform=ok | kubectl=infra-blocked
- optional-1: PASS | yq=ok | kubeconform=ok | kubectl=infra-blocked
- optional-4: PASS | yq=ok | kubeconform=ok | kubectl=infra-blocked
- mutated-defaults: PASS | yq=ok | kubeconform=ok | kubectl=infra-blocked

