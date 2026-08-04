# Bootstrap Workflow

## Purpose

Define the execution process for the **Initialize Project Environment** intent.

This workflow describes how the system should establish or complete a runnable project baseline from approved project knowledge before Specification-driven implementation begins.

It covers repository inspection, environment verification, scaffolding, configuration, dependency locking, and baseline validation without introducing business behavior.

---

## Intent

Initialize or complete the project's environment, frameworks, dependencies, infrastructure, and baseline structure using only approved project knowledge.

---

## When to Use

Use this workflow when:

* The repository is empty or does not yet contain a runnable project baseline.
* A required framework, runtime, dependency manifest, infrastructure service, or baseline subsystem has not been initialized.
* The existing project environment is incomplete and prevents later implementation work.
* The task is limited to environment setup, scaffolding, configuration, or baseline verification.

Do not use this workflow when:

* The required project baseline already exists and the requested work changes business behavior.
* The task implements or modifies behavior defined by a Specification.
* The task is limited to review, testing, refactoring, or architectural exploration.
* The requested change belongs to an existing feature implementation rather than project initialization.

Once the relevant baseline exists, subsequent behavior changes must follow the appropriate Workflow.

---

## Inputs

### User Input

Authorization to initialize or complete the project baseline, including any explicit task-specific constraints.

Examples may include:

* Target environment
* Operating system constraints
* Local port restrictions
* Containerization requirements
* Restricted tools or providers
* Scope limited to a specific subsystem

### Project Context

The current repository state, including existing files, configuration, dependency manifests, lockfiles, services, scaffolding, and user-owned changes.

### Required Knowledge

* `STACK.md`
* `ARCHITECTURE.md`
* `CONVENTIONS.md`
* Applicable ADRs
* Existing repository state

`STACK.md` owns approved technologies, runtime requirements, versions, tools, and environment constraints.

`ARCHITECTURE.md` and applicable ADRs own baseline structural and architectural decisions.

`CONVENTIONS.md` owns applicable naming and organization rules.

---

## Execution

### 1. Inspect

**Purpose**

Determine the current repository and environment state before making changes.

**Required Knowledge**

* Existing repository structure
* Existing manifests, lockfiles, configuration, and services
* `STACK.md`

**Expected Result**

A baseline gap analysis identifying:

* What already exists
* Which baseline components are missing
* Which existing files and user-owned changes must be preserved
* Which parts of the environment are incomplete
* Which detected issues are relevant to the authorized bootstrap scope

---

### 2. Verify

**Purpose**

Confirm that the knowledge required to perform the authorized bootstrap scope is sufficiently defined.

**Required Knowledge**

* `STACK.md`
* `ARCHITECTURE.md`
* Applicable ADRs
* User-provided constraints

**Expected Result**

A verified bootstrap plan containing:

* Approved runtimes
* Approved frameworks
* Approved tools and infrastructure
* Required versions and constraints
* Required baseline structure
* Blocking unresolved decisions
* Deferrable unresolved decisions
* The smallest coherent bootstrap scope

An unresolved item is blocking only when the authorized bootstrap scope cannot proceed correctly without resolving it.

Items outside the current scope should be reported as deferred rather than treated as blockers.

---

### 3. Scaffold

**Purpose**

Initialize the required frameworks, dependency manifests, and baseline project structure.

**Required Knowledge**

* Verified bootstrap plan
* `STACK.md`
* `ARCHITECTURE.md`
* `CONVENTIONS.md`
* Applicable ADRs

**Expected Result**

A baseline project structure that:

* Uses only approved technologies
* Respects architectural boundaries
* Follows applicable project conventions
* Preserves existing repository content
* Contains no Specification-driven business behavior
* Includes only what is necessary for the authorized bootstrap scope

---

### 4. Configure

**Purpose**

Configure the runtime environment and required baseline services.

**Required Knowledge**

* `STACK.md`
* Environment constraints
* Infrastructure-related ADRs
* Existing configuration

**Expected Result**

A working baseline environment with the required configuration for applicable services, runtimes, and subsystems.

Only approved and currently required services may be introduced.

Infrastructure intended solely for hypothetical future needs must not be added.

---

### 5. Lock

**Purpose**

Make dependency resolution reproducible.

**Required Knowledge**

* Dependency manifests
* Approved package managers
* `STACK.md` version policy

**Expected Result**

* Required lockfiles
* Reproducible dependency installation
* Recorded resolved versions
* No avoidable floating versions where locking is supported

---

### 6. Validate

**Purpose**

Confirm that the completed baseline is runnable, reproducible, and internally consistent.

**Required Knowledge**

* `STACK.md`
* Build configuration
* Testing configuration
* Applicable baseline verification requirements

**Expected Result**

Evidence that, where applicable:

* Dependencies install successfully.
* Required services start successfully.
* The project builds successfully.
* The application or subsystem starts successfully.
* Baseline checks pass.
* Baseline tests pass.
* No Specification-level behavior has been introduced.

---

## Rules

* Inspect the repository before making changes.
* Preserve existing user-owned files and changes.
* Work only within the explicitly authorized bootstrap scope.
* Use only technologies, tools, services, and versions approved by project knowledge.
* Do not introduce commonly paired tools or dependencies unless they are approved and required.
* Do not implement business rules, domain behavior, or Specification-driven functionality.
* Do not introduce structural or architectural decisions not defined by `ARCHITECTURE.md` or an approved ADR.
* Follow `CONVENTIONS.md` where it applies to baseline naming and organization.
* Distinguish blocking decisions from deferrable decisions.
* Do not treat out-of-scope or future subsystem decisions as bootstrap blockers.
* Do not report unrelated repository conditions as bootstrap issues.
* Record resolved dependency versions using the approved locking mechanism.
* Prefer the smallest runnable and verifiable baseline.
* Do not perform destructive operations without explicit authorization.
* Do not change project knowledge merely to justify an incorrect bootstrap implementation.

---

## Outputs

### Primary Output

A runnable and verifiable project baseline ready for later Specification-driven implementation.

### Supporting Outputs

* Required project scaffolding
* Dependency manifests
* Lockfiles
* Environment configuration
* Baseline service configuration
* Build and test configuration
* Verification evidence
* Explicitly reported blockers, deferred decisions, and limitations

---

## Failure Handling

### Missing Information

When required information is missing:

1. Identify the missing decision.
2. Determine which project document owns it.
3. Decide whether it blocks the authorized bootstrap scope.
4. Stop only the affected work when the decision is blocking.
5. Continue unaffected bootstrap work when possible.
6. Report deferrable decisions without resolving them.
7. Request clarification only when progress materially depends on the decision.

Do not select material technology, provider, tooling, or architecture defaults on the project's behalf.

---

### Conflicting Information

When project knowledge conflicts:

1. Identify the conflicting sources.
2. Apply the project's knowledge ownership rules.
3. Do not silently choose an interpretation.
4. Stop only the affected work when the conflict cannot be resolved within the Workflow's authority.
5. Continue unaffected work when possible.

---

### Existing Repository Conflict

When scaffolding or configuration would overwrite or invalidate existing content:

1. Stop the affected operation.
2. Identify the files or directories at risk.
3. Preserve existing user-owned changes.
4. Propose a non-destructive alternative when possible.
5. Require explicit authorization before replacement, migration, or deletion.

---

### Verification Failure

When installation, build, startup, or baseline checks fail:

1. Diagnose the failure.
2. Correct bootstrap-caused issues within scope.
3. Re-run the smallest relevant verification.
4. Expand verification when broader impact is indicated.
5. Report unresolved failures accurately.

Do not bypass valid checks, weaken constraints, or modify approved project knowledge merely to produce a passing result.

---

## Completion Criteria

The workflow is complete when:

* The authorized baseline scope has been established.
* Only approved technologies, versions, and structural decisions have been used.
* Required dependency manifests and lockfiles exist.
* Required baseline services and configuration are operational.
* Applicable build, startup, and baseline checks pass.
* Existing user-owned content has been preserved.
* No Specification-driven behavior has been implemented.
* Blocking issues have been resolved or explicitly reported.
* Deferrable decisions have been recorded without unnecessarily stopping progress.
* The resulting baseline is ready for the next appropriate Workflow.

---

## Workflow Constraints

* Bootstrap is limited to environment, scaffolding, configuration, dependency setup, and baseline verification.
* Bootstrap must not implement behavior owned by a Specification.
* Bootstrap may initialize only the project baseline or subsystem included in the authorized scope.
* Existing and operational subsystems should not be re-scaffolded without explicit justification.
* Prefer minimal, reversible, reproducible, and verifiable changes.
* Avoid speculative tooling and infrastructure intended only for possible future use.

---

## Notes

Bootstrap may be used at project inception or later when a required baseline subsystem does not yet exist.

Once the relevant baseline is operational, subsequent work must follow the Workflow corresponding to the new Intent.
