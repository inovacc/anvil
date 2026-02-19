// Package output provides unified output formatting for CLI commands.
// It supports text, JSON, and table formats with consistent behavior.
package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

// Format represents the output format type
type Format int

const (
	// FormatText is the default human-readable text format
	FormatText Format = iota
	// FormatJSON outputs data as JSON
	FormatJSON
	// FormatTable outputs data in aligned columns
	FormatTable
)

// Formatter handles output in various formats
type Formatter struct {
	w      io.Writer
	format Format
}

// New creates a new Formatter with the given writer and format
func New(w io.Writer, format Format) *Formatter {
	return &Formatter{w: w, format: format}
}

// NewText creates a text formatter
func NewText(w io.Writer) *Formatter {
	return New(w, FormatText)
}

// NewJSON creates a JSON formatter
func NewJSON(w io.Writer) *Formatter {
	return New(w, FormatJSON)
}

// NewTable creates a table formatter
func NewTable(w io.Writer) *Formatter {
	return New(w, FormatTable)
}

// Format returns the current format
func (f *Formatter) Format() Format {
	return f.format
}

// IsJSON returns true if format is JSON
func (f *Formatter) IsJSON() bool {
	return f.format == FormatJSON
}

// Print outputs data according to the format
func (f *Formatter) Print(data any) error {
	switch f.format {
	case FormatJSON:
		return f.printJSON(data)
	case FormatTable:
		return f.printTable(data)
	case FormatText:
		return f.printText(data)
	}

	return f.printText(data)
}

// PrintLines outputs lines with optional line numbers
func (f *Formatter) PrintLines(lines []string, numbered bool) error {
	if f.format == FormatJSON {
		return f.printJSON(lines)
	}

	for i, line := range lines {
		if numbered {
			_, err := fmt.Fprintf(f.w, "%6d\t%s\n", i+1, line)
			if err != nil {
				return err
			}
		} else {
			_, err := fmt.Fprintln(f.w, line)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Println outputs a single line
func (f *Formatter) Println(a ...any) error {
	if f.format == FormatJSON {
		return f.printJSON(a)
	}

	_, err := fmt.Fprintln(f.w, a...)

	return err
}

// Printf outputs formatted text (only for text format)
func (f *Formatter) Printf(format string, a ...any) error {
	_, err := fmt.Fprintf(f.w, format, a...)
	return err
}

func (f *Formatter) printJSON(data any) error {
	enc := json.NewEncoder(f.w)
	enc.SetIndent("", "  ")

	return enc.Encode(data)
}

func (f *Formatter) printText(data any) error {
	switch v := data.(type) {
	case string:
		_, err := fmt.Fprintln(f.w, v)
		return err
	case []string:
		for _, s := range v {
			if _, err := fmt.Fprintln(f.w, s); err != nil {
				return err
			}
		}

		return nil
	case fmt.Stringer:
		_, err := fmt.Fprintln(f.w, v.String())
		return err
	default:
		_, err := fmt.Fprintf(f.w, "%v\n", data)
		return err
	}
}

func (f *Formatter) printTable(data any) error {
	// For table format, use tabwriter for aligned columns
	tw := tabwriter.NewWriter(f.w, 0, 0, 2, ' ', 0)

	switch v := data.(type) {
	case [][]string:
		for _, row := range v {
			_, _ = fmt.Fprintln(tw, strings.Join(row, "\t"))
		}
	case []string:
		for _, s := range v {
			_, _ = fmt.Fprintln(tw, s)
		}
	default:
		_, _ = fmt.Fprintf(tw, "%v\n", data)
	}

	return tw.Flush()
}

// Result represents a command result that can be formatted
type Result struct {
	// Data is the main output data
	Data any `json:"data,omitempty"`
	// Error is any error message
	Error string `json:"error,omitempty"`
	// Message is a human-readable message
	Message string `json:"message,omitempty"`
	// Count is the number of items (for counting operations)
	Count int `json:"count,omitempty"`
	// Success indicates if the operation succeeded
	Success bool `json:"success"`
}

// NewResult creates a successful result with data
func NewResult(data any) *Result {
	return &Result{Data: data, Success: true}
}

// NewError creates an error result
func NewError(err error) *Result {
	return &Result{Error: err.Error(), Success: false}
}

// NewMessage creates a result with a message
func NewMessage(msg string) *Result {
	return &Result{Message: msg, Success: true}
}

// Print outputs the result in the given format
func (r *Result) Print(f *Formatter) error {
	if f.IsJSON() {
		return f.Print(r)
	}

	if r.Error != "" {
		return f.Printf("error: %s\n", r.Error)
	}

	if r.Message != "" {
		return f.Println(r.Message)
	}

	if r.Data != nil {
		return f.Print(r.Data)
	}

	return nil
}

// WriteError writes an error as a formatted result
func (f *Formatter) WriteError(err error) error {
	if f.format == FormatJSON {
		return f.printJSON(map[string]any{
			"error":   err.Error(),
			"success": false,
		})
	}

	_, wErr := fmt.Fprintf(f.w, "error: %v\n", err)

	return wErr
}

// FormatName returns the name of the current format (for debugging)
func (f *Formatter) FormatName() string {
	switch f.format {
	case FormatJSON:
		return "json"
	case FormatTable:
		return "table"
	default:
		return "text"
	}
}

// Options that can be embedded in command options
type Options struct {
	JSON  bool // --json flag
	Table bool // --table flag
}

// GetFormat returns the format based on options
func (o Options) GetFormat() Format {
	if o.JSON {
		return FormatJSON
	}

	if o.Table {
		return FormatTable
	}

	return FormatText
}

// NewFormatter creates a formatter based on options
func (o Options) NewFormatter(w io.Writer) *Formatter {
	return New(w, o.GetFormat())
}
