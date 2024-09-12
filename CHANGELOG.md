# Intercept Changelog

All notable changes to this project will be documented in this file.


## Unreleased - feature/loadremote

**Commit**: [e95c0ed](https://github.com/xfhg/intercept/commit/e95c0ed)

**Branch** [feature/loadremote](https://github.com/xfhg/intercept/tree/feature/loadremote)

**Summary**: Capability to load the main policy file from remote endpoint (and check their SHA256)

### Breaking 
- None

### Added
- Added this CHANGELOG
- Added shorthand for policy (-p)
- Added shorthand for tag filtering "tags_any" (-f)
- Added sha256 checksum on command line for policy (--checksum)
- INTERCEPT can now load a remote policy (ex: https://raw.githubusercontent.com/xfhg/intercept/master/playground/policies/test_scan.yaml)
- INTERCEPT can verify the checksum of remote policies

### Changed
- Modified go build version to 1.23

### Removed
- None

