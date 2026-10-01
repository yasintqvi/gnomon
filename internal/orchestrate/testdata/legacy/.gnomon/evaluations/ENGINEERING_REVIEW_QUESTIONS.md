# Engineering Review Questions

## Purpose

Define reusable engineering questions used to evaluate whether the current evaluation sufficiently covers important engineering scenarios.

Engineering Review produces findings.

It complements verification but does not redefine project knowledge.

---

# Authorization

- Can a privileged actor modify their own privileges?
- Can authorization be bypassed?
- Can authorization depend on client-controlled data?
- Can stale authorization survive permission or role changes?

---

# Authentication

- Can authentication be bypassed?
- Can revoked credentials remain usable?
- Can multiple authentication mechanisms create inconsistent behavior?

---

# Data Integrity

- Can concurrent operations violate consistency?
- Can partial failures leave persistent data inconsistent?
- Can duplicated operations create unintended side effects?
- Can orphaned or unreachable data be created?

---

# State Management

- Can an invalid state transition occur?
- Can conflicting state changes occur simultaneously?
- Can a valid state become permanently unreachable?

---

# Failure Handling

- What happens when a dependency becomes unavailable?
- Can retries introduce unintended side effects?
- Can failures be recovered safely?
- Are failures observable?

---

# Security

- Can privilege escalation occur?
- Can sensitive information be exposed?
- Can confidential data leak through errors or logs?
- Can trust decisions depend on client-controlled input?

---

# AI & Asynchronous Processing

- Can asynchronous execution produce inconsistent results?
- Can retries generate duplicate outputs?
- Can outdated AI outputs remain visible?
- Can failed background jobs leave incomplete artifacts?

---

# Maintainability

- Does the implementation introduce unnecessary complexity?
- Are responsibilities clearly separated?
- Is the design understandable?
- Will future modifications remain straightforward?

---

# Completeness

- Which important scenarios remain unevaluated?
- Which assumptions remain implicit?
- Which edge cases have not been explored?
- Which behaviors are undefined by current project knowledge?

---

# Review Result

Every question should produce one of the following outcomes:

- PASS
- DEFECT
- RISK
- KNOWLEDGE GAP
- NOT APPLICABLE

Supporting reasoning is mandatory.