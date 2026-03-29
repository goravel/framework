---
name: create-update-pr
description: Create or update the pull request for the current branch and enforce a strict PR body format. Use this skill whenever the user asks to create a PR, update PR description, sync PR text with latest branch logic, or asks for Summary/Why style PR content.
---

# Create Or Update PR

## Purpose

Create a PR for the current branch when none exists, otherwise update the existing PR description from the latest diff.

---

## PR Body Specification

### Structure

The PR body must contain exactly these two top-level sections, in this order:

- `## Summary`
- `## Why`

No additional top-level sections are permitted.

### Template

````markdown
## Summary
- <behavior bullet 1>
- <behavior bullet 2>
- <optional behavior bullet 3>

<optional: Closes https://github.com/goravel/goravel/issues/1234>

## Why
<1–2 short paragraphs about what changed and why it matters.>

```go
<real user-facing code derived from test cases>
```

<1–2 short paragraphs about what was fixed and why it matters.>

```go
<real user-facing code derived before the fix from test cases, add comments to clarify what was wrong>

<real user-facing code derived after the fix from test cases, add comments to clarify the fix>
```
````

### Summary Rules

- Include exactly 2 to 3 bullet points.
- Bullets must describe behavior changes, not file-by-file edits.
- If an issue number is detected, append a standalone closing line after the bullets:
  `Closes https://github.com/goravel/goravel/issues/<number>`

### Why Rules

**All PR types**
- Include 1 to 2 short paragraphs explaining what changed and why.
- Code blocks must contain only real user-facing code; never include implementation snippets, placeholder markers, or pseudo-code inside code fences.

**Feature PRs**
- Include exactly one fenced code block generated from test cases.
- The block must reflect end-user usage semantics, not internal implementation detail.
- Prefer integration/black-box tests; fall back to the nearest relevant test case.

**Bug-fix PRs**
- Include one fenced code block per distinct bug.
- Describe before-fix vs. after-fix behavior in prose; code blocks contain only real user-facing code.

**Mixed PRs (feature + bug fix)**
- Use bug-fix formatting when multiple bug fixes are present; otherwise use feature formatting.

---

## Workflow

1. **Detect branch**
   - Run `git branch --show-current` → `branch`.
   - If empty, stop and report failure.

2. **Detect issue number**
   - Extract the first issue-like number from `branch` (e.g. `type/1234-description` → `1234`).
   - If found, set `closing_line = Closes https://github.com/goravel/goravel/issues/<number>`; otherwise omit it.

3. **Detect existing PR**
   - Run `gh pr view --json number,title,body,headRefName --head "$branch"`.
   - Success → existing PR. Not found → no PR.

4. **Build change context**
   - Compute base: `base=$(git merge-base HEAD origin/master)`.
   - Compute diff: `git diff "$base"...HEAD`.
   - Derive behavior-oriented summary bullets from the diff.
   - Detect PR type (`feature` or `bug fix`) from branch intent and diff semantics.
   - Derive code example(s) from test cases per the Why Rules above.

5. **Compose PR body**
   - Fill the PR Body Specification template using the context from step 4.

6. **Validate before applying**
   - Confirm the composed body satisfies all constraints in the PR Body Specification.
   - Ensure no placeholder markers (`<...>`) remain in the final body.

7. **Determine title**
   - Summarize from the logic changes, or fall back to the branch name.
   - Title must satisfy: https://github.com/goravel/.github/blob/master/.github/workflows/check_pr_title.yml

8. **Apply**
   - Existing PR: `gh pr edit <number> --body-file <tempfile>`
   - No PR: `gh pr create --title "<title>" --body-file <tempfile> --base master --head "$branch"`

9. **Verify**
   - Run `gh pr view --json number,title,url,body --head "$branch"`.
   - Confirm required sections and constraints are satisfied.

---

## Output Contract

Return:

- Action: `created` or `updated`
- PR number and URL
- Extracted issue number (or `none`)
- Final Summary bullets (2–3)
- Exact code example block(s) used in `## Why`

---

## Guardrails

- Never create duplicate PRs for the same branch.
- Never omit `## Summary` or `## Why`.
- Never reorder sections or add extra top-level headings.
- Prefer concise, deterministic output.
