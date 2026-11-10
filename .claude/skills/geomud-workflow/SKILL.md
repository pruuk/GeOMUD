---
name: geomud-workflow
description: Use before starting any slice, phase step, feature or fix in new_geo_mudbase that needs more than a one-line change. Gives the order (brainstorm, question, spec, Calabe reviews, plan, blind adversarial review, Calabe reviews, execute, optional contextual review) and who approves each step.
---

# Workflow (Calabe, ruled 2026-11-10)

Every piece of work goes through these steps in order. A step does not start until the one before
it is done; the two review gates are Calabe's. This is the process his other projects follow.

1. **Brainstorm** with Calabe (`superpowers:brainstorming`). Start from what exists: the phase
   design, its "Notes for later slices", the rulings, the inventories, and the source.
2. **Poke and question.** Challenge the idea before writing it down: edge cases, what a player
   sees, what the change locks in, conflicts with earlier rulings, the cheaper route. Ask Calabe only
   what is his to decide (global CLAUDE.md, "Working with Calabe" 3); settle the rest from the code.
3. **Write the spec.** A facts table verified against source first (global CLAUDE.md), then the
   design, then "Decisions taken without asking". Every design decision lives here, including
   interface shapes, guards, which slice a thing moves to, and player-visible rules. A phase design
   does not stand in for a slice spec: the slice spec builds on it and records where it departs.
4. **Calabe reviews the spec.** Open it with `code "<project folder>" "<file>"`. No plan until he
   approves it.
5. **Write the plan** from the approved spec (`superpowers:writing-plans`). The plan is execution
   only: tasks, code, commands, expected output. A design question found while planning goes back
   into the spec (and to Calabe if a player would notice it), not into the plan.
6. **Blind adversarial review of the plan,** repeated until satisfied. A fresh Opus subagent gets the
   spec, the plan and the repo, not your reasoning. It checks every facts-table row against source,
   challenges the design as well as the tasks, and where it can, applies the plan to a scratch copy
   and runs it. Fix what it finds and send a new blind reviewer; round two usually finds defects in
   round one's fixes. Stop when a round finds nothing material, and say how many rounds ran.
7. **Calabe reviews the plan.** No execution until he approves it.
8. **Execute** subagent-driven (`superpowers:subagent-driven-development`): a fresh agent per task,
   your own review of each task and of the whole branch, `geomud-verifying` before any claim,
   `geomud-git` for the branch and PR.
9. **Optional contextual review of the outcome:** for critical systems or changes with security
   implications. Recommend it when it applies and say why; Calabe decides.

## Where the documents go

- Spec and plan for work in progress sit at the repo root (`YYYY-MM-DD-<phase>-<slice>-spec.md`,
  `...-plan.md`), and move to `docs/completed/` when the work merges, as Phase 5's docs do.
- Review notes from step 6 go in the scratchpad unless a finding changes the spec, in which case the
  spec records it.

## Red flags

- Writing a plan, or a "Decisions" section in a plan, before a spec exists: stop, go back to step 1.
- A plan that adds a design decision the spec does not have: move it to the spec.
- Starting execution because the plan "looks done": Calabe approves it first.
