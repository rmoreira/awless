package template

import (
	"encoding/json"
	"fmt"
	"io"
)

type Printer interface {
	Print(*TemplateExecution) error
}

func NewJSONPrinter(w io.Writer) *jsonPrinter {
	enc := json.NewEncoder(w)
	enc.SetIndent("", " ")
	return &jsonPrinter{
		enc: enc,
	}
}

type jsonPrinter struct {
	enc *json.Encoder
}

func (p *jsonPrinter) Print(t *TemplateExecution) error {
	if err := p.enc.Encode(t); err != nil {
		return fmt.Errorf("json printer: %s", err)
	}
	return nil
}
