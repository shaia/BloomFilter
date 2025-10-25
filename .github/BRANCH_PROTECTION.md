# Branch Protection Configuration

To ensure all releases go through PR approval, configure the following branch protection rules for the `main` branch.

## GitHub Settings Location

Navigate to: **Settings** → **Branches** → **Branch protection rules** → **Add rule** (for `main`)

## Required Settings

### 1. Basic Protection

- ✅ **Branch name pattern:** `main`
- ✅ **Require a pull request before merging**
  - ✅ Require approvals: **1** (or more)
  - ✅ Dismiss stale pull request approvals when new commits are pushed
  - ✅ Require review from Code Owners (optional)

### 2. Status Checks

- ✅ **Require status checks to pass before merging**
  - ✅ Require branches to be up to date before merging
  - Select required status checks:
    - `build` (from Go workflow)
    - Any other CI checks you want to enforce

### 3. Additional Restrictions

- ✅ **Require conversation resolution before merging**
- ✅ **Do not allow bypassing the above settings**
- ⚠️ **Include administrators** (recommended to enforce for everyone)

### 4. Tag Protection (Optional but Recommended)

Navigate to: **Settings** → **Tags** → **Protected tags** → **Add rule**

- ✅ **Tag name pattern:** `v*.*.*`
- ✅ **Restrict tag creation to administrators only**

This prevents unauthorized tag creation and ensures only maintainers can trigger releases.

## Configuration via Terraform (Optional)

If you manage GitHub settings as code:

```hcl
resource "github_branch_protection" "main" {
  repository_id = github_repository.repo.node_id
  pattern       = "main"

  required_pull_request_reviews {
    dismiss_stale_reviews           = true
    require_code_owner_reviews      = true
    required_approving_review_count = 1
  }

  required_status_checks {
    strict = true
    contexts = [
      "build",
    ]
  }

  enforce_admins                  = true
  require_conversation_resolution = true
  require_signed_commits          = false

  allows_deletions    = false
  allows_force_pushes = false
}

# Optional: Protect version tags
resource "github_repository_tag_protection" "version_tags" {
  repository = github_repository.repo.name
  pattern    = "v*"
}
```

## Verification

After setting up branch protection, verify it works:

1. Try pushing directly to main (should fail):
   ```bash
   git push origin main
   # Should see: "required status checks" or "pull request required"
   ```

2. Create a test PR and verify:
   - You cannot merge without approval
   - Tests must pass before merging
   - Tag creation requires appropriate permissions

## Alternative: CODEOWNERS File

Create `.github/CODEOWNERS` to specify who can approve releases:

```
# Default owners for everything
*       @yourusername

# Release-related files require approval from maintainers
/CHANGELOG.md           @yourusername @maintainer2
/.github/workflows/     @yourusername @maintainer2
/VERSIONING.md          @yourusername @maintainer2
```

## Troubleshooting

### Issue: "Can't merge even with approvals"

- Check if all required status checks are passing
- Ensure branch is up to date with main

### Issue: "Can still push tags directly"

- Configure tag protection rules
- Ensure "Include administrators" is enabled if you're an admin

### Issue: "Workflow has no permission to create PRs"

- Check repository settings → Actions → General
- Enable "Allow GitHub Actions to create and approve pull requests"
