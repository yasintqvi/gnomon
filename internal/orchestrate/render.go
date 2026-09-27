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

// renderObjectSections renders one JSON object's fields as Sections, in the schema's own declared
// order (rc.PropertyOrder, keyed by dotted path) — never a hardcoded field name. A field matching
// the next unconsumed segment of remainingTerminal is either the terminal leaf itself (suppressed —
// already reflected in the Report's Outcome/Summary) or a container the terminal value lives
// inside (recursed into, so its own sibling fields still render). Every other field renders
// normally, absent/null/empty values omitted.
func renderObjectSections(rc contract.ResultContract, path string, remainingTerminal []string, obj map[string]interface{}) []present.Section {
	order := rc.PropertyOrder[path]
	if len(order) == 0 {
		// Defensive fallback only — Validate() already guarantees PropertyOrder is non-empty here.
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
// the Result Contract schema vocabulary can produce — never by inspecting a specific field name.
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

// renderObject renders one object (e.g. one finding) as a small indented block: a leading "- "
// bullet on its first field, every other field on its own continuation line. Fields are sorted
// alphabetically — an array item's shape comes from the schema's `items`, with no declared order.
func renderObject(obj map[string]interface{}) string {
	keys := make([]string, 0, len(obj))
	for k := range obj {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var lines []string
	for _, k := range keys {
		val := renderFieldValue(obj[k])
		if val == "" {
			continue
		}
		lines = append(lines, fmt.Sprintf("%s: %s", humanizeFieldName(k), val))
	}
	if len(lines) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString("- ")
	b.WriteString(indentContinuation(lines[0]))
	for _, l := range lines[1:] {
		b.WriteString("\n    ")
		b.WriteString(indentContinuation(l))
	}
	return b.String()
}

// indentContinuation aligns an embedded newline in a field's own value under renderObject's
// 4-space continuation indent, so it never dedents back to column zero.
func indentContinuation(s string) string {
	return strings.ReplaceAll(s, "\n", "\n    ")
}

// humanizeFieldName turns a snake_case schema field name into a Title Case label — mechanically,
// never by matching a specific known name.
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

// humanizeTerminalValue turns a workflow's SCREAMING_SNAKE_CASE terminal value into a sentence
// fragment, e.g. "IMPLEMENTATION_COMPLETE" -> "Implementation complete". This becomes the Report's
// Summary; it never repeats the workflow's identity or name.
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
