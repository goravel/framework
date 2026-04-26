---
name: solve-pr-comments
description: >
  Workflow for addressing pull request review comments. Covers when to change
  code, when to push back on incorrect feedback, and when to answer questions
  without modifying code. Use this skill whenever the user asks to resolve,
  address, or respond to PR review comments.
---

# Solve PR Comments

A PR review comment is not an order — it is input. Your job is to evaluate
each comment with the same rigor you apply to code, then take the right action:
change, decline, or answer.

## Decision tree

For every comment, determine which category it falls into before acting:

```
Is the comment technically correct?
├── No  → Decline and explain (no code change)
└── Yes → Is it a question or a request for clarification?
          ├── Yes → Answer in a reply (no code change)
          └── No  → Apply the change
```

---

## 1. Apply the change

Change code when the comment identifies a real problem: a bug, a violation of
project conventions, a clarity issue, or a missed edge case.

**Rules:**
- Apply the minimal fix that addresses the concern. Do not refactor unrelated code.
- After applying, reply to the comment with a one-sentence summary of what changed.
- If the fix is non-trivial, briefly explain the approach taken.

---

## 2. Decline and explain

Not every comment is correct. Push back when:

- The suggestion introduces a bug or regression.
- The suggestion contradicts an established pattern in this codebase (point to the location).
- The suggestion is a matter of style preference with no objective benefit, and the existing code
  already follows a consistent convention.
- The suggestion is based on a misunderstanding of the code's intent.

**Rules:**
- Always reply with a clear, respectful explanation. Never silently ignore a comment.
- State the specific reason: wrong behavior, conflicts with `file:line`, performance trade-off, etc.
- If appropriate, offer a counter-proposal.
- Do not change the code to pacify a reviewer when you believe the original is correct.

**Example reply:**
> This change would bypass the nil-check on line 42 and cause a panic when the
> cache is cold. The current guard is intentional — I'd rather keep it as-is.
> Happy to add a comment if the intent is unclear.

---

## 3. Answer the question

Some comments are genuine questions: "Why did you do X?", "What does this return
when Y?", "Is this thread-safe?". These need an answer, not a code change.

**Rules:**
- Reply with a direct answer.
- If the question reveals that the code is genuinely confusing, consider adding a
  comment or renaming — but only if it actually improves clarity, not just to
  satisfy the reviewer.
- If you only answer the question, do not mark the thread resolved. Leave it open
  for the reviewer to decide whether the answer resolves their concern.
- Never change code solely to signal that you read the comment.

---

## Workflow

1. **Fetch comments**
   ```bash
   gh pr view {pr_number} --json comments,reviewThreads
   # or for inline review comments:
   gh api repos/{owner}/{repo}/pulls/{pr_number}/comments
   ```

2. **Triage each comment** using the decision tree above.

3. **Group related changes** — if multiple comments touch the same file or
   function, batch the edits together before replying.

4. **Reply to every comment** — resolved or not. Leaving a comment without a
   reply signals that it was missed.

5. **Mark actioned threads resolved** after replying, using the GitHub GraphQL
   `resolveReviewThread` mutation (requires the thread node ID). Resolve only
   when you applied a code/docs change, or when you are explicitly declining an
   incorrect request with a clear explanation. Do not resolve genuine question
   threads when your only action was to reply with an answer.
   ```bash
   gh api graphql -f query='
     mutation {
       resolveReviewThread(input: {threadId: "<thread_node_id>"}) {
         thread { isResolved }
       }
     }'
   ```
   To get thread node IDs (type `PRRT_...`, not comment IDs):
   ```bash
   gh pr view {pr_number} --json reviewThreads --jq '.reviewThreads[].id'
   ```
   Alternatively, resolve threads directly in the GitHub UI.

6. **Push the updated branch** if any code changed:
   ```bash
   git push
   ```

---

## Guardrails

- Never mark a comment as resolved without either applying the change or
  explaining why you did not.
- Never mark a genuine question thread resolved when the only action was an AI
  reply. The reviewer owns closure for answer-only threads.
- Never apply a change you believe is wrong just to close a thread.
- Never reply with vague acknowledgements ("Sure!", "Done") — every reply
  must state what was done or why nothing was done.
- Never open a PR comment thread with a different concern from the original;
  raise separate issues separately.
