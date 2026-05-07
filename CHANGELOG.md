# Changelog

## Nudgebee fork

This section tracks changes made by Nudgebee since forking
[FairwindsOps/nova](https://github.com/FairwindsOps/nova) in June 2024.
Versions and dates above this section are from upstream and are preserved
for historical context.

### Added
* `RepositoryUrl` field on `ReleaseOutput` exposing the source Helm repository
  URL for matched releases (`pkg/output/output.go`).
* `Date` field on `VersionInfo` exposing the installed release's
  `FirstDeployed` timestamp and the latest chart's `Created` timestamp
  (`pkg/output/output.go`, `pkg/helm/chartrepo.go`, `pkg/helm/findscore.go`).
* `pkg/helm/helm_repo_exporter.go` — client for ArtifactHub's `/helm-exporter`
  API used to discover the source repository for an installed chart.
  Supports a local file cache via the `ARTIFACT_HELM_REPO_HUB_CACHE_FILE`
  environment variable.
* `FindHelmRepoMatch` in `pkg/helm/findscore.go` — scores helm-exporter
  packages against an installed release and attaches the matched repository
  URL to the output.
* GitHub Actions workflows that build a multi-arch container image and push
  to AWS ECR (`.github/workflows/build-image-{dev,prod}.yaml`).

### Changed
* `Dockerfile` now performs a multi-stage build of `nova` from source
  (`golang:alpine` builder + `alpine` runtime) instead of copying a
  pre-built binary.
* Many dependency upgrades, including Helm, the Kubernetes client libraries,
  controller-runtime, and the Go toolchain. See `git log` for the full list.

### Removed
* The "Want more? Automate Nova for free with Fairwinds Insights!" message
  printed to stderr after every CLI run (`cmd/root.go`).
* The `docs/` directory and its VuePress site.
* `.circleci/` configuration — the upstream CircleCI pipeline was tied to
  Fairwinds-private orbs, registries, and credentials.
* `.goreleaser.yml` — the upstream goreleaser config targeted the
  FairwindsOps GitHub release, FairwindsOps homebrew tap, and
  `quay.io/fairwinds/nova`. Nudgebee does not currently release via
  goreleaser.

---

## Upstream history (FairwindsOps/nova)

# 2.2.0
* Add support for config files
* Add support for overriding desired-versions

# 2.1.0
* Support for `--context` flag to specify k8s context

# 2.0.2
* Added Helm version to output

# 2.0.1
* Fixed CI

# 2.0.0
* Changed output format
* Added new auth providers

## 1.0.1
* Build docker image for tags
* Fix #8: update release version during build

## 1.0.0

* Migrated existing functionality from private repository
