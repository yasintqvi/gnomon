# AI Software Engineering System

## Overview

A structured, document-driven system for AI-assisted software engineering.

The goal is to provide AI with a predictable workflow, clear responsibilities, and reliable knowledge sources throughout the software development lifecycle.

---

## Philosophy

* AI assists, humans decide.
* Documents are the source of truth.
* Every document has a single responsibility.
* Avoid duplicated knowledge.
* Prefer deterministic workflows over implicit reasoning.
* Evolve the system through real-world usage.

---

## Core Concepts

* Intent
* Workflow
* Specification
* Context
* Artifact Contract
* ADR
* Verification

---

## Directory Structure

```text
.ai-engineering/
├── context/
├── workflows/
├── specifications/
├── contracts/
├── decisions/
└── verification/
```

---

## Execution Flow

```text
User Prompt
    ↓
Intent
    ↓
Workflow
    ↓
Load Required Documents
    ↓
Reason
    ↓
Generate
    ↓
Verify
```

---

## Document Responsibilities

| Document         | Responsibility             |
| ---------------- | -------------------------- |
| PROJECT          | Project context            |
| DOMAIN           | Business knowledge         |
| ARCHITECTURE     | System structure           |
| CONVENTIONS      | Implementation consistency |
| STACK            | Technologies               |
| SPECIFICATION    | Business requirements      |
| CONTRACT         | Artifact behavior          |
| ADR              | Architectural decisions    |
| REVIEW CHECKLIST | Verification               |

---

## Knowledge Loading

The active Workflow determines which documents are required for a task.

Not every task requires every document.

---

## Design Principles

* Single source of truth
* Separation of concerns
* Explicit responsibilities
* Reusable workflows
* Reusable contracts
* Minimal duplication

---

## Repository Evolution

This system evolves incrementally.

New documents, workflows, or abstractions should only be introduced when they solve demonstrated problems.
