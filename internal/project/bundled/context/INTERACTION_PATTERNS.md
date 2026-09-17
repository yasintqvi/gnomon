# Interaction Patterns

## Purpose

This document defines the shared interaction behavior of the product.

It is the authoritative project knowledge for:

* feedback behavior;
* confirmations;
* loading and progress behavior;
* asynchronous interactions;
* form interaction;
* destructive-action behavior;
* collection interaction;
* search;
* filtering;
* sorting;
* pagination;
* editor behavior;
* navigation interaction;
* transient and persistent feedback;
* responsive interaction behavior;
* interaction accessibility;
* common recovery patterns.

Its purpose is to ensure that independently implemented features behave consistently and predictably.

A feature is not interaction-complete merely because its backend behavior works.

---

# Ownership Boundaries

This document owns shared interaction behavior.

PRODUCT_EXPERIENCE.md owns product organization, journeys, discoverability, and information hierarchy.

UI_FOUNDATION.md owns visual presentation, responsive containment, themes, and formatting.

Specifications own business behavior, authorization, business states, feature-specific flows, and acceptance criteria.

This document must not invent new business capabilities.

---

# Interaction Principles

## Immediate Feedback

Every meaningful user action should produce understandable feedback.

Users should understand whether an action:

* was received;
* is pending;
* succeeded;
* failed;
* requires further input.

## Preserve Context

Interactions should preserve relevant user context whenever practical.

## Prevent Accidental Harm

Destructive or materially significant actions should require stronger intent than ordinary actions.

## Progressive Disclosure

Reveal secondary controls and information when relevant without hiding primary tasks.

## Responsive Interaction

Primary interactions must remain operable across supported viewport sizes.

## Accessible by Default

Shared interactions should support keyboard operation, focus management, understandable labels, and accessible state communication.

---

# Feedback Model

Use three conceptual levels:

1. Inline Feedback
2. Persistent Contextual Feedback
3. Toast Notification

---

# Inline Feedback

Use inline feedback when the message belongs directly to a field, control, or local interaction.

Do not use Toast as the primary mechanism for validation errors.

---

# Persistent Contextual Feedback

Use persistent feedback when the information remains relevant after an interaction.

Examples include:

* durable failures;
* unavailable states;
* stale or outdated state;
* partial completion;
* missing prerequisites.

Do not replace durable state with short-lived feedback.

---

# Toast Notifications

Use Toast for transient confirmation or information where resulting state is otherwise visible.

Do not use Toast for:

* required decisions;
* blocking failures;
* validation;
* long-running operation state;
* information users must remember.

All features should use the shared Toast mechanism.

---

# Confirmation Pattern

Destructive and materially significant actions should use the shared confirmation pattern when confirmation is required.

Browser-native alert and confirm interactions should not be used for normal product workflows.

The confirmation should communicate:

* action;
* affected resource;
* consequence;
* reversibility where relevant;
* confirm action;
* cancel action.

Use action-specific labels rather than vague confirmation labels.

---

# Confirmation Execution

Protected actions must not execute before confirmation completes.

Cancel must leave the protected resource unchanged.

Repeated confirmation must not create unsafe duplicate submissions.

Pending execution should be visible.

Focus should be managed appropriately.

Exact business consequences remain owned by Specifications.

---

# Unsaved Changes

Where substantial user work can be lost, use an approved protection strategy such as:

* navigation guard;
* unsaved-change confirmation;
* approved autosave;
* draft preservation.

Do not introduce autosave without supporting project knowledge.

---

# Loading Model

Distinguish:

* initial page loading;
* local content loading;
* action loading;
* background processing.

Do not represent every loading state with the same generic interaction.

---

# Initial Page Loading

Prefer structural loading states when destination structure is known.

Avoid unnecessary blank screens and disruptive layout shifts.

---

# Local Action Loading

While an action is pending:

* prevent unsafe duplicate submission;
* communicate pending state;
* preserve surrounding context;
* keep the triggering action understandable.

---

# Background Processing

Long-running operations should remain represented after the initiating request completes.

The user should understand:

* current state;
* available actions;
* completion;
* failure;
* recovery where available.

---

# Async Status Refresh

Known asynchronous work should update without requiring users to guess that manual refresh is necessary.

Use the simplest mechanism consistent with approved architecture and stack knowledge.

Do not introduce infrastructure merely for interaction convenience without justification.

---

# Success Behavior

After successful actions:

* update visible state;
* preserve context;
* provide transient confirmation where useful;
* expose the next meaningful action where applicable.

---

# Error Behavior

Interaction should distinguish user-relevant error categories such as:

* validation;
* conflict or stale state;
* permission failure;
* missing prerequisite;
* processing failure;
* unavailable dependency;
* unexpected failure.

Technical details should not replace user-meaningful explanations.

---

# Forms

Shared form behavior should cover:

* submission;
* pending state;
* duplicate prevention;
* validation;
* field errors;
* required fields;
* help text;
* disabled state;
* preservation of entered values where appropriate.

Server-side behavior remains authoritative.

---

# Long Forms

Long forms should use meaningful grouping.

Possible patterns include:

* sections;
* tabs;
* progressive disclosure;
* contextual panels.

Do not introduce multi-step workflows solely to shorten a page.

---

# Structured Text Editing

Define when substantial structured text requires a richer editing surface instead of a plain Textarea.

If Markdown is selected by project knowledge as the shared structured-text format, define:

* supported formatting;
* read/edit consistency;
* preview behavior;
* responsive behavior;
* theme behavior;
* writing-direction support;
* validation;
* sanitization expectations.

Do not assume Markdown is required for every project.

---

# Navigation Interaction

Define shared behavior for:

* breadcrumbs;
* workspace navigation;
* compact navigation;
* parent navigation;
* overlay-based navigation.

Navigation interactions should preserve orientation and context.

---

# Collections

Growing collections should use interaction mechanisms appropriate to their task and supported backend capabilities.

Potential mechanisms include:

* search;
* filtering;
* sorting;
* pagination.

Not every collection requires every mechanism.

---

# Search

Search should:

* target meaningful fields;
* preserve applicable collection state;
* handle no-results intentionally;
* avoid fake client-only search over incomplete server-side datasets.

Define the project's shared search execution pattern where applicable.

---

# Filtering

Filters must correspond to supported data attributes.

Users should be able to understand active filters and clear them appropriately.

Do not invent unsupported filters for visual completeness.

---

# Sorting

Sorting should exist only where it materially helps the task and is correctly supported.

Applicable search and filter state should be preserved.

---

# Pagination

Large server-backed collections should use an appropriate bounded navigation strategy.

Where pagination is used, it should:

* preserve search;
* preserve filters;
* preserve sorting;
* communicate position;
* provide usable navigation;
* handle empty pages safely.

---

# Table and Row Actions

Row actions should preserve clear priority.

Destructive actions should use the shared confirmation pattern where required.

Avoid duplicating low-frequency detail actions into every row without product justification.

---

# Responsive Collections

Collections must have an intentional constrained-viewport strategy.

Possible approaches include:

* contained horizontal scrolling;
* priority-based column reduction;
* stacked rows;
* compact list/card presentation.

The whole page should not become horizontally scrollable because of one collection.

---

# Long Content in Collections

Potentially long content should use an intentional strategy such as:

* constrained width;
* wrapping;
* recoverable truncation;
* movement to the detail experience.

Product Experience determines whether the information belongs in the collection at all.

---

# Empty Collection Results

Distinguish between:

* no data exists;
* no data matches current discovery controls.

The latter should provide an understandable recovery path.

---

# Editors and Review Workflows

Where revisioned or reviewable content exists, interaction should clearly distinguish:

* read state;
* edit state;
* unsaved state;
* current state;
* historical state;
* approved or stale state where applicable.

Exact revision semantics remain owned by Specifications.

---

# Progressive Disclosure

Use progressive disclosure for secondary information where it materially improves interaction.

Do not hide primary information merely to reduce page length.

---

# Sticky Actions

Persistent action regions may be used where important actions would otherwise become difficult to reach.

They must remain usable across supported viewport sizes and must not obscure important content.

---

# Keyboard Interaction

Shared controls should support expected keyboard behavior.

Complex controls should use accessible primitives supported by the project's stack where possible.

---

# Focus Management

Modal and overlay interactions should manage focus predictably.

When closed, focus should return to a meaningful originating context where practical.

---

# Disabled vs Hidden Actions

Hide actions users are not authorized to know about.

Use disabled presentation when a legitimate capability is temporarily unavailable and understanding the prerequisite is useful.

Authorization remains server-authoritative.

---

# Optimistic Updates

Use optimistic interaction only when:

* failure is uncommon;
* rollback is simple;
* the operation is safe;
* stale state is unlikely to mislead the user.

Sensitive state transitions should not use optimistic behavior without explicit design justification.

---

# Theme Interaction

Interaction meaning must not depend on the active visual theme.

All supported themes must preserve understandable hover, active, focus, selected, disabled, destructive, and pending states.

---

# Reduced Motion

Interaction meaning must not depend on animation.

Respect reduced-motion preferences where applicable.

---

# Interaction Completeness

A user-facing feature is not interaction-complete merely because its primary request can execute.

Where applicable, completion includes:

* discoverable entry;
* pending state;
* duplicate prevention;
* success;
* validation;
* failure;
* empty state;
* unavailable state;
* confirmation;
* destructive-action safety;
* navigation continuity;
* recovery;
* responsive interaction;
* keyboard behavior;
* focus behavior;
* collection navigation;
* structured editing behavior.

---

# Shared Pattern Reuse

Before implementing a common interaction, determine whether an approved shared pattern already exists.

Repeated defects originating from a shared pattern should be corrected at the shared level when appropriate.

---

# Anti-Patterns

Avoid:

* browser-native alert and confirm for ordinary workflows;
* silent actions;
* Toast-only validation;
* Toast-only durable failures;
* unnecessary manual refresh;
* unsafe duplicate submission;
* destructive actions without required confirmation;
* inconsistent confirmation systems;
* unsupported filtering or sorting;
* client-only search over incomplete server-side data;
* collection state lost during navigation;
* uncontrolled long table content;
* giant unstructured forms;
* multiple unrelated editor implementations;
* unexplained disabled controls;
* blank loading or failure surfaces;
* page-specific implementations of already shared interaction patterns;
* responsive interactions whose required actions become unreachable.

---

# Relationship to UI Foundation

UI_FOUNDATION.md defines how shared interactions look and remain visually contained.

This document defines how they behave.

---

# Relationship to Product Experience

PRODUCT_EXPERIENCE.md determines where interactions occur and how they participate in product journeys.

This document determines their shared behavior.

---

# Relationship to Specifications

Specifications remain authoritative for whether an action exists and what business consequence it has.

The existence of a shared interaction pattern does not authorize the corresponding business capability.

---

# Change Policy

Change this document when:

* repeated interactions require shared behavior;
* feedback policy changes;
* confirmation policy changes;
* loading behavior changes;
* collection interaction standards change;
* editor behavior changes;
* responsive interaction rules change;
* asynchronous interaction conventions change;
* implementation experience reveals a reusable interaction rule.

Feature-specific business flows remain in their owning Specifications.