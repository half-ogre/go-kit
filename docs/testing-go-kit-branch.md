# Testing Unreleased go-kit Features

This guide explains how to test features from a go-kit branch before they've been released to main.

## Option 1: Using go get with a Specific Branch

The simplest approach is to use `go get` to fetch a specific branch:

```bash
go get github.com/half-ogre/go-kit@branch-name
```

For example, to test features from a branch called `feature/pgkit`:

```bash
go get github.com/half-ogre/go-kit@feature/pgkit
```

After running this command:
- Your `go.mod` will reference the latest commit from that branch
- Run `go mod tidy` to ensure all dependencies are updated

### Using a Specific Commit

For more stability, you can reference a specific commit hash:

```bash
go get github.com/half-ogre/go-kit@abc1234
```

Where `abc1234` is the commit hash (can be shortened to 7+ characters).

## Option 2: Using the replace Directive

For more control, especially during development, use a `replace` directive in your `go.mod`:

```go
module your-module

go 1.24

require (
    github.com/half-ogre/go-kit v0.0.0
    // ... other dependencies
)

replace github.com/half-ogre/go-kit => github.com/half-ogre/go-kit branch-name
```

Then run:

```bash
go mod tidy
```

### Using a Local Clone

If you're actively developing go-kit features, you can use a local clone:

```go
replace github.com/half-ogre/go-kit => /path/to/your/local/go-kit
```

This is useful when:
- Making changes to go-kit and testing them immediately
- Debugging issues in go-kit
- Contributing features back to go-kit

## Verifying the Branch

To confirm which version you're using, check `go.mod`:

```bash
cat go.mod | grep go-kit
```

Or use `go list`:

```bash
go list -m github.com/half-ogre/go-kit
```

This will show the exact commit or branch being used.

## Reverting to Latest Release

When you're done testing and want to return to the latest released version:

### If you used go get:

```bash
go get github.com/half-ogre/go-kit@latest
go mod tidy
```

### If you used replace:

Remove the `replace` directive from `go.mod`, then:

```bash
go mod tidy
```

### Bypassing the Module Proxy

When reverting to the latest version, Go may use a cached version from the module proxy. To ensure you get the latest version directly from the repository, bypass the proxy:

```bash
GOPROXY=direct go get github.com/half-ogre/go-kit@latest
go mod tidy
```

This is especially useful when:
- A new release was just published and isn't yet in the proxy cache
- You need to verify you're getting the most recent release
- You're testing immediately after a version tag is pushed

## Best Practices

1. **Pin to a commit** - For reproducible builds, use specific commit hashes rather than branch names
2. **Document the reason** - Add a comment in `go.mod` explaining why you're using a non-main version:
   ```go
   // Testing pgkit connection pooling features before v1.2.0 release
   replace github.com/half-ogre/go-kit => github.com/half-ogre/go-kit@feature/pgkit
   ```
3. **Track the branch** - Keep a note of which features you're testing and when they're expected to be released
4. **Update regularly** - If testing a branch for an extended period, periodically update to the latest commit:
   ```bash
   go get github.com/half-ogre/go-kit@branch-name
   go mod tidy
   ```
5. **Clean up** - Remove `replace` directives before committing to production

## Troubleshooting

### "Module not found" error

If you get a module not found error, ensure the branch exists:

```bash
git ls-remote https://github.com/half-ogre/go-kit refs/heads/branch-name
```

### Dependency conflicts

If you encounter dependency conflicts:

```bash
go get -u github.com/half-ogre/go-kit@branch-name
go mod tidy
```

### Cached modules

If changes aren't appearing, clear the Go module cache:

```bash
go clean -modcache
go get github.com/half-ogre/go-kit@branch-name
```

## Example Workflow

Here's a complete workflow for testing a new feature:

```bash
# 1. Reference the feature branch
go get github.com/half-ogre/go-kit@feature/pgkit

# 2. Update dependencies
go mod tidy

# 3. Run your tests
go test ./...

# 4. When done, revert to latest release (bypassing proxy to get latest)
GOPROXY=direct go get github.com/half-ogre/go-kit@latest
go mod tidy
```
