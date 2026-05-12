# Security Policy

## Reporting a vulnerability

Please **do not** open a public GitHub issue for security vulnerabilities.

Use **GitHub's private vulnerability reporting**:

1. Go to the [Security tab](https://github.com/nudgebee/nova/security) of this
   repository.
2. Click **Report a vulnerability**.
3. Fill in the form with details and, if possible, a minimal reproduction.

If for any reason that path is unavailable, contact
[security@nudgebee.com](mailto:security@nudgebee.com) instead. Encrypt
sensitive details if you have a key for that address; otherwise plain text is
acceptable for an initial contact.

## What to include

- Affected version (commit SHA or release tag).
- A clear description of the issue and the impact.
- Steps to reproduce, ideally a minimal case.
- Suggested fix or mitigation, if any.

## What to expect

- Acknowledgement within **3 business days**.
- An initial assessment (severity, scope, whether the issue is in this fork
  or in upstream code) within **10 business days**.
- A fix or mitigation plan, with timeline, communicated back to the reporter.
- Public disclosure coordinated with the reporter once a fix is available.

## Scope

This policy covers the code in **this repository**
(`github.com/nudgebee/nova`).

For vulnerabilities in **upstream nova**
(`github.com/FairwindsOps/nova`) that have not yet been backported here,
report them via Fairwinds' security process. We will pick up upstream fixes
on the next sync.

For vulnerabilities in **third-party dependencies**, file with the dependency
maintainer first; we will bump our `go.mod` to the patched version once it is
available.

## Supported versions

Only the latest `v<major>.<minor>.<patch>` release tag on `main` is actively
patched. Users on older tags should upgrade before reporting; backports are
case-by-case.
