# Release Process

This document describes how to create and publish releases for the Kuadrant MCP Server.

## Overview

The project uses GitHub Actions to automatically build and publish Docker images:
- Continuous deployment: Every merge to `main` updates the `:latest` tag
- Release versions: Creating a GitHub release publishes versioned tags

## Creating a Release

### 1. Choose Version Number

Follow semantic versioning (MAJOR.MINOR.PATCH):
- MAJOR: Breaking changes to tools/resources API
- MINOR: New features, backward compatible
- PATCH: Bug fixes, documentation updates

**Important**: Tags must be valid semver format:
- Correct: `0.2.0`, `1.0.0`, `v1.2.3`
- Wrong: `0.2`, `1.0`, `v1.2`

### 2. Create GitHub Release

#### Option A: Via GitHub UI
1. Go to [Releases](https://github.com/kuadrant/kuadrant-mcp-server/releases)
2. Click "Draft a new release"
3. Choose a tag (create new):
   - Format: `v1.2.3` or `1.2.3` or `0.1`
   - Both `v`-prefixed and plain versions work
4. Set release title (e.g., "v1.2.3 - Feature Name")
5. Add release notes:
   - New features
   - Bug fixes
   - Breaking changes
   - Credits
6. Click "Publish release"

#### Option B: Via GitHub CLI
```bash
# Create and push a tag
git tag -a v1.2.3 -m "Release v1.2.3"
git push origin v1.2.3

# Create release with notes
gh release create v1.2.3 \
  --title "v1.2.3 - Feature Name" \
  --notes "## What's Changed
- Feature X added
- Bug Y fixed
- Breaking: Removed Z"
```

### 3. Automated Publishing

Once the release is created, GitHub Actions will automatically:
1. Build multi-platform Docker images (linux/amd64, linux/arm64)
2. Push to GitHub Container Registry with tags:
   - `ghcr.io/kuadrant/kuadrant-mcp-server:1.2.3` (exact version)
   - `ghcr.io/kuadrant/kuadrant-mcp-server:1.2` (minor version)
   - `ghcr.io/kuadrant/kuadrant-mcp-server:1` (major version)

### 4. Verify Release

Check that images are published:
```bash
# Pull specific version
docker pull ghcr.io/kuadrant/kuadrant-mcp-server:1.2.3

# Verify it works
docker run --rm ghcr.io/kuadrant/kuadrant-mcp-server:1.2.3 --version
```

## Release Notes Template

```markdown
## What's Changed

### New Features
- Feature description (#PR)

### Bug Fixes
- Fix description (#PR)

### Documentation
- Docs update (#PR)

### Breaking Changes
- Change description and migration path

### Credits
- @contributor1 - Feature X
- @contributor2 - Bug fix Y

Full Changelog: https://github.com/kuadrant/kuadrant-mcp-server/compare/v1.2.2...v1.2.3
```

## Continuous Deployment

The `main` branch is continuously deployed:
- Every merge to `main` triggers a build
- Updates the `:latest` Docker tag
- Available immediately at `ghcr.io/kuadrant/kuadrant-mcp-server:latest`

Note: `:latest` may be unstable. For production use, pin to a specific version.

## Version Support

- Latest: Always tracks `main` branch (may be unstable)
- Major versions (e.g., `:1`): Latest release in that major version
- Minor versions (e.g., `:1.2`): Latest patch in that minor version
- Exact versions (e.g., `:1.2.3`): Immutable, specific release

## Troubleshooting

### Docker Build Failed
Check the [Actions tab](https://github.com/kuadrant/kuadrant-mcp-server/actions) for build logs.

### Wrong Version Published
- Tags are immutable once pushed
- To fix: Create a new patch version with the fix
- Never force-push or delete published tags

### Multi-platform Issues
The workflow builds for both `linux/amd64` and `linux/arm64`. If one fails:
1. Check architecture-specific dependencies
2. Test locally with: `docker buildx build --platform linux/arm64 .`

## Pre-release Checklist

Before creating a release:
- [ ] Update documentation if needed
- [ ] Run tests: `go test ./...`
- [ ] Build locally: `docker build -t test .`
- [ ] Test the Docker image: `docker run --rm test`
- [ ] Update CHANGELOG if maintained
- [ ] Check for security vulnerabilities: `go mod audit`
