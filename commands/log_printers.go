package commands

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/fatih/color"
	"github.com/wallix/awless/console"
	"github.com/wallix/awless/template"
)

type logPrinter struct {
	w io.Writer
}

func (p *logPrinter) Print(t *template.TemplateExecution) error {
	writeMetadata(t, p.w)

	for _, cmd := range t.CommandNodesIterator() {
		var status string
		if cmd.CmdErr != nil {
			status = renderRedFn("KO")
		} else {
			status = renderGreenFn("OK")
		}

		var line string
		if v, ok := cmd.CmdResult.(string); ok && v != "" {
			line = fmt.Sprintf("    %s\t%s\t[%s]", status, cmd.String(), v)
		} else {
			line = fmt.Sprintf("    %s\t%s", status, cmd.String(), v)
		}
		fmt.Fprintln(p.w, line)

		if cmd.CmdErr != nil {
			for _, err := range formatMultiLineErrMsg(cmd.CmdErr.Error()) {
				fmt.Fprintf(p.w, "%s\t%s\n", "", err)
			}
		}

	}
	return nil
}

type shortLogPrinter struct {
	w io.Writer
}

func (p *shortLogPrinter) Print(t *template.TemplateExecution) error {
	var ko, ok int
	var oneAction, oneEntity string
	for _, cmd := range t.CommandNodesIterator() {
		if cmd.CmdErr != nil {
			ko++
		} else {
			ok++
		}
		oneAction = cmd.Action
		oneEntity = cmd.Entity
	}

	fmt.Fprint(p.w, t.ID)
	if ko == 0 {
		color.New(color.FgGreen).Fprint(p.w, " OK")
	} else {
		color.New(color.FgRed).Fprint(p.w, " KO")
	}

	fmt.Fprint(p.w, " - ")

	cmdCount := ko + ok
	if cmdCount == 1 {
		fmt.Fprintf(p.w, "%s %s", oneAction, oneEntity)
	} else {
		fmt.Fprintf(p.w, "%d commands", cmdCount)
	}

	fmt.Fprintf(p.w, " (%s ago)", console.HumanizeTime(t.Date()))

	if t.Author != "" {
		if t.Profile != "" {
			fmt.Fprintf(p.w, " <%s:%s>", t.Profile, t.Author)
		} else {
			fmt.Fprintf(p.w, " <%s>", t.Author)
		}
	}
	if t.Locale != "" {
		fmt.Fprintf(p.w, " [%s]", t.Locale)
	}
	if !template.IsRevertible(t.Template) {
		fmt.Fprintf(p.w, " (not revertible)")
	}

	tabw := tabwriter.NewWriter(p.w, 0, 8, 0, '\t', 0)
	tabw.Flush()

	return nil
}

func NewDefaultTemplatePrinter(w io.Writer) template.Printer {
	return &defaultPrinter{w}
}

type defaultPrinter struct {
	w io.Writer
}

func (p *defaultPrinter) Print(t *template.TemplateExecution) error {
	tabw := tabwriter.NewWriter(p.w, 0, 8, 0, '\t', 0)
	for _, cmd := range t.CommandNodesIterator() {
		var status string
		if cmd.Err() != nil {
			status = renderRedFn("KO")
		} else {
			status = renderGreenFn("OK")
		}

		var line string
		if v, ok := cmd.Result().(string); ok && v != "" {
			line = fmt.Sprintf("    %s\t%s = %s\t", status, cmd.Entity, v)
		} else {
			line = fmt.Sprintf("    %s\t%s %s\t", status, cmd.Action, cmd.Entity)
		}

		fmt.Fprintln(tabw, line)
		if cmd.Err() != nil {
			for _, err := range formatMultiLineErrMsg(cmd.Err().Error()) {
				fmt.Fprintf(tabw, "%s\t%s\n", "", err)
			}
		}
	}

	tabw.Flush()
	return nil
}

type idOnlyPrinter struct {
	w io.Writer
}

func (p *idOnlyPrinter) Print(t *template.TemplateExecution) error {
	fmt.Fprint(p.w, t.ID)
	return nil
}

func writeMetadata(t *template.TemplateExecution, w io.Writer) {
	fmt.Fprintf(w, "ID: %s\tDate: %s", t.ID, t.Date().Format(time.Stamp))
	if t.Author != "" {
		fmt.Fprintf(w, "\tAuthor: %s", t.Author)
	}
	if t.Locale != "" {
		fmt.Fprintf(w, "\tRegion: %s", t.Locale)
	}
	if t.Profile != "" {
		fmt.Fprintf(w, "\tProfile: %s", t.Profile)
	}
	if !template.IsRevertible(t.Template) {
		fmt.Fprintf(w, " (not revertible)")
	}
	fmt.Fprintln(w)
}

func formatMultiLineErrMsg(msg string) []string {
	notabs := strings.Replace(msg, "\t", "", -1)
	var indented []string
	for _, line := range strings.Split(notabs, "\n") {
		indented = append(indented, fmt.Sprintf("    %s", line))
	}
	return indented
}
