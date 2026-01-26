# Changesets Usage Guide

This directory is managed by the project’s release process and is used to track semantic version changes between releases.

## When to add a changeset

Add a new Markdown file in this folder whenever you change user-facing behavior—features, bug fixes, or other release-worthy modifications. Each file should contain a short description of the change and the version bump type (major, minor, or patch). Use human-readable names such as `add-windows-support.md` to make the intent clear.

## How to write a changeset

Each file follows this rough structure:

```
---
"cloudcompare-automation": minor
---

Introduce the new feature or fix in one or two sentences.
```

Replace the package name with the module(s) affected, adjust the bump type, and summarize the change in plain language.

## Cleaning up

The folder should be empty on `main` except for this README. Once the release automation consumes the pending changesets, their files are deleted automatically as part of the “version” commit.

Keeping this directory up to date ensures that CI can generate accurate release notes and version bumps with no manual bookkeeping.