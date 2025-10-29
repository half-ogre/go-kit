# @half-ogre's Go Kit
This is my personal "kit" for Go. Put another way, it's full of the stuff I use in just about all of my Go projects, bundled up into one convenient package for sharing.

## Packages

- **actionskit** - GitHub Actions utilities
- **pgkit** - PostgreSQL migration library
- **dynamodbkit** - AWS DynamoDB helpers
- **echokit** - Echo web framework utilities
- **envkit** - Environment variable helpers
- **ginkit** - Gin web framework utilities
- **kit** - Core utilities
- **logkit** - Logging utilities
- **versionkit** - Version management

## CLI Tools

- **pgkit** - PostgreSQL toolkit CLI with migrate, create, and drop commands. See [cmd/pgkit/README.md](cmd/pgkit/README.md) for details. Install with `make install-pgkit`.

## Versioning

This repository uses [Semantic Versioning (SemVer)](https://semver.org/) for versioning. Each release will be tagged with its full version (e.g., `v1.2.3`). The latest release of each major version will also be tagged with `v{Major}` (e.g., `v1`) and that tag will move to the latest version as new versions are released.

## License

This project is licensed under the terms found in the [LICENSE.md](LICENSE.md) file.

## Support

**No Support Guarantee:** This repository is provided as-is without any warranty or support guarantee, as outlined in the [LICENSE.md](LICENSE.md). This is a personal toolkit shared for convenience. While issues and pull requests are welcome, please note that this is primarily maintained for personal use and there are no guarantees of support or maintenance.
