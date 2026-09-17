package orchestrate

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"gnomon/internal/contract"
	"gnomon/internal/present"
)

// renderPayloadSections builds present.Section bodies from a validated payload, generically —
// the entry point into renderObjectSections below, seeded with the schema root ("") and the full
// terminal_path split into segments.
func renderPayloadSections(rc contract.ResultContract, payload map[string]interface{}) []present.Section {
	return renderObjectSections(rc, "", strings.Split(rc.TerminalPath, "."), payload)
}

// renderObjectSections renders one JSON object's own fields as Sections, in the order the Result
// Contract's own schema declared them (contract.ResultContract.PropertyOrder, keyed by this
// object's own dotted path) — never a hardcoded per-workflow field name. Path-aware terminal
// suppression happens here, generically, at whatever nesting depth the schema actually declares:
// a field matching the next unconsumed segment of the terminal path is either the terminal leaf
// itself (the last segment — suppressed entirely, already reflected in the Report's Outcome/
// Summary) or a container the terminal value lives inside (an earlier segment — rendered by
// recursing into it, so every sibling field at every level still renders, and only the one true
// leaf is ever hidden). Every other field renders normally, absent/null/empty values omitted.
//
// path is this object's own PropertyOrder key ("" at the schema root, "summary" one level down,
// and so on). remainingTerminal is the terminal path's segments still to be matched from this
// object downward — once it no longer matches any field here, every field at and below this
// point renders unconditionally, exactly as a flat, unrelated object always has.
func renderObjectSections(rc contract.ResultContract, path string, remainingTerminal []string, obj map[string]interface{}) []present.Section {
	order := rc.PropertyOrder[path]
	if len(order) == 0 {
		// Defensive fallback only — Validate() already guarantees a schema with a declared
		// terminal property, so PropertyOrder is expected to be non-empty for any Contract that
		// reached this point. Falls back to this object's own keys, sorted, rather than silently
		// rendering nothing.
		for k := range obj {
			order = append(order, k)
		}
		sort.Strings(order)
	}

	var sections []present.Section
	for _, field := range order {
		if len(remainingTerminal) > 0 && field == remainingTerminal[0] {
			if len(remainingTerminal) == 1 {
				continue // the terminal leaf itself — never repeated here
			}
			child, ok := obj[field].(map[string]interface{})
			if !ok {
				continue // absent/null container: nothing beneath it to render
			}
			childPath := field
			if path != "" {
				childPath = path + "." + field
			}
			sections = append(sections, renderObjectSections(rc, childPath, remainingTerminal[1:], child)...)
			continue
		}
		body := renderFieldValue(obj[field])
		if body == "" {
			continue
		}
		sections = append(sections, present.Section{Label: humanizeFieldName(field), Body: body})
	}
	return sections
}

// renderFieldValue renders one payload field's value as plain text, generically over every shape
// the restricted Result Contract schema vocabulary (type/enum/required/properties/items) can
// produce: absent/null renders as "" (omitted by the caller); a string renders trimmed; a boolean
// renders as Yes/No; a number renders in its shortest decimal form; an array renders as one line
// per item (each object item rendered as its own small labeled block); an object renders as a
// small labeled block. No case here ever inspects a specific field name.
func renderFieldValue(v interface{}) string {
	switch val := v.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(val)
	case bool:
		if val {
			return "Yes"
		}
		return "No"
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	case []interface{}:
		return renderList(val)
	case map[string]interface{}:
		return renderObject(val)
	default:
		return fmt.Sprintf("%v", val)
	}
}

func renderList(items []interface{}) string {
	var lines []string
	for _, item := range items {
		if obj, ok := item.(map[string]interface{}); ok {
			if block := renderObject(obj); block != "" {
				lines = append(lines, block)
			}
			continue
		}
		if s := renderFieldValue(item); s != "" {
			lines = append(lines, "- "+s)
		}
	}
	return strings.Join(lines, "\n")
}

// renderObject renders one object (e.g. one finding, one obligation) as a single labeled line.
// Its own fields are sorted alphabetically — unlike the top-level payload, an array item's shape
// comes from the schema's `items`, not `properties`, so there is no declared top-level order to
// draw on here.
func renderObject(obj map[string]interface{}) string {
	keys := make([]string, 0, len(obj))
	for k := range obj {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var parts []string
	for _, k := range keys {
		val := renderFieldValue(obj[k])
		if val == "" {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s: %s", humanizeFieldName(k), val))
	}
	if len(parts) == 0 {
		return ""
	}
	return "- " + strings.Join(parts, "; ")
}

// humanizeFieldName turns a snake_case schema field name into a Title Case label — the one place
// presentation ever derives text from a field name, and it does so mechanically (splitting on
// "_" and capitalizing each word), never by matching a specific known name.
func humanizeFieldName(field string) string {
	parts := strings.Split(field, "_")
	for i, p := range parts {
		if p == "" {
			continue
		}
		parts[i] = strings.ToUpper(p[:1]) + strings.ToLower(p[1:])
	}
	return strings.Join(parts, " ")
}

// humanizeTerminalValue turns a workflow's own SCREAMING_SNAKE_CASE terminal value into a
// natural sentence fragment — e.g. "IMPLEMENTATION_COMPLETE" -> "Implementation complete",
// "BLOCKED" -> "Blocked", "KNOWLEDGE GAP" -> "Knowledge gap" — the one generic transform every
// workflow's terminal vocabulary already fits, since each declares its values as underscore- (or,
// in one case, space-) separated words for exactly this reason. This is the Report's Summary; it
// never repeats the workflow's identity or name, matching how every other command's Report has
// always read.
func humanizeTerminalValue(v string) string {
	words := strings.Split(v, "_")
	for i, w := range words {
		lw := strings.ToLower(w)
		if i == 0 && lw != "" {
			lw = strings.ToUpper(lw[:1]) + lw[1:]
		}
		words[i] = lw
	}
	return strings.Join(words, " ")
}
