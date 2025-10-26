# Versioning and Release Process

This project follows [Semantic Versioning](https://semver.org/) (SemVer) for version numbering.

**🔒 All releases require PR approval before tagging.**

## Version Format

`vMAJOR.MINOR.PATCH`

- **MAJOR**: Incompatible API changes (breaking changes)
- **MINOR**: New functionality in a backward-compatible manner
- **PATCH**: Backward-compatible bug fixes and performance improvements

## Creating a New Release

### Prerequisites

1. **Branch Protection**: Ensure `main` branch has protection rules enabled (see [`.github/BRANCH_PROTECTION.md`](.github/BRANCH_PROTECTION.md))
2. **PR Approval Required**: All changes to `main` must go through an approved pull request

### Option 1: Automated Release Preparation (Recommended)

Use the GitHub Actions workflow to prepare a release PR automatically:

1. **Navigate to Actions** → **Prepare Release** → **Run workflow**

2. **Fill in the form:**
   - **Version**: Enter version number (e.g., `0.1.0`, `1.0.0`)
   - **Release Type**: Select `patch`, `minor`, or `major`

3. **The workflow will automatically:**
   - ✅ Validate version format
   - ✅ Run all tests and build checks
   - ✅ Generate changelog from commits
   - ✅ Create a release branch (`release/vX.Y.Z`)
   - ✅ Update CHANGELOG.md
   - ✅ Create a Pull Request with release notes

4. **Review and merge the PR:**
   - Review the generated changelog
   - Get required approvals
   - Merge the PR to `main`

5. **Tag the release** (after PR is merged):
   ```bash
   git checkout main
   git pull origin main
   git tag v0.1.0
   git push origin v0.1.0
   ```

6. **The release workflow automatically:**
   - ✅ Verifies tag is on `main` branch
   - ✅ Checks PR was approved
   - ✅ Runs tests and builds
   - ✅ Creates GitHub Release
   - ✅ Notifies Go module proxy

### Option 2: Manual Release Process

1. **Create a feature branch** with your changes:
   ```bash
   git checkout -b release/v0.1.0
   ```

2. **Update CHANGELOG.md** with release notes

3. **Create a Pull Request** to `main`:
   - Title: `Release v0.1.0`
   - Include detailed release notes
   - List all changes and breaking changes

4. **Get PR reviewed and approved** by required reviewers

5. **Merge the PR** to `main`

6. **Tag the merged commit:**
   ```bash
   git checkout main
   git pull origin main
   git tag -a v0.1.0 -m "Release v0.1.0 - Initial SIMD-optimized Bloom Filter"
   git push origin v0.1.0
   ```

### Release Verification

After pushing the tag, the release workflow will:

1. ✅ **Verify** the tag is on the `main` branch
2. ✅ **Find** the associated PR and check for approvals
3. ✅ **Run** all tests (including SIMD correctness tests)
4. ✅ **Build** the project
5. ✅ **Generate** changelog from commits
6. ✅ **Create** a GitHub Release
7. ✅ **Notify** Go module proxy

You can monitor progress:

- [Actions tab](https://github.com/shaia/BloomFilter/actions)
- [Releases page](https://github.com/shaia/BloomFilter/releases)

Users can then install:

```bash
go get github.com/shaia/BloomFilter@v0.1.0
```

## Pre-release Versions

For beta/alpha releases, use the format:

```bash
git tag v0.2.0-beta.1
git tag v1.0.0-rc.1
```

## Examples

### Patch Release (Bug Fix)

1. Use "Prepare Release" workflow with version `0.1.1` and type `patch`
2. Review and approve the generated PR
3. Merge to `main`
4. Tag the release:

```bash
git checkout main
git pull origin main
git tag -a v0.1.1 -m "Fix memory leak in AVX2 operations"
git push origin v0.1.1
```

### Minor Release (New Feature)

1. Use "Prepare Release" workflow with version `0.2.0` and type `minor`
2. Review and approve the generated PR
3. Merge to `main`
4. Tag the release:

```bash
git checkout main
git pull origin main
git tag -a v0.2.0 -m "Add AVX-512 SIMD support"
git push origin v0.2.0
```

### Major Release (Breaking Change)

1. Use "Prepare Release" workflow with version `1.0.0` and type `major`
2. Review changelog, ensure breaking changes are documented
3. Get approval from required reviewers
4. Merge to `main`
5. Tag the release:

```bash
git checkout main
git pull origin main
git tag -a v1.0.0 -m "Stable release with new API"
git push origin v1.0.0
```

## Best Practices

1. **Always test before tagging**: Ensure all tests pass locally
   ```bash
   go test -v ./...
   go test -v -run=TestSIMDCorrectness
   ```

2. **Write descriptive tag messages**: Summarize key changes
   ```bash
   git tag -a v0.1.0 -m "Initial release
   - AVX2 SIMD optimizations
   - Cache-aligned storage
   - Cross-platform support (AMD64/ARM64)"
   ```

3. **Update documentation**: Ensure README and docs reflect new features

4. **Keep a CHANGELOG**: Consider maintaining a CHANGELOG.md file

## Troubleshooting

### Delete a Tag (if needed)

```bash
# Delete local tag
git tag -d v0.1.0

# Delete remote tag
git push origin --delete v0.1.0
```

### List All Tags

```bash
# List all tags
git tag -l

# List tags with messages
git tag -n
```

### View Changes Since Last Tag

```bash
# See what's changed since the last tag
git log $(git describe --tags --abbrev=0)..HEAD --oneline
```

## Go Module Users

Users can install specific versions:

```bash
# Latest version
go get github.com/shaia/BloomFilter

# Specific version
go get github.com/shaia/BloomFilter@v0.1.0

# Specific commit (for testing)
go get github.com/shaia/BloomFilter@commit-hash

# Latest of a major version
go get github.com/shaia/BloomFilter@v1
```
