# Contributing to nova (Nudgebee fork)

Thanks for considering a contribution. This repository is a Nudgebee-maintained
fork of [FairwindsOps/nova](https://github.com/FairwindsOps/nova). Please read
the [`README`](README.md) and [`NOTICE`](NOTICE) for the relationship to
upstream and the trademark disclaimer.

## Where to file what

- **Bugs or features that exist upstream**: prefer
  [FairwindsOps/nova](https://github.com/FairwindsOps/nova/issues). Fixes that
  land upstream get pulled in here on the next sync.
- **Bugs or features specific to this fork**: file in
  [this repository's issues](https://github.com/nudgebee/nova/issues). Use the
  bug / feature-request templates under `.github/ISSUE_TEMPLATE/`.
- **Security vulnerabilities**: see [`SECURITY.md`](SECURITY.md). Do **not**
  open a public issue for security reports.

## Branch model

```
main (development) ──PR──▶ prod (production)
```

- All work targets `main`.
- `prod` is fast-forwarded from `main` for releases.
- Direct pushes to `main` and `prod` are not accepted; everything goes through
  a pull request.

## Local development

Requires Go 1.26+ (matching the `go` directive in `go.mod`) and
[`golangci-lint`](https://golangci-lint.run) v2.11.3 or later.

```bash
go build ./...
golangci-lint run ./...
go test ./...
```

These three commands are the same ones the CI workflow
(`.github/workflows/ci.yaml`) runs on every pull request. A PR will not be
merged with any of them red.

## Commit style

Conventional Commit style, lowercase type, optional scope:

```
<type>(<scope>): <short summary>
```

Examples from the existing history:

- `fix(helm): handle nil release.Info`
- `ci: migrate image publishing from ECR to GHCR`
- `chore(deps): bump k8s.io/client-go from 0.35.1 to 0.36.0`

Common types: `feat`, `fix`, `chore`, `refactor`, `test`, `ci`, `docs`.

## Pull requests

- Fill out [`.github/pull_request_template.md`](.github/pull_request_template.md).
- Link an issue with `Fixes #<n>` when one exists.
- Keep the PR scoped to a single concern. If reviewers ask for unrelated
  changes, push back or split into a follow-up PR.
- Add or update tests for behavior changes. Pure refactors do not need new
  tests but must keep the existing suite green.
- Dependency bumps are handled by Dependabot; manual `go.mod` edits should
  explain *why* in the commit body.

## License

By contributing, you agree that your contributions will be licensed under the
[Apache License, Version 2.0](LICENSE). Significant fork-specific
contributions are summarized in [`NOTICE`](NOTICE).
