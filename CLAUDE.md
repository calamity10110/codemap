# 🛑 STOP — Run codemap before ANY task

```bash
codemap .                     # Project structure
codemap --deps                # How files connect
codemap --diff                # What changed vs main
codemap --diff --ref <branch> # Changes vs specific branch
```

## Required Usage

**BEFORE starting any task**, run `codemap .` first.

**ALWAYS run `codemap --deps` when:**
- User asks how something works
- Refactoring or moving code
- Tracing imports or dependencies

**ALWAYS run `codemap --diff` when:**
- Reviewing or summarizing changes
- Before committing code
- User asks what changed
- Use `--ref <branch>` when comparing against something other than main
