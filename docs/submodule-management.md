# Git Submodule Management

This document explains how git submodules work in this project and how to manage them for deterministic builds.

---

## Table of Contents

1. [Overview](#overview)
2. [How Submodule Pinning Works](#how-submodule-pinning-works)
3. [Daily Workflow](#daily-workflow)
4. [Updating the Submodule](#updating-the-submodule)
5. [Troubleshooting](#troubleshooting)

---

## Overview

This project uses a git submodule to include OpenAPI specifications from an external repository:

```
external/sdk/  <- Git submodule pointing to SDK repository
```

The submodule is **pinned to a specific commit** to ensure:
- ✅ **Deterministic builds** - Same code generates same output
- ✅ **Reproducibility** - Anyone can build from the same specs
- ✅ **Version control** - Changes to specs are tracked and reviewable
- ✅ **Stability** - Unexpected spec changes don't break builds

---

## How Submodule Pinning Works

### What is a Pinned Commit?

When you clone this repository, the `.gitmodules` file and `.git/modules/` directory contain a reference to a specific commit SHA in the submodule repository.

```bash
$ git submodule status
 a1b2c3d4e5f6 external/sdk (v1.2.3)
#^           ^            ^
#|           |            └─ Submodule path
#|           └─ Specific commit (pinned)
#└─ Status indicator (space = clean, + = modified, - = not initialized)
```

### Why Pinning Matters

**Without pinning (pulling latest):**
- Developer A builds today → Gets specs version X
- Developer B builds tomorrow → Gets specs version Y
- **Problem:** Different outputs, non-reproducible builds ❌

**With pinning (using fixed commit):**
- Developer A builds today → Gets specs at commit ABC123
- Developer B builds tomorrow → Gets specs at commit ABC123
- **Result:** Identical outputs, reproducible builds ✅

---

## Daily Workflow

### First Time Setup

```bash
# Clone the repository
git clone <repository-url>
cd openapi-go

# Initialize submodules (uses pinned commits)
git submodule update --init --recursive

# Alternatively, clone with submodules in one step
git clone --recurse-submodules <repository-url>
```

### Generate Clients (Normal Usage)

```bash
# This uses the pinned submodule commit
task generate

# What happens:
# 1. Initializes submodule if needed (to pinned commit)
# 2. Generates clients from the pinned specs
# 3. Deterministic and reproducible
```

**The generate task will NOT update the submodule to latest.** It uses whatever commit is currently pinned in the repository.

---

## Updating the Submodule

### When to Update

Update the submodule when you want to:
- Pick up new API specifications
- Include spec fixes or changes
- Sync with the latest SDK repository

### How to Update

#### Step 1: Update to Latest

```bash
# Update submodule to latest commit on main
task update-submodule

# Output:
# Updating SDK submodule to latest commit on main...
# ✓ Submodule updated. Review changes with 'git status' and commit if desired.
```

#### Step 2: Review Changes

```bash
# Check what changed
git status

# Output:
# modified:   external/sdk (new commits)

# See the diff
git diff external/sdk

# Output shows old commit → new commit
# See what specs changed
cd external/sdk
git log --oneline HEAD@{1}..HEAD  # commits since last update
git diff HEAD@{1} -- sdk-packages/  # actual spec changes
cd ../..
```

#### Step 3: Test the Changes

```bash
# Generate with new specs
task generate

# Run tests to ensure nothing broke
go test ./...

# Test the generated clients
# ... your validation steps ...
```

#### Step 4: Commit the New Pin

If everything looks good, commit the submodule update:

```bash
# Stage the submodule change
git add external/sdk

# Commit with descriptive message
git commit -m "chore: Update SDK submodule to include new API specs

- Updated funding API with new endpoints
- Fixed schema validation in holidays API
- Submodule: abc123..def456"

# Push the change
git push
```

**Important:** Once you commit and push this change, all developers will use the new pinned commit when they pull your changes and run `task generate`.

---

## Troubleshooting

### Submodule is Empty

**Problem:** The `external/sdk` directory is empty or missing files.

**Solution:**
```bash
# Initialize submodules
git submodule update --init --recursive

# Verify
ls external/sdk/sdk-packages/
```

### Submodule Has Local Changes

**Problem:** `git status` shows `modified: external/sdk` but you didn't change anything.

**Cause:** You or a script may have checked out a different commit in the submodule.

**Solution:**
```bash
# Reset submodule to pinned commit
git submodule update --force --recursive

# Or reset manually
cd external/sdk
git checkout .
git clean -fd
cd ../..
```

### Wrong Commit After Pull

**Problem:** After `git pull`, the submodule is at the wrong commit.

**Solution:**
```bash
# Update submodules after pulling
git submodule update --recursive

# This resets submodules to the commits pinned in the parent repo
```

### Accidental Update to Latest

**Problem:** You ran `cd external/sdk && git pull` and now have uncommitted changes.

**Solution:**
```bash
# Option 1: Reset to pinned commit
git submodule update --force --recursive

# Option 2: Keep the update and commit it
git add external/sdk
git commit -m "chore: Update SDK submodule"
```

### Merge Conflicts in Submodule

**Problem:** Git shows a merge conflict in `external/sdk`.

**Explanation:** This happens when:
- You updated the submodule to commit A
- Someone else updated it to commit B
- You're merging their branch

**Solution:**
```bash
# Accept their version (recommended if unsure)
git checkout --theirs external/sdk
git add external/sdk

# Or accept your version
git checkout --ours external/sdk
git add external/sdk

# Or manually choose a commit
cd external/sdk
git checkout <desired-commit-sha>
cd ../..
git add external/sdk

# Complete the merge
git commit
```

---

## Best Practices

1. **Never run `git pull` inside the submodule directory directly** during normal development
   - Use `task update-submodule` instead
   - This makes updates intentional and trackable

2. **Always review submodule changes before committing**
   - Check what specs changed
   - Test the generated output
   - Ensure tests pass

3. **Keep submodule updates in separate commits**
   - Don't mix submodule updates with code changes
   - Makes it easier to revert if needed

4. **Document why you're updating**
   - Mention what APIs changed
   - Include the commit range in the message
   - Example: `chore: Update SDK submodule (abc123..def456)`

5. **Use `task generate`, not manual commands**
   - Ensures consistent behavior
   - Automatically uses pinned commits
   - Prevents accidental updates

---

## Quick Reference

```bash
# Daily usage - Generate from pinned commit
task generate

# Update submodule to latest (intentional)
task update-submodule
git add external/sdk
git commit -m "chore: Update SDK submodule"

# Reset submodule to pinned commit
git submodule update --force --recursive

# See current submodule status
git submodule status

# See submodule history
cd external/sdk && git log --oneline && cd ../..
```

---

## CI/CD Considerations

In CI/CD pipelines:

```yaml
# ✅ Good - Uses pinned commit
- task generate

# ❌ Bad - Always pulls latest, non-deterministic
- task update-submodule
- task generate
```

The CI pipeline should **always use pinned commits** to ensure reproducible builds.

Only update submodules:
- Manually by developers
- Via automated daily/weekly jobs with review
- Never automatically in production builds

---

## Summary

- **Pinned commits** ensure deterministic, reproducible builds
- **`task generate`** uses the pinned commit (safe, deterministic)
- **`task update-submodule`** updates to latest (intentional, requires commit)
- **Always commit submodule updates** so others get the same specs
- **Review changes** before committing submodule updates

For questions or issues, consult this guide or ask the team!
