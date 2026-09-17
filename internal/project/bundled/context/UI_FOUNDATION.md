# UI Foundation

## Purpose

This document defines the shared visual and presentation foundation of the product.

It is the authoritative project knowledge for:

* visual direction;
* design tokens;
* color system;
* typography;
* spacing;
* sizing;
* borders and radius;
* elevation;
* layout primitives;
* responsive visual behavior;
* responsive containment;
* overlay and dialog presentation;
* iconography;
* visual hierarchy;
* shared component presentation;
* theme presentation;
* localization-related presentation;
* user-facing data formatting;
* visual accessibility;
* visual consistency.

Its purpose is to ensure that independently implemented user-facing features visually form one coherent product.

A user-facing implementation is not visually complete merely because all required controls and information are present.

---

# Ownership Boundaries

This document owns:

* how the product looks;
* shared visual rules;
* visual hierarchy;
* design tokens;
* component appearance;
* layout primitives;
* responsive visual adaptation;
* viewport containment;
* theme presentation;
* visual accessibility;
* user-facing formatting conventions.

PRODUCT_EXPERIENCE.md owns:

* information architecture;
* product navigation;
* workspace organization;
* feature discoverability;
* cross-feature journeys;
* product-level page hierarchy.

INTERACTION_PATTERNS.md owns:

* interaction behavior;
* feedback behavior;
* confirmations;
* loading behavior;
* collection interaction;
* form interaction;
* editor behavior;
* asynchronous interaction.

Specifications own:

* business behavior;
* actors and authorization;
* business states;
* feature-specific workflows;
* acceptance criteria.

This document must not introduce or alter business behavior.

---

# Visual Direction

Define the intended visual identity of the product.

The project should explicitly establish:

* desired visual character;
* desired level of density;
* desired level of formality;
* relationship between content and controls;
* visual qualities the product should communicate;
* visual styles the product should explicitly avoid.

Project-specific visual direction belongs here.

---

# Visual Principles

Define the reusable visual principles governing the product.

At minimum, consider:

## Hierarchy Before Decoration

Visual decisions should communicate context, importance, relationship, state, and available action before decoration.

## Consistency Before Novelty

Equivalent concepts should use equivalent visual treatment.

## Intentional Density

Information density should reflect the product's actual tasks.

## Content Priority

The visual hierarchy should prioritize the information or content that provides primary product value.

Additional project-specific visual principles may be defined here.

---

# Design Token System

Define the shared token system for:

* colors;
* typography;
* spacing;
* sizing;
* radius;
* borders;
* shadows;
* container widths;
* breakpoints;
* transitions;
* focus treatment.

Feature-level implementations should consume shared tokens instead of introducing arbitrary values when an appropriate token exists.

---

# Color System

Define:

* primary palette;
* neutral palette;
* semantic colors;
* surface hierarchy;
* text hierarchy;
* interactive states.

Semantic roles should include, where applicable:

* success;
* warning;
* error;
* information.

Color must not be the only means of communicating important state.

Project-specific palette values belong here.

---

# Theme System

Define the themes supported by the product.

Specify:

* supported themes;
* default theme;
* theme initialization;
* user preference behavior;
* system preference behavior;
* persistence behavior;
* semantic token mapping across themes.

Theme detection and application should be centralized.

Feature pages must not independently determine or implement the active theme.

If multiple themes are supported, shared components must remain coherent in every supported theme.

---

# Typography

Define:

* primary UI typeface;
* optional content typefaces;
* heading hierarchy;
* body hierarchy;
* supporting text;
* metadata and caption treatment;
* line-height conventions.

Typography must support the languages and writing systems required by the product.

---

# Spacing System

Define a shared spacing scale.

Specify rules for:

* related elements;
* component internals;
* action groups;
* sections;
* major page regions;
* page-edge spacing.

Spacing should communicate relationship.

Arbitrary feature-specific margins should not become the primary layout mechanism.

---

# Layout System

Define shared layout primitives such as:

* application shell;
* page container;
* page header;
* content section;
* grid;
* reading surface;
* workspace surface;
* form layout;
* split layout.

Define:

* standard content widths;
* reading widths;
* form widths;
* page padding;
* section spacing;
* alignment rules.

Feature pages should not independently invent major layout rules.

---

# Responsive Foundation

Responsive design is a first-class usability requirement.

Define the supported viewport ranges and centralized breakpoint strategy.

At minimum, consider:

* compact/mobile;
* medium/tablet;
* desktop;
* wide desktop.

Responsive correctness must be evaluated by usability and containment rather than merely by the existence of responsive classes.

---

## Responsive Containment

User-facing elements must remain within the usable viewport.

At constrained sizes:

* primary workflows must remain usable;
* primary actions must remain reachable;
* essential navigation must remain accessible;
* unintended page-level horizontal overflow must not occur;
* forms and controls must adapt;
* long content must wrap or reflow safely;
* overlays and dialogs must remain usable.

---

## Responsive Reflow

Define how common structures adapt when available space decreases.

Consider:

* multi-column layouts;
* side content;
* action groups;
* workspace navigation;
* tables;
* page padding.

Responsive adaptation should preserve hierarchy rather than simply shrink the desktop layout.

---

# Overlay and Dialog Foundation

Define the visual and responsive contract for:

* dialogs;
* menus;
* popovers;
* drawers;
* overlays.

Shared overlay primitives should own common viewport containment.

Dialogs should:

* remain within viewport width and height;
* preserve edge spacing;
* support internal scrolling when required;
* preserve access to actions;
* adapt action layout on constrained viewports;
* support the project's themes and writing direction;
* preserve visible focus treatment.

Interaction semantics remain owned by INTERACTION_PATTERNS.md.

---

# Component Foundation

Define the shared UI primitives required by the project.

Potential primitives include:

* Button;
* IconButton;
* Input;
* Textarea;
* Select;
* Card;
* Badge;
* StatusBadge;
* Alert;
* Tabs;
* Dialog;
* EmptyState;
* Spinner;
* Skeleton;
* PageContainer;
* PageHeader.

Only introduce primitives justified by repeated product needs.

---

# Action Hierarchy

Define shared visual treatment for:

* primary actions;
* secondary actions;
* tertiary actions;
* destructive actions.

Several unrelated actions should not receive equal visual emphasis when their importance differs.

---

# Iconography

Define:

* approved icon family;
* sizing conventions;
* directional behavior;
* icon-only control requirements.

Avoid mixing unrelated icon systems without an approved reason.

---

# Status Presentation

Define the shared visual language used to present business states.

The existence and meaning of business states remain owned by Specifications.

This document governs only their presentation.

---

# Loading, Empty, and Error Presentation

Define shared visual presentation for:

* loading;
* skeletons;
* empty states;
* unavailable states;
* errors;
* partial content.

Behavioral rules belong to INTERACTION_PATTERNS.md.

---

# Form Presentation

Define shared visual treatment for:

* labels;
* required indicators;
* controls;
* descriptions;
* validation messages;
* disabled states;
* focus states;
* grouping;
* form actions.

---

# Table and Collection Presentation

Define visual rules for:

* headers;
* cells;
* spacing;
* row separation;
* action areas;
* long content;
* responsive presentation;
* overflow containment.

Collection information hierarchy belongs to PRODUCT_EXPERIENCE.md.

Interaction behavior belongs to INTERACTION_PATTERNS.md.

---

# Editor Presentation

Define how structured editing surfaces visually integrate with the product.

Editors should use the same visual language as the rest of the interface and remain usable across supported viewports, themes, languages, and writing directions.

Editor behavior belongs to INTERACTION_PATTERNS.md.

---

# Localization and Data Presentation

Define project-specific user-facing presentation conventions.

Where applicable, define:

* primary language;
* writing direction;
* digit presentation;
* calendar system;
* date formats;
* time formats;
* duration formats;
* percentages;
* file sizes;
* domain-specific value formatting.

Repeated transformations should use shared formatting utilities.

Do not assume a particular language, calendar, writing direction, or numeral system unless the project explicitly requires it.

---

# Accessibility Foundation

Define visual accessibility requirements.

At minimum:

* sufficient contrast;
* visible focus;
* understandable disabled states;
* identifiable interactive elements;
* non-color-only state communication;
* readable typography;
* reachable controls across supported viewports.

---

# Motion

Define shared motion principles.

Motion should remain functional and respect reduced-motion preferences where applicable.

---

# Visual Completeness

Define what must be true before a user-facing implementation is considered visually complete.

Where applicable, evaluate:

* hierarchy;
* spacing;
* composition;
* typography;
* semantic colors;
* responsive behavior;
* viewport containment;
* overlays;
* themes;
* localization;
* iconography;
* loading;
* empty states;
* error states;
* accessibility.

Rendered-interface verification should be used where source inspection alone cannot establish compliance.

---

# Component Reuse

Existing shared primitives should be preferred over feature-local equivalents.

Repeated visual defects originating from a shared primitive should be corrected at the shared level when appropriate.

---

# Extension Rule

Introduce a new shared visual pattern only when:

1. the existing foundation cannot correctly represent the requirement;
2. the requirement is legitimate according to approved project knowledge;
3. the new pattern does not unnecessarily duplicate an existing one.

---

# Anti-Patterns

Record project-wide visual anti-patterns here.

At minimum avoid:

* page-local design systems;
* arbitrary feature-specific styling;
* desktop-only layouts;
* overlays escaping the viewport;
* hidden overflow used to conceal responsive defects;
* inconsistent shared-component presentation;
* duplicated cross-cutting formatting logic.

---

# Relationship to Implementation

Implementation should translate this foundation into shared design tokens, layout primitives, components, theme infrastructure, formatting utilities, and responsive rules.

The exact code structure remains an implementation concern unless constrained by approved architecture or stack knowledge.

---

# Relationship to Specifications

Specifications determine what the interface must enable.

This document determines how those capabilities integrate with the shared visual system.

---

# Change Policy

Change this document when a project-wide visual rule, design-system decision, presentation convention, or reusable visual pattern changes.

Feature-specific business behavior must remain in its owning Specification.