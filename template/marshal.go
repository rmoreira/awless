package template

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/oklog/ulid"
	"github.com/wallix/awless/template/internal/ast"
)

// Allow template executions serialization with context for JSON storage
// without altering the template.Template model
type TemplateExecution struct {
	*Template
	Author, Source, Locale, Profile string
	Fillers                         map[string]interface{}
}

// The ID of a template is stored as an ULID therefore the ID
// capture the date and template IDs are sortable lexicagraphically
func (t *TemplateExecution) Date() time.Time {
	parsed, err := ulid.Parse(t.ID)
	if err != nil {
		panic(err)
	}
	return time.Unix(int64(parsed.Time())/int64(1000), time.Nanosecond.Nanoseconds())
}

func (t *TemplateExecution) MarshalJSON() ([]byte, error) {
	out := &toJSON{}
	out.ID = t.ID
	out.Author = t.Author
	out.Source = t.Source
	out.Locale = t.Locale
	out.Profile = t.Profile
	out.Fillers = t.Fillers
	if out.Fillers == nil {
		out.Fillers = make(map[string]interface{}, 0) // friendlier for json, avoiding "fillers": null,
	}
	out.Commands = []command{}

	for _, cmd := range t.CommandNodesIterator() {
		newCmd := command{}
		newCmd.Line = cmd.String()
		if cmd.CmdErr != nil {
			newCmd.Errors = append(newCmd.Errors, cmd.CmdErr.Error())
		}
		if cmd.CmdResult != nil {
			if s, ok := cmd.CmdResult.(string); ok {
				newCmd.Results = append(newCmd.Results, s)
			}
		}
		out.Commands = append(out.Commands, newCmd)
	}

	return json.MarshalIndent(out, "", " ")
}

func (t *TemplateExecution) UnmarshalJSON(b []byte) error {
	if t == nil {
		t = new(TemplateExecution)
	}

	if t.Template == nil {
		t.Template = new(Template)
	}

	var v toJSON

	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}

	t.Source = v.Source
	t.Locale = v.Locale
	t.Profile = v.Profile
	t.Author = v.Author
	t.Fillers = v.Fillers

	tpl := &Template{ID: v.ID, AST: &ast.AST{
		Statements: make([]*ast.Statement, 0),
	}}

	for _, c := range v.Commands {
		node, err := parseStatement(c.Line)
		if err != nil {
			return err
		}

		switch node.(type) {
		case *ast.CommandNode:
			n := node.(*ast.CommandNode)
			if len(c.Results) > 0 {
				n.CmdResult = c.Results[0]
			}
			if len(c.Errors) > 0 {
				n.CmdErr = errors.New(c.Errors[0])
			}
			tpl.Statements = append(tpl.Statements, &ast.Statement{Node: n})
		}
	}

	*(t.Template) = *tpl

	return nil
}

func (t *TemplateExecution) Stats() map[string]int {
	stats := make(map[string]int)
	for _, cmd := range t.CommandNodesIterator() {
		stats[fmt.Sprintf("%s %s", cmd.Action, cmd.Entity)]++
	}
	return stats
}

type toJSON struct {
	ID       string                 `json:"id"`
	Author   string                 `json:"author,omitempty"`
	Source   string                 `json:"source"`
	Locale   string                 `json:"locale"`
	Profile  string                 `json:"profile,omitempty"`
	Fillers  map[string]interface{} `json:"fillers"`
	Commands []command              `json:"commands"`
}

type command struct {
	Line    string   `json:"line"`
	Errors  []string `json:"errors,omitempty"`
	Results []string `json:"results,omitempty"`
}
