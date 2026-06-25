# `/githooks`

Git hooks.

## Hooks

- **pre-commit**: Runs formatting, vetting, and linting checks on staged files before each commit.
- **prepare-commit-msg**: Validates that the commit message follows the project's commit convention.
- **pre-push**: Runs the full test suite before pushing to remote.

## Setup

Run `make githooks` from the project root to activate all hooks.
