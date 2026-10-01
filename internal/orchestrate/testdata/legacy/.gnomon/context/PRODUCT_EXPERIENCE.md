# Product Experience

## Purpose

This document defines the intended product-level experience of the project.

It is the authoritative project knowledge for:

* product information architecture;
* primary navigation;
* page and workspace hierarchy;
* cross-feature user journeys;
* feature discoverability;
* actor-oriented product experience;
* navigation continuity;
* contextual relationships between product areas;
* page-level information hierarchy;
* product-level responsive priorities;
* collection-level information density;
* resource-detail experience.

Its purpose is to ensure that independently implemented features form one coherent and usable product.

A feature being technically functional or directly reachable by URL is not sufficient evidence that it is correctly integrated into the product experience.

---

# Ownership Boundaries

This document owns:

* where capabilities belong;
* how major product areas relate;
* how users discover capabilities;
* how users move between related capabilities;
* what contextual information remains visible;
* high-level user journeys;
* product-level page hierarchy;
* workspace organization;
* collection-versus-detail allocation;
* product-level responsive priorities.

Specifications own:

* business behavior;
* actors and authorization;
* domain rules;
* feature-specific workflows;
* business states;
* acceptance criteria.

UI_FOUNDATION.md owns visual and presentation concerns.

INTERACTION_PATTERNS.md owns shared interaction behavior.

This document must not invent business capabilities not authorized by approved project knowledge.

---

# Experience Principles

Define the product-level principles governing the intended experience.

Consider principles such as:

## Task Oriented

Organize the product around user goals rather than technical architecture.

## Context Preserving

Preserve meaningful resource and workflow context while users move between related capabilities.

## Discoverable

Every user-triggered capability must have an appropriate discoverable entry point.

## Progressive

Help users understand what happened, what is happening, what requires attention, and what they can do next.

## Actor Aware

Adapt navigation and available actions according to legitimate actor needs and permissions without unnecessarily fragmenting the product.

## State Aware

Communicate important lifecycle and asynchronous states in user-meaningful terms.

## Scannable Before Detailed

Overview and collection surfaces should prioritize identification, comparison, navigation, and decision-making before complete detail.

Add or refine principles according to the product domain.

---

# Primary Product Model

Define the conceptual hierarchy of the product.

Example structure:

Application

├── Product Area A

│   └── Primary Resource

│       └── Related Resource

├── Product Area B

└── Administration

This hierarchy represents product organization rather than technical module boundaries.

Every major product area must be justified by user goals and approved product capabilities.

---

# Primary Contexts

Define the major contexts users operate within.

For each context specify:

* purpose;
* primary resource or task;
* important surrounding context;
* capabilities that belong within it;
* capabilities that explicitly do not belong within it.

Avoid defining contexts based solely on backend modules.

---

# Application Shell

Define the stable authenticated or primary application shell where applicable.

Specify:

* product identity;
* primary navigation;
* current destination indication;
* actor-appropriate destinations;
* account/user access;
* primary workspace behavior.

Primary navigation should represent major product areas rather than every feature.

---

# Navigation Hierarchy

Define the conceptual navigation levels used by the product.

Where applicable, distinguish:

## Product Navigation

Major product areas.

## Resource Context

Parent/child resource hierarchy.

## Workspace Navigation

Capabilities belonging to the same resource or task context.

Navigation must preserve orientation and avoid exposing raw implementation structure.

---

# Breadcrumbs and Context

Define when hierarchical context should be represented.

Breadcrumbs or equivalent mechanisms should:

* represent meaningful product hierarchy;
* link to useful ancestors;
* avoid mirroring raw URL segments;
* preserve orientation on constrained viewports.

Breadcrumbs must not substitute for required workspace navigation.

---

# Actor Experiences

For each important actor defined by approved project knowledge, describe the high-level product experience.

For each actor define:

* primary goals;
* primary product areas;
* important journeys;
* information priorities;
* capabilities requiring attention;
* capabilities that should remain hidden or secondary.

Do not introduce new actors here.

---

# Core User Journeys

Document important cross-feature journeys.

For each journey describe:

* starting context;
* major steps;
* context that must remain visible;
* meaningful state transitions;
* important next actions;
* expected completion context.

Journeys describe product continuity.

Detailed business behavior remains owned by Specifications.

---

# Workspace Model

Where multiple capabilities belong to one resource or task context, define the shared workspace model.

Specify:

* workspace identity;
* persistent context;
* workspace navigation;
* overview/orientation surface;
* capability integration;
* state communication.

Related capabilities should not feel like unrelated applications.

---

# Feature Discoverability

Every user-triggered capability required by an approved Specification must be reachable from the context where its authorized actor would reasonably expect it.

A capability must not rely solely on:

* direct URLs;
* API knowledge;
* browser history;
* implementation-specific navigation.

For new capabilities determine:

1. owning product context;
2. discoverable entry point;
3. prerequisite context;
4. meaningful next step.

---

# Page Composition

Pages should establish:

1. current context;
2. page purpose;
3. important state;
4. primary action;
5. primary content;
6. secondary information.

Not every piece of information should receive equal visual or spatial importance.

---

# Collection Experiences

Define how growing collections fit the product experience.

Collection surfaces should primarily support:

* scanning;
* identification;
* comparison;
* selection;
* navigation.

They should not duplicate complete resource details.

Define which information belongs in:

* collection;
* detail;
* workspace;
* progressive disclosure.

Search, filtering, sorting, and pagination behavior belongs to INTERACTION_PATTERNS.md.

---

# Collection-to-Detail Relationship

Collection surfaces should answer:

* What resources are available?
* Which resource is relevant?
* What important state should I know before opening it?

Detail or workspace surfaces should answer:

* What is this resource?
* What complete information belongs to it?
* What actions can I perform?
* What related capabilities are available?

---

# Product State Communication

Define how users conceptually understand states required by approved Specifications.

Where applicable distinguish:

* not created;
* unavailable;
* pending;
* processing;
* requires attention;
* completed;
* partially completed;
* failed;
* outdated;
* archived;
* other domain-specific states.

This document governs experience-level communication, not business semantics.

---

# Empty Experiences

Define how meaningful empty states should guide users.

An empty experience should communicate:

* what is missing;
* whether absence is expected;
* prerequisites;
* available action;
* actor responsibility where useful.

Blank content areas are not intentional empty experiences.

---

# Failure Experience

Failures should preserve user context and expose meaningful recovery where available.

Technical infrastructure details should not become part of normal product experience.

---

# Responsive Experience Priorities

Define what must remain available when space becomes constrained.

A useful default priority is:

1. current context;
2. primary content;
3. primary action;
4. essential navigation;
5. important state;
6. secondary information.

Project-specific priorities may refine this ordering.

Responsive experience must preserve task completion rather than merely reproduce desktop information density.

---

# Product Consistency

Equivalent concepts should appear in equivalent product contexts.

A new feature should extend the existing product model before introducing a new navigation, workspace, or product-area concept.

---

# New Feature Integration

When implementing a new user-facing Specification, determine:

1. which product context owns the feature;
2. where authorized users discover it;
3. how users reach it;
4. what context remains visible;
5. what prerequisites exist;
6. what meaningful next step follows;
7. whether it extends an existing workspace;
8. what belongs in collection versus detail surfaces;
9. how it affects existing journeys and navigation.

A new Specification does not automatically justify a new page, workspace, or top-level product area.

---

# Experience Completeness

A user-facing feature is not fully integrated merely because its business behavior works.

Where applicable, integration requires:

* discoverability;
* correct product context;
* understandable hierarchy;
* navigation continuity;
* prerequisite communication;
* state communication;
* meaningful next actions;
* intentional empty and failure experiences;
* usable responsive hierarchy;
* appropriate collection/detail allocation;
* integration with surrounding journeys.

---

# Anti-Patterns

Avoid:

* isolated pages per Specification;
* backend modules exposed as product navigation;
* URL-only discoverability;
* unrelated layouts for related capabilities;
* raw metadata dumps as resource experiences;
* complete resource detail inside collection rows;
* excessive top-level navigation;
* dead-end completion states;
* technical processing details replacing user-meaningful state;
* mobile experiences treated as scaled desktop layouts.

---

# Relationship to Specifications

This document defines the experience framework into which Specifications are integrated.

It does not authorize capabilities not defined by approved project knowledge.

---

# Change Policy

Change this document when:

* information architecture changes;
* major product areas change;
* navigation philosophy changes;
* cross-feature journeys change;
* actor-oriented experience changes;
* workspace concepts change;
* collection/detail responsibilities change;
* repeated implementation experience reveals a product-level UX rule.

Feature-specific business behavior remains in Specifications.