# nova (Nudgebee fork)

**Find outdated or deprecated Helm charts running in your cluster.**

> **This is a Nudgebee-maintained fork of [FairwindsOps/nova](https://github.com/FairwindsOps/nova).**
> Nudgebee is not affiliated with, endorsed by, or sponsored by FairwindsOps, Inc.
> The names "Nova" and "FairwindsOps" and the original Nova logo are trademarks of
> FairwindsOps, Inc.; this fork does not use the upstream logo in any released
> build (see [`NOTICE`](NOTICE) for details).
> The original work is © FairwindsOps, Inc. and is licensed under the
> Apache License 2.0 — see [`LICENSE`](LICENSE) and [`NOTICE`](NOTICE).

`nova` scans your cluster for installed Helm charts, then cross-checks them
against known Helm repositories. If it finds an updated version of a chart
you're using, or notices that your current version is deprecated, it will let
you know. It can also scan your cluster for out-of-date container images.

For the original feature documentation see the
[upstream docs](https://nova.docs.fairwinds.com).

---

## What's different in this fork

| Area | Change |
|---|---|
| Output | `ReleaseOutput` now includes the source Helm `RepositoryUrl` for matched releases. |
| Output | `VersionInfo` now includes a `Date` field — `FirstDeployed` for the installed version, `Created` for the latest version. |
| Discovery | New `pkg/helm/helm_repo_exporter.go` calls ArtifactHub's `/helm-exporter` API and (with `FindHelmRepoMatch` in `pkg/helm/findscore.go`) attaches the source repo URL to each release. Optional file cache via `ARTIFACT_HELM_REPO_HUB_CACHE_FILE`. |
| CLI | Removed the upstream "Want more? Automate Nova for free with Fairwinds Insights!" message that was printed after every command. |
| Build | Multi-stage `Dockerfile` that builds `nova` from source instead of copying a pre-built binary. |
| CI | GitHub Actions workflows that build a multi-arch image and push to Nudgebee's container registry. |
| Misc | Removed the upstream CircleCI pipeline, goreleaser config, and VuePress docs site (all of which were tied to FairwindsOps-private infra). |
| Deps | Many dependency upgrades. See `CHANGELOG.md` and `git log` for the full history. |

A more detailed list lives in [`CHANGELOG.md`](CHANGELOG.md). Significant
modifications are also summarized in [`NOTICE`](NOTICE).

## Installation

### From source

```sh
git clone https://github.com/nudgebee/nova.git
cd nova
make build
./nova --help
```

### Docker

```sh
docker build -t nova:dev .
docker run --rm -v ~/.kube:/root/.kube nova:dev find --wide
```

## Usage

```sh
# Find outdated Helm releases in the current kube context
nova find

# Include containers
nova find --containers

# Output JSON
nova find --format json
```

Run `nova --help` for the full command surface.

## Development

```sh
make lint    # golangci-lint
make test    # go test ./... with coverage
make build   # build the binary
```

## License

This project is licensed under the Apache License, Version 2.0 — see
[`LICENSE`](LICENSE). It is a fork of
[FairwindsOps/nova](https://github.com/FairwindsOps/nova) (also Apache 2.0).
The required attribution and a summary of modifications live in
[`NOTICE`](NOTICE).
