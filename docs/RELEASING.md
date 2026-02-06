# Release Process

Complete guide for creating and publishing sortTF releases.

## Table of Contents

- [Overview](#overview)
- [Versioning Strategy](#versioning-strategy)
- [Release Workflow](#release-workflow)
- [Pre-Release Checklist](#pre-release-checklist)
- [Creating a Release](#creating-a-release)
- [Post-Release Tasks](#post-release-tasks)
- [Release Types](#release-types)
- [Troubleshooting](#troubleshooting)
- [Rollback Procedure](#rollback-procedure)

## Overview

sortTF uses an automated release process triggered by Git tags. The release workflow:

1. Validates version format
2. Runs full test suite
3. Builds binaries for all platforms
4. Generates checksums
5. Creates comprehensive changelog
6. Publishes GitHub Release

## Versioning Strategy

sortTF follows [Semantic Versioning (SemVer)](https://semver.org/) 2.0.0:

```
v<MAJOR>.<MINOR>.<PATCH>[-<PRERELEASE>]

Examples:
  v1.0.0          - Stable release
  v1.2.3          - Stable release with fixes
  v2.0.0          - Major version (breaking changes)
  v1.0.0-rc.1     - Release candidate
  v1.0.0-beta.2   - Beta release
  v1.0.0-alpha.1  - Alpha release
```

### Version Components

- **MAJOR**: Breaking changes, incompatible API changes
- **MINOR**: New features, backward-compatible
- **PATCH**: Bug fixes, backward-compatible
- **PRERELEASE**: Pre-release identifier (alpha, beta, rc)

### Version Increment Rules

| Change Type | Example | When to Use |
|-------------|---------|-------------|
| **Major** | v1.0.0 → v2.0.0 | Breaking API changes, removed features |
| **Minor** | v1.0.0 → v1.1.0 | New features, enhancements (backward-compatible) |
| **Patch** | v1.0.0 → v1.0.1 | Bug fixes, security patches |
| **RC** | v1.0.0-rc.1 | Release candidate before major release |
| **Beta** | v1.0.0-beta.1 | Feature-complete but needs testing |
| **Alpha** | v1.0.0-alpha.1 | Early development, unstable |

### Examples of Version Increments

**Bug Fix** (Patch):
```bash
# Current: v1.2.3
# Fix: Nested blocks not sorted correctly
# New version: v1.2.4
```

**New Feature** (Minor):
```bash
# Current: v1.2.4
# Feature: Add support for `moved` blocks
# New version: v1.3.0
```

**Breaking Change** (Major):
```bash
# Current: v1.3.0
# Breaking: Change API function signatures
# New version: v2.0.0
```

**Pre-release**:
```bash
# Testing v2.0.0 features
v2.0.0-alpha.1  # Initial testing
v2.0.0-alpha.2  # More testing
v2.0.0-beta.1   # Feature complete
v2.0.0-rc.1     # Release candidate
v2.0.0          # Final release
```

## Release Workflow

sortTF uses a **tag-based release workflow** with automated GitHub Actions.

### Workflow Diagram

```
Developer                Git                GitHub Actions
    │                     │                       │
    │  Create & Push Tag  │                       │
    │────────────────────▶│                       │
    │                     │                       │
    │                     │   Trigger Workflow    │
    │                     │──────────────────────▶│
    │                     │                       │
    │                     │                 Validate Version
    │                     │                       │
    │                     │                   Run Tests
    │                     │                       │
    │                     │               Build Binaries
    │                     │                       │
    │                     │              Create Checksums
    │                     │                       │
    │                     │            Generate Changelog
    │                     │                       │
    │                     │           Create GitHub Release
    │                     │◀──────────────────────│
    │                     │                       │
    │  Release Published  │                       │
    │◀────────────────────│                       │
    │                     │                       │
```

### Trigger Methods

#### 1. Automatic (Recommended) - Tag Push

Push a version tag to trigger automatic release:

```bash
git tag v1.0.0
git push origin v1.0.0
```

#### 2. Manual - Workflow Dispatch

Manually trigger via GitHub Actions UI:

1. Go to GitHub Actions → Release workflow
2. Click "Run workflow"
3. Enter version (e.g., `v1.0.0`)
4. Optionally mark as draft or pre-release
5. Click "Run workflow"

## Pre-Release Checklist

Before creating a release, ensure:

### 1. Code Quality

- [ ] All tests pass: `go test ./...`
- [ ] Linter passes: `golangci-lint run`
- [ ] Code formatted: `go fmt ./...`
- [ ] No race conditions: `go test -race ./...`
- [ ] Dependencies updated: `go mod tidy`

### 2. Testing

- [ ] Unit tests pass (155+ tests)
- [ ] Integration tests pass (29 tests)
- [ ] Manual testing of key features
- [ ] Tested on multiple platforms (Linux, macOS, Windows)
- [ ] Test coverage maintained (≥90%)

### 3. Documentation

- [ ] README updated with new features
- [ ] CHANGELOG updated (if maintained separately)
- [ ] API documentation current
- [ ] Usage examples verified
- [ ] Migration guide for breaking changes (major versions)

### 4. Version Check

- [ ] Determined correct version number (major/minor/patch)
- [ ] No conflicts with existing tags
- [ ] Version follows SemVer conventions

### 5. Branch State

- [ ] Working on `main` branch (or release branch)
- [ ] All changes committed
- [ ] Branch up-to-date with remote
- [ ] No uncommitted changes: `git status`

## Creating a Release

### Step-by-Step Process

#### 1. Prepare the Release

```bash
# Ensure you're on main branch
git checkout main

# Pull latest changes
git pull origin main

# Verify everything is clean
git status

# Run full test suite
go test ./...
go test -race ./...

# Run linter
golangci-lint run

# Verify build works
go build ./cmd/sorttf
```

#### 2. Determine Version Number

```bash
# Check current version
git describe --tags --abbrev=0

# Example output: v1.2.3

# Decide next version based on changes:
# - Bug fixes only → v1.2.4 (patch)
# - New features → v1.3.0 (minor)
# - Breaking changes → v2.0.0 (major)
```

#### 3. Create and Push Tag

```bash
# Create annotated tag
VERSION="v1.3.0"
git tag -a "$VERSION" -m "Release $VERSION"

# Verify tag was created
git tag -l "$VERSION"

# Push tag to trigger release
git push origin "$VERSION"
```

#### 4. Monitor Release Workflow

```bash
# Watch GitHub Actions
# https://github.com/OBerger96/sortTF/actions

# Workflow will:
# 1. Validate version format
# 2. Run tests
# 3. Build binaries (5 platforms)
# 4. Create checksums
# 5. Generate changelog
# 6. Create GitHub Release
```

#### 5. Verify Release

```bash
# Check release was created
# https://github.com/OBerger96/sortTF/releases

# Verify artifacts are present:
# - sorttf-linux-amd64
# - sorttf-linux-arm64
# - sorttf-darwin-amd64
# - sorttf-darwin-arm64
# - sorttf-windows-amd64.exe
# - checksums.txt

# Download and test a binary
wget https://github.com/OBerger96/sortTF/releases/download/$VERSION/sorttf-linux-amd64
chmod +x sorttf-linux-amd64
./sorttf-linux-amd64 --version
```

### Quick Reference

```bash
# Standard release process
git checkout main
git pull origin main
go test ./...
golangci-lint run
git tag -a v1.3.0 -m "Release v1.3.0"
git push origin v1.3.0
```

## Post-Release Tasks

After release is published:

### 1. Verify Installation

Test installation methods:

```bash
# Test go install
go install github.com/OBerger96/sortTF/cmd/sorttf@v1.3.0
sorttf --version

# Test binary download
wget https://github.com/OBerger96/sortTF/releases/download/v1.3.0/sorttf-linux-amd64
chmod +x sorttf-linux-amd64
./sorttf-linux-amd64 --version
```

### 2. Update Documentation

- [ ] Update README badges if needed
- [ ] Update installation instructions
- [ ] Announce release in discussions/community

### 3. Communicate Release

- [ ] Post release announcement (GitHub Discussions, Twitter, etc.)
- [ ] Notify major users of breaking changes (if applicable)
- [ ] Update any external documentation or integrations

### 4. Monitor Issues

- Watch for any issues reported with new release
- Be ready to create hotfix release if critical bugs found

## Release Types

### Stable Release

```bash
# For production use
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

**Characteristics:**
- Fully tested
- Production-ready
- Marked as "Latest" on GitHub
- Recommended for all users

### Pre-release (Alpha/Beta/RC)

```bash
# Alpha - Early testing
git tag -a v2.0.0-alpha.1 -m "Release v2.0.0-alpha.1"
git push origin v2.0.0-alpha.1

# Beta - Feature complete
git tag -a v2.0.0-beta.1 -m "Release v2.0.0-beta.1"
git push origin v2.0.0-beta.1

# Release Candidate - Final testing
git tag -a v2.0.0-rc.1 -m "Release v2.0.0-rc.1"
git push origin v2.0.0-rc.1
```

**Characteristics:**
- Marked as "Pre-release" on GitHub
- Not recommended for production
- Used for testing before stable release
- NOT marked as "Latest"

### Hotfix Release

```bash
# Critical bug fix
git checkout main
# Fix the bug
git commit -m "fix: critical bug in file processing"
git tag -a v1.0.1 -m "Hotfix v1.0.1 - Fix critical bug"
git push origin main v1.0.1
```

**When to use:**
- Critical bugs in production
- Security vulnerabilities
- Data loss or corruption issues

### Manual Release (Testing)

Via GitHub Actions UI:

1. Go to Actions → Release workflow
2. Click "Run workflow"
3. Enter version: `v1.0.0-test`
4. Check "Create as draft release"
5. Click "Run workflow"

**Use cases:**
- Testing release process
- Creating draft releases for review
- Emergency releases when tag push fails

## Troubleshooting

### Issue: Release workflow failed

**Check:**
1. View workflow logs in GitHub Actions
2. Common causes:
   - Tests failed → Fix tests, re-tag
   - Build failed → Check build configuration
   - Invalid version format → Use correct format (v1.2.3)

**Solution:**
```bash
# Delete failed tag
git tag -d v1.0.0
git push origin :refs/tags/v1.0.0

# Fix issue, re-tag
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

### Issue: Tag already exists

**Check:**
```bash
git tag -l v1.0.0
```

**Solution:**
```bash
# Delete local tag
git tag -d v1.0.0

# Delete remote tag
git push origin :refs/tags/v1.0.0

# Create new tag
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

### Issue: Wrong version tagged

**Solution:**
```bash
# Delete incorrect tag
git tag -d v1.0.0
git push origin :refs/tags/v1.0.0

# Tag correct version
git tag -a v1.0.1 -m "Release v1.0.1"
git push origin v1.0.1
```

### Issue: Release created but binaries missing

**Check:**
1. GitHub Actions workflow logs
2. Build step output
3. Asset upload step

**Solution:**
- Re-run the workflow from GitHub Actions UI
- Or delete release and tag, then re-create

### Issue: Tests pass locally but fail in CI

**Common causes:**
- Race conditions (use `go test -race`)
- Platform-specific issues
- Missing dependencies in CI

**Solution:**
1. Reproduce locally: `act -j test` (using nektos/act)
2. Fix issues
3. Push fixes and re-tag

## Rollback Procedure

If a release has critical issues:

### 1. Quick Mitigation

```bash
# Delete the problematic release from GitHub UI
# Users can still use previous version
```

### 2. Create Hotfix Release

```bash
# Checkout main
git checkout main

# Revert problematic changes (if needed)
git revert <commit-sha>

# Or fix the issue
# ... make fixes ...

# Create hotfix release
git tag -a v1.0.1 -m "Hotfix v1.0.1"
git push origin v1.0.1
```

### 3. Communicate

- Update release notes explaining the issue
- Post notice in GitHub Discussions
- If critical, notify users directly

## Best Practices

### DO

✅ Test thoroughly before releasing
✅ Follow semantic versioning strictly
✅ Write clear release notes
✅ Use pre-releases for major changes
✅ Keep `main` branch stable
✅ Create annotated tags (not lightweight)
✅ Monitor release workflow completion
✅ Verify artifacts after release

### DON'T

❌ Release without testing
❌ Skip version numbers
❌ Release from feature branches
❌ Delete old releases (keep history)
❌ Modify releases after publishing
❌ Use inconsistent version format
❌ Release with failing tests
❌ Forget to update documentation

## Release Checklist Template

Copy this for each release:

```markdown
## Release v1.X.X Checklist

### Pre-Release
- [ ] All tests pass locally
- [ ] Linter passes
- [ ] Integration tests verified
- [ ] Documentation updated
- [ ] Version number determined
- [ ] No uncommitted changes
- [ ] Main branch up-to-date

### Release
- [ ] Created annotated tag
- [ ] Pushed tag to GitHub
- [ ] Workflow completed successfully
- [ ] Release created on GitHub
- [ ] All artifacts present (5 binaries + checksums)

### Post-Release
- [ ] Tested `go install` command
- [ ] Downloaded and tested binary
- [ ] Release announcement posted
- [ ] Documentation links verified
- [ ] Monitoring for issues

### Notes
<!-- Add any notes about this release -->
```

## Automation

The release process is highly automated. The only manual steps are:

1. **Prepare code** (development, testing, documentation)
2. **Create tag** (`git tag -a vX.Y.Z -m "Release vX.Y.Z"`)
3. **Push tag** (`git push origin vX.Y.Z`)

Everything else is handled by GitHub Actions:
- Testing
- Building
- Checksums
- Changelog
- Release creation
- Asset uploads

## Questions?

- **Release process issues**: Check [GitHub Actions logs](https://github.com/OBerger96/sortTF/actions)
- **Version numbering questions**: See [SemVer spec](https://semver.org/)
- **General help**: Open a [GitHub Discussion](https://github.com/OBerger96/sortTF/discussions)

---

Happy releasing! 🚀
