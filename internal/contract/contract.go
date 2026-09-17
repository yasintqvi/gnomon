// Package contract loads and structurally validates a workflow's Workflow Contract and Result
// Contract from its own YAML frontmatter, per cli/WORKFLOW_CONTRACT.md. It never interprets or
// duplicates any Core semantic rule beyond what that frontmatter itself declares.
package contract

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v5"
	"gopkg.in/yaml.v3"
)

// SpecReference is a workflow's declared Specification-reference cardinality.
type SpecReference string

const (
	SpecReferenceNone     SpecReference = "none"
	SpecReferenceOptional SpecReference = "optional"
	SpecReferenceRequired SpecReference = "required"
)

// ResultContract is the terminal-result location, exhaustive terminal classification, and
// payload schema a workflow declares.
type ResultContract struct {
	TerminalPath string `yaml:"terminal_path"`
	// Classification maps every value the terminal_path property's schema enum permits to
	// "success" or "blocked" — CLI/runtime handling of a valid, already-classified terminal
	// result. It never reinterprets the workflow's own conclusion (a Verification "FAIL" is a
	// successfully completed run reporting a negative conclusion, not a process/Agent failure)
	// and it is exhaustive by construction: Validate() rejects any Contract where a permitted
	// terminal value has no entry here, or where an entry references a value the schema does not
	// permit, or where a value classifies as anything other than "success"/"blocked". "failed" is
	// deliberately never a classification value — that outcome is reserved for actual CLI/
	// process/protocol failure, computed entirely outside this map.
	Classification map[string]string      `yaml:"classification"`
	Schema         map[string]interface{} `yaml:"schema"`

	// PropertyOrder maps a dotted path prefix — "" for the schema's own top level, or a nested
	// object property's own dotted path (e.g. "summary") — to that level's declared
	// result.schema.properties key order. Information a plain map[string]interface{} decode
	// necessarily discards, and the one piece of Contract structure generic rendering needs to
	// present a payload's fields in the order their own author declared them, at whatever nesting
	// depth a Result Contract's own terminal_path happens to pass through — never a hardcoded
	// field name. Only object properties are recorded this way; array `items` have no single
	// declared order a rendered list of them could follow, so rendering falls back to alphabetical
	// there, unaffected by this map.
	PropertyOrder map[string][]string `yaml:"-"`
}

// Workflow is one workflow file's fully-loaded machine contract.
type Workflow struct {
	Identity                      string         `yaml:"identity"`
	SpecificationReference        SpecReference  `yaml:"specification_reference"`
	RequiresApprovedSpecification bool           `yaml:"requires_approved_specification"`
	Result                        ResultContract `yaml:"result"`
}

// Load reads and structurally validates a workflow Markdown file's frontmatter.
func Load(path string) (Workflow, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Workflow{}, err
	}
	fm, err := splitFrontmatter(data)
	if err != nil {
		return Workflow{}, fmt.Errorf("%s: %w", path, err)
	}
	var w Workflow
	if err := yaml.Unmarshal(fm, &w); err != nil {
		return Workflow{}, fmt.Errorf("%s: malformed frontmatter: %w", path, err)
	}

	order, err := schemaPropertyOrder(fm)
	if err != nil {
		return Workflow{}, fmt.Errorf("%s: %w", path, err)
	}
	w.Result.PropertyOrder = order

	if err := w.Validate(); err != nil {
		return Workflow{}, fmt.Errorf("%s: %w", path, err)
	}
	return w, nil
}

// schemaPropertyOrder walks the raw frontmatter a second time, via yaml.Node, purely to capture
// the declared order of result.schema.properties at every nesting level — order a plain-map
// yaml.Unmarshal necessarily discards. fm is already known-valid YAML at the point this is
// called (the first Unmarshal into Workflow above already succeeded against the same bytes), so
// failure here is not expected in practice, but is still reported rather than silently swallowed.
func schemaPropertyOrder(fm []byte) (map[string][]string, error) {
	var root yaml.Node
	if err := yaml.Unmarshal(fm, &root); err != nil {
		return nil, fmt.Errorf("malformed frontmatter: %w", err)
	}
	if len(root.Content) == 0 {
		return nil, nil
	}
	schemaNode := mappingValue(mappingValue(root.Content[0], "result"), "schema")
	order := map[string][]string{}
	collectPropertyOrder(schemaNode, "", order)
	return order, nil
}

// collectPropertyOrder recursively walks a JSON-Schema-shaped YAML node, recording each object
// node's own declared properties order under its dotted path prefix (joined with "." — the same
// convention TerminalPath and TerminalValue already use). It descends only into nested object
// properties (schema.properties.X.properties...), never into array `items` — an array-typed
// property node has no "properties" key of its own (it has "items" instead), so the mere absence
// of one naturally halts descent there without needing to special-case "type: array" explicitly.
// This is what lets generic rendering follow a terminal_path of arbitrary depth through arbitrary
// future/custom Result Contracts, not only the one level "summary.aggregate" happens to need.
func collectPropertyOrder(schemaNode *yaml.Node, prefix string, order map[string][]string) {
	properties := mappingValue(schemaNode, "properties")
	if properties == nil || properties.Kind != yaml.MappingNode {
		return
	}
	names := make([]string, 0, len(properties.Content)/2)
	for i := 0; i+1 < len(properties.Content); i += 2 {
		name := properties.Content[i].Value
		names = append(names, name)
		childPath := name
		if prefix != "" {
			childPath = prefix + "." + name
		}
		collectPropertyOrder(properties.Content[i+1], childPath, order)
	}
	order[prefix] = names
}

// mappingValue returns the value node for key within a YAML mapping node, or nil if node is not
// a mapping or key is absent.
func mappingValue(node *yaml.Node, key string) *yaml.Node {
	if node == nil || node.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			return node.Content[i+1]
		}
	}
	return nil
}

// splitFrontmatter extracts the YAML block delimited by leading/trailing "---" lines.
func splitFrontmatter(data []byte) ([]byte, error) {
	const delim = "---"
	s := string(data)
	if !strings.HasPrefix(s, delim) {
		return nil, fmt.Errorf("no frontmatter block found")
	}
	rest := s[len(delim):]
	idx := strings.Index(rest, "\n"+delim)
	if idx == -1 {
		return nil, fmt.Errorf("unterminated frontmatter block")
	}
	return []byte(strings.TrimPrefix(rest[:idx], "\n")), nil
}

// Validate performs the structural checks cli/WORKFLOW_CONTRACT.md's validation table defines.
func (w Workflow) Validate() error {
	if w.Identity == "" {
		return fmt.Errorf("missing identity")
	}
	switch w.SpecificationReference {
	case SpecReferenceNone, SpecReferenceOptional, SpecReferenceRequired:
	default:
		return fmt.Errorf("invalid specification_reference: %q", w.SpecificationReference)
	}
	if w.SpecificationReference == SpecReferenceNone && w.RequiresApprovedSpecification {
		return fmt.Errorf("requires_approved_specification is true while specification_reference is none")
	}
	if w.Result.TerminalPath == "" {
		return fmt.Errorf("missing result.terminal_path")
	}
	if w.Result.Schema == nil {
		return fmt.Errorf("missing result.schema")
	}
	if _, err := w.Result.compile(); err != nil {
		return fmt.Errorf("invalid result.schema: %w", err)
	}

	enumValues, err := w.Result.terminalEnum()
	if err != nil {
		return err
	}
	if err := w.Result.validateClassification(enumValues); err != nil {
		return err
	}
	return nil
}

// terminalEnum returns the declared enum values for the terminal_path property — every terminal
// value classification must exhaustively cover, read directly from the schema rather than
// duplicated anywhere else, so the two can never drift apart. terminal_path is a dotted path
// (schemaNodeAt walks it generically, at whatever nesting depth it declares), matching the same
// convention TerminalValue already uses to extract the runtime payload value.
func (rc ResultContract) terminalEnum() ([]string, error) {
	node, err := rc.schemaNodeAt(rc.TerminalPath)
	if err != nil {
		return nil, err
	}
	rawEnum, ok := node["enum"].([]interface{})
	if !ok || len(rawEnum) == 0 {
		return nil, fmt.Errorf("result.schema.properties.%s has no enum declared", rc.TerminalPath)
	}
	values := make([]string, 0, len(rawEnum))
	for _, v := range rawEnum {
		s, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("result.schema.properties.%s.enum contains a non-string value", rc.TerminalPath)
		}
		values = append(values, s)
	}
	return values, nil
}

// schemaNodeAt walks rc.Schema's properties tree along a dotted path, returning the JSON-Schema
// node describing that property — generic over any nesting depth a Result Contract's schema
// actually declares, never assuming a flat, single-segment terminal_path.
func (rc ResultContract) schemaNodeAt(dottedPath string) (map[string]interface{}, error) {
	segments := strings.Split(dottedPath, ".")
	current, ok := rc.Schema["properties"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("result.schema has no properties")
	}
	var node map[string]interface{}
	for i, seg := range segments {
		node, ok = current[seg].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("result.schema.properties has no %q entry matching result.terminal_path %q", seg, dottedPath)
		}
		if i == len(segments)-1 {
			return node, nil
		}
		current, ok = node["properties"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("result.schema.properties.%s has no nested properties to continue result.terminal_path %q", seg, dottedPath)
		}
	}
	return node, nil
}

// validateClassification enforces exhaustive, consistent terminal classification: every enum
// value the schema permits must have exactly one classification entry, every classification
// entry must reference an actual enum value, and every classification value must be one of the
// two the CLI's presentation layer understands ("failed" is never a classification value — that
// outcome is reserved for actual process/protocol failure). Nothing is ever guessed at runtime —
// a missing, unknown, inconsistent, or incomplete declaration is rejected here, at load time, not
// discovered later against a real Agent result.
func (rc ResultContract) validateClassification(enumValues []string) error {
	if len(rc.Classification) == 0 {
		return fmt.Errorf("missing result.classification (required for every terminal value in %v)", enumValues)
	}
	enumSet := make(map[string]bool, len(enumValues))
	for _, v := range enumValues {
		enumSet[v] = true
	}
	for value := range rc.Classification {
		if !enumSet[value] {
			return fmt.Errorf("result.classification references %q, which is not a declared terminal value", value)
		}
	}
	for _, value := range enumValues {
		class, ok := rc.Classification[value]
		if !ok {
			return fmt.Errorf("result.classification is missing an entry for terminal value %q", value)
		}
		switch class {
		case "success", "blocked":
		default:
			return fmt.Errorf("result.classification[%q] = %q is not a recognized classification (valid: success, blocked)", value, class)
		}
	}
	return nil
}

// ClassificationFor returns the declared classification ("success" or "blocked") for a terminal
// value already confirmed valid against the schema's enum. It performs no additional validation
// itself — Validate() has already guaranteed, at load time, that every schema-permitted value has
// exactly one of these two classifications.
func (rc ResultContract) ClassificationFor(terminalValue string) (string, bool) {
	c, ok := rc.Classification[terminalValue]
	return c, ok
}

func (rc ResultContract) compile() (*jsonschema.Schema, error) {
	b, err := json.Marshal(rc.Schema)
	if err != nil {
		return nil, err
	}
	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource("result-schema.json", bytes.NewReader(b)); err != nil {
		return nil, err
	}
	return compiler.Compile("result-schema.json")
}

// ValidatePayload validates a decoded payload value against this Result Contract's schema.
func (rc ResultContract) ValidatePayload(payload interface{}) error {
	schema, err := rc.compile()
	if err != nil {
		return err
	}
	return schema.Validate(payload)
}

// TerminalValue extracts the terminal string value from payload at TerminalPath (a dotted path).
func (rc ResultContract) TerminalValue(payload map[string]interface{}) (string, error) {
	parts := strings.Split(rc.TerminalPath, ".")
	var cur interface{} = payload
	for _, p := range parts {
		m, ok := cur.(map[string]interface{})
		if !ok {
			return "", fmt.Errorf("terminal_path %q not found in payload", rc.TerminalPath)
		}
		v, ok := m[p]
		if !ok {
			return "", fmt.Errorf("terminal_path %q not found in payload", rc.TerminalPath)
		}
		cur = v
	}
	s, ok := cur.(string)
	if !ok {
		return "", fmt.Errorf("terminal value at %q is not a string", rc.TerminalPath)
	}
	return s, nil
}
