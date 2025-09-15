# Updating Documentation from Kuadrant Sources

This repository includes scripts to automatically pull and process documentation from the official Kuadrant repositories.

## Overview

The documentation update process:
1. Clones/updates the `kuadrant-operator` and `authorino` repositories
2. Extracts markdown documentation files
3. Optionally processes them into Go code for the MCP resources

## Scripts

### `update-docs.sh`
Main script that pulls the latest documentation from source repositories.

```bash
./update-docs.sh
```

This will:
- Clone or update the kuadrant-operator repo
- Clone or update the authorino repo
- Extract reference documentation for all policies
- Copy user guides and overviews
- Place extracted markdown in `extracted-docs/`

### `process-docs.go`
Optional Go script that converts extracted markdown into Go resource handlers.

```bash
go run process-docs.go
```

This generates `resources-generated.go` with handler functions that can be integrated into `resources.go`.

## Directory Structure

After running `update-docs.sh`:

```
extracted-docs/
├── authpolicy-reference.md
├── dnspolicy-reference.md
├── ratelimitpolicy-reference.md
├── tlspolicy-reference.md
├── tokenratelimitpolicy-reference.md
├── kuadrant-reference.md
├── overviews/
│   ├── auth.md
│   ├── dns.md
│   ├── rate-limiting.md
│   └── ...
├── user-guides/
│   ├── auth/
│   ├── dns/
│   ├── ratelimiting/
│   └── ...
└── authorino/
    ├── features.md
    ├── terminology.md
    └── user-guides/
```

## Updating MCP Resources

After extracting docs, you can either:

1. **Manual Integration**: Copy relevant content from `extracted-docs/` into `resources.go`
2. **Automated Generation**: Run `process-docs.go` and review/integrate the generated handlers

## Keeping Docs Current

Run periodically to keep documentation in sync:

```bash
# Full update
./update-docs.sh

# Generate Go code (optional)
go run process-docs.go

# Review and integrate changes
git diff resources.go
```

## Notes

- Temp directories (`.kuadrant-operator-temp/`, `.authorino-temp/`) are git-ignored
- `extracted-docs/` and `resources-generated.go` are also git-ignored
- The scripts pull from the `main` branch of source repos
- Documentation is truncated at 10KB per resource to keep size manageable