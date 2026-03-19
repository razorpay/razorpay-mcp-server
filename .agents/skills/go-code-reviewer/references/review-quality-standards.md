# Code Review Quality Standards

## Purpose

This document defines what makes a "good" code review based on research from Google's Engineering Practices, academic studies, and industry best practices. Use this as a quality gate for the go-code-reviewer skill itself.

## Source of Truth

Primary source: [Google's Code Review Developer Guide](https://google.github.io/eng-practices/review/)

## The Golden Rule

**"Reviewers should favor approving a CL once it is in a state where it definitely improves the overall code health of the system being worked on, even if the CL isn't perfect."**

Code review is about **continuous improvement**, not **perfection**.

---

## What Makes a GOOD Code Review

### 1. **Focus on Code Health, Not Perfection**

✅ **Good**: "This change improves the codebase. The minor naming issue is a nit."
❌ **Bad**: "I won't approve until you rename all variables to my preferred style."

**Principle**: Does this change make the codebase better than before? If yes → approve.

### 2. **Balance Progress with Quality**

✅ **Good**: Approve improvements quickly, even if not perfect
❌ **Bad**: Make it "very difficult for any change to go in"

**Principle**: Enable developer velocity while maintaining quality standards.

### 3. **Technical Facts Over Personal Opinions**

✅ **Good**: "This approach has O(n²) complexity. Consider using a map (O(n))."
❌ **Bad**: "I prefer tabs over spaces." (when style guide says spaces)

**Principle**: Technical facts and data overrule opinions and personal preferences.

### 4. **Educational, Not Gatekeeping**

✅ **Good**: "Good use of channels! FYI, you can also use sync.WaitGroup here."
❌ **Bad**: "You should know this already. Figure it out."

**Principle**: Share knowledge to improve long-term code health.

### 5. **Distinguish Blocking vs Non-Blocking Feedback**

✅ **Good**: "🚨 Critical: Transaction context bug. 💡 Nit: Consider renaming."
❌ **Bad**: All feedback treated as equally important, blocking approval

**Principle**: Use severity markers (Critical/Important/Optional, or Nit:) clearly.

### 6. **Constructive and Specific**

✅ **Good**: "Add nil check at line 42 before dereferencing payment.ID"
❌ **Bad**: "This code is bad. Rewrite it."

**Principle**: Actionable feedback with specific suggestions.

### 7. **Review Every Line**

✅ **Good**: Actually read the code, don't just skim
❌ **Bad**: Auto-approve or rubber-stamp without reading

**Principle**: Code health depends on thorough review.

---

## What to Look For (Google's Checklist)

Reviewers should examine these areas **in order of importance**:

### 1. **Design** (Most Important)
- Do the interactions of various pieces make sense?
- Does this change belong in the codebase or in a library?
- Does it integrate well with the rest of the system?
- Is now a good time to add this functionality?

### 2. **Functionality**
- Does the code do what the developer intended?
- Is what the developer intended good for users?
- What edge cases exist? Are they handled?
- Could this cause issues in production?

### 3. **Complexity**
- Is the code more complex than it needs to be?
- Can another developer understand this code quickly?
- Over-engineering: Does this solve future problems we don't have yet?

### 4. **Tests**
- Are there appropriate automated tests?
- Will the tests actually catch bugs when the code breaks?
- Do tests cover edge cases?

### 5. **Naming**
- Did the developer choose clear names for variables, classes, methods?
- Are names too long? Too short? Misleading?

### 6. **Comments**
- Do comments explain **why** (not what)?
- Is the "what" obvious from the code itself?
- Are there unnecessary comments that should be removed?

### 7. **Style & Consistency**
- Does code follow the style guide?
- Is it consistent with surrounding code?
- Don't block on style if it follows the guide (even if not your preference)

### 8. **Documentation**
- Are relevant docs updated (README, API docs, migration guides)?
- Is the documentation clear and accurate?

### 9. **Every Line**
- Did you review every line of code?
- Are you confident this change improves code health?

---

## What Makes a BAD Code Review

### 1. **Perfectionism**
❌ Delaying approval over minor polish issues
❌ Blocking on subjective preferences not in style guide
❌ Demanding rewrites for working code that's "not ideal"

### 2. **Opinion-Driven**
❌ "I would have done it differently" (without technical justification)
❌ Rejecting valid alternatives based on personal taste
❌ Style preferences that contradict the style guide

### 3. **Gatekeeping**
❌ Making reviews so difficult that developers avoid contributing
❌ Treating reviews as opportunities to demonstrate superiority
❌ Nitpicking instead of teaching

### 4. **Superficial**
❌ Auto-approving without reading (rubber stamping)
❌ Only checking formatting/style, missing logic bugs
❌ Skimming instead of reading every line

### 5. **Unresolved Conflicts**
❌ Letting CLs stall indefinitely without escalation
❌ Endless back-and-forth without reaching consensus
❌ Not bringing in senior engineers when stuck

### 6. **Unhelpful**
❌ "This is wrong" without explaining why or how to fix
❌ No actionable feedback
❌ Vague complaints ("this feels messy")

---

## Self-Assessment Checklist for Reviews

Use this to evaluate if your review meets quality standards:

### Progress & Velocity
- [ ] Did I approve improvements even if not perfect?
- [ ] Did I enable developer progress without sacrificing quality?
- [ ] Did I avoid perfectionism and gatekeeping?

### Technical Rigor
- [ ] Did I review every line of code?
- [ ] Did I base feedback on technical facts, not opinions?
- [ ] Did I verify the code improves overall code health?

### Feedback Quality
- [ ] Is my feedback constructive and specific?
- [ ] Did I distinguish blocking (Critical/Important) vs non-blocking (Optional/Nit)?
- [ ] Did I provide actionable suggestions?

### Educational Value
- [ ] Did I share knowledge to help the developer improve?
- [ ] Did I explain the "why" behind my feedback?
- [ ] Did I recognize good patterns when I saw them?

### Coverage
- [ ] Design: Does this fit well in the codebase?
- [ ] Functionality: Does it work correctly? Edge cases handled?
- [ ] Complexity: Is it unnecessarily complex?
- [ ] Tests: Are there appropriate tests?
- [ ] Naming: Are names clear?
- [ ] Comments: Do they explain why, not what?
- [ ] Style: Does it follow the guide?
- [ ] Docs: Are relevant docs updated?

### Review Experience
- [ ] Would I want to receive this review if I were the author?
- [ ] Does this review help the team improve collectively?
- [ ] Is the tone respectful and collaborative?

---

## Common Pitfalls to Avoid

### 1. **Scope Creep in Review**
❌ "While you're here, can you also refactor this unrelated code?"
✅ Keep reviews focused on the current change

### 2. **Bike-Shedding**
❌ Spending more time on trivial issues (variable names) than critical bugs
✅ Prioritize feedback by impact on code health

### 3. **Analysis Paralysis**
❌ Endless discussion of theoretical edge cases
✅ Ship improvements, iterate based on real issues

### 4. **Ignoring Context**
❌ Applying standards without understanding constraints
✅ Consider deadlines, team experience, system criticality

### 5. **Not Reading Carefully**
❌ Missing critical bugs while commenting on formatting
✅ Review every line, prioritize correctness over style

---

## Metrics for Review Quality

### Quantitative Metrics
- **Review turnaround time**: <24 hours for typical PR
- **Approval rate**: >70% approved with minor/no changes (if <50%, too strict)
- **Bug escape rate**: <5% of bugs found in production were in reviewed code
- **Review length**: <300 lines for typical PR, <400 for large PRs

### Qualitative Metrics
- **Developer satisfaction**: Do developers feel reviews are helpful?
- **Knowledge transfer**: Are developers learning from reviews?
- **Code health trend**: Is the codebase improving over time?
- **False positive rate**: <10% of flagged issues are not actually issues

---

## Application to go-code-reviewer Skill

The go-code-reviewer skill should embody these standards by:

1. **Focusing on code health**: Flag real issues, not perfectionist nitpicks
2. **Balancing velocity**: Quick Layer 1 fail-fast gates, focused Layer 3 review
3. **Technical rigor**: Use references (Google, Uber, Razorpay) as source of truth
4. **Clear severity**: 🚨 Critical (P0), ⚠️ Important (P1), 💡 Optional (P2)
5. **Constructive feedback**: Specific suggestions with line numbers
6. **Educational**: Explain why patterns are important, reference docs
7. **Comprehensive coverage**: Design → Functionality → Complexity → Tests → Style
8. **Respectful tone**: Professional, collaborative, no gatekeeping
9. **Inline comments**: 3-5 specific comments on actual code lines
10. **Concise output**: <300 lines, avoid redundancy

---

## References

- [Google Code Review Developer Guide](https://google.github.io/eng-practices/review/)
- [The Standard of Code Review](https://google.github.io/eng-practices/review/reviewer/standard.html)
- [What to Look For in Code Review](https://google.github.io/eng-practices/review/reviewer/looking-for.html)
- [How Google Takes Pain Out of Code Reviews](https://read.engineerscodex.com/p/how-google-takes-the-pain-out-of)
- [Software Engineering at Google - Chapter 9: Code Review](https://abseil.io/resources/swe-book/html/ch09.html)

---

## Version History

- v1.0 (2026-02-05): Initial standards based on Google Engineering Practices
