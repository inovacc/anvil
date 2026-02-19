package output

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	var buf bytes.Buffer

	f := New(&buf, FormatJSON)
	if f == nil {
		t.Fatal("New() returned nil")
	}

	if f.Format() != FormatJSON {
		t.Errorf("Format() = %v, want %v", f.Format(), FormatJSON)
	}
}

func TestNewText(t *testing.T) {
	var buf bytes.Buffer

	f := NewText(&buf)
	if f.IsJSON() {
		t.Error("NewText should not be JSON format")
	}

	if f.Format() != FormatText {
		t.Errorf("Format() = %v, want %v", f.Format(), FormatText)
	}
}

func TestNewJSON(t *testing.T) {
	var buf bytes.Buffer

	f := NewJSON(&buf)
	if !f.IsJSON() {
		t.Error("NewJSON should be JSON format")
	}
}

func TestFormatter_PrintJSON(t *testing.T) {
	var buf bytes.Buffer

	f := NewJSON(&buf)

	data := map[string]string{"key": "value"}

	err := f.Print(data)
	if err != nil {
		t.Fatalf("Print() error = %v", err)
	}

	if !strings.Contains(buf.String(), `"key"`) {
		t.Error("JSON output should contain key")
	}

	if !strings.Contains(buf.String(), `"value"`) {
		t.Error("JSON output should contain value")
	}
}

func TestFormatter_PrintText(t *testing.T) {
	var buf bytes.Buffer

	f := NewText(&buf)

	err := f.Print("hello world")
	if err != nil {
		t.Fatalf("Print() error = %v", err)
	}

	if strings.TrimSpace(buf.String()) != "hello world" {
		t.Errorf("Print() = %q, want %q", buf.String(), "hello world\n")
	}
}

func TestFormatter_PrintLines(t *testing.T) {
	tests := []struct {
		name     string
		lines    []string
		numbered bool
		format   Format
		wantSub  string
	}{
		{
			name:     "text unnumbered",
			lines:    []string{"line1", "line2"},
			numbered: false,
			format:   FormatText,
			wantSub:  "line1\nline2\n",
		},
		{
			name:     "text numbered",
			lines:    []string{"line1", "line2"},
			numbered: true,
			format:   FormatText,
			wantSub:  "1\tline1",
		},
		{
			name:     "json lines",
			lines:    []string{"line1", "line2"},
			numbered: false,
			format:   FormatJSON,
			wantSub:  `"line1"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			f := New(&buf, tt.format)

			err := f.PrintLines(tt.lines, tt.numbered)
			if err != nil {
				t.Fatalf("PrintLines() error = %v", err)
			}

			if !strings.Contains(buf.String(), tt.wantSub) {
				t.Errorf("PrintLines() = %q, want to contain %q", buf.String(), tt.wantSub)
			}
		})
	}
}

func TestFormatter_Println(t *testing.T) {
	var buf bytes.Buffer

	f := NewText(&buf)

	err := f.Println("test", "output")
	if err != nil {
		t.Fatalf("Println() error = %v", err)
	}

	if !strings.Contains(buf.String(), "test output") {
		t.Errorf("Println() = %q, want to contain %q", buf.String(), "test output")
	}
}

func TestFormatter_Printf(t *testing.T) {
	var buf bytes.Buffer

	f := NewText(&buf)

	err := f.Printf("value: %d\n", 42)
	if err != nil {
		t.Fatalf("Printf() error = %v", err)
	}

	if buf.String() != "value: 42\n" {
		t.Errorf("Printf() = %q, want %q", buf.String(), "value: 42\n")
	}
}

func TestFormatter_PrintTable(t *testing.T) {
	var buf bytes.Buffer

	f := NewTable(&buf)

	data := [][]string{
		{"Name", "Value"},
		{"foo", "bar"},
	}

	err := f.Print(data)
	if err != nil {
		t.Fatalf("Print() error = %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "Name") {
		t.Error("Table output should contain headers")
	}

	if !strings.Contains(output, "foo") {
		t.Error("Table output should contain data")
	}
}

func TestFormatter_PrintText_Slice(t *testing.T) {
	var buf bytes.Buffer

	f := NewText(&buf)

	err := f.Print([]string{"a", "b", "c"})
	if err != nil {
		t.Fatalf("Print() error = %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "a") || !strings.Contains(output, "b") {
		t.Error("Should print all slice elements")
	}
}

func TestResult_NewResult(t *testing.T) {
	r := NewResult("data")

	if !r.Success {
		t.Error("NewResult should set Success to true")
	}

	if r.Data != "data" {
		t.Errorf("Data = %v, want %v", r.Data, "data")
	}
}

func TestResult_NewError(t *testing.T) {
	err := errors.New("test error")
	r := NewError(err)

	if r.Success {
		t.Error("NewError should set Success to false")
	}

	if r.Error != "test error" {
		t.Errorf("Error = %v, want %v", r.Error, "test error")
	}
}

func TestResult_NewMessage(t *testing.T) {
	r := NewMessage("hello")

	if !r.Success {
		t.Error("NewMessage should set Success to true")
	}

	if r.Message != "hello" {
		t.Errorf("Message = %v, want %v", r.Message, "hello")
	}
}

func TestResult_Print(t *testing.T) {
	tests := []struct {
		name    string
		result  *Result
		format  Format
		wantSub string
	}{
		{
			name:    "json result",
			result:  NewResult("test"),
			format:  FormatJSON,
			wantSub: `"success": true`,
		},
		{
			name:    "text message",
			result:  NewMessage("hello"),
			format:  FormatText,
			wantSub: "hello",
		},
		{
			name:    "text error",
			result:  NewError(errors.New("oops")),
			format:  FormatText,
			wantSub: "error: oops",
		},
		{
			name:    "text data",
			result:  NewResult("data"),
			format:  FormatText,
			wantSub: "data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			f := New(&buf, tt.format)

			err := tt.result.Print(f)
			if err != nil {
				t.Fatalf("Print() error = %v", err)
			}

			if !strings.Contains(buf.String(), tt.wantSub) {
				t.Errorf("Print() = %q, want to contain %q", buf.String(), tt.wantSub)
			}
		})
	}
}

func TestFormatter_WriteError(t *testing.T) {
	tests := []struct {
		name    string
		format  Format
		wantSub string
	}{
		{"text error", FormatText, "error: boom"},
		{"json error", FormatJSON, `"error"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			f := New(&buf, tt.format)

			err := f.WriteError(errors.New("boom"))
			if err != nil {
				t.Fatalf("WriteError() error = %v", err)
			}

			if !strings.Contains(buf.String(), tt.wantSub) {
				t.Errorf("WriteError() = %q, want to contain %q", buf.String(), tt.wantSub)
			}
		})
	}
}

func TestFormatter_FormatName(t *testing.T) {
	tests := []struct {
		format Format
		want   string
	}{
		{FormatText, "text"},
		{FormatJSON, "json"},
		{FormatTable, "table"},
	}

	for _, tt := range tests {
		var buf bytes.Buffer

		f := New(&buf, tt.format)
		if got := f.FormatName(); got != tt.want {
			t.Errorf("FormatName() = %q, want %q", got, tt.want)
		}
	}
}

func TestOptions_GetFormat(t *testing.T) {
	tests := []struct {
		name string
		opts Options
		want Format
	}{
		{
			name: "default is text",
			opts: Options{},
			want: FormatText,
		},
		{
			name: "json flag",
			opts: Options{JSON: true},
			want: FormatJSON,
		},
		{
			name: "table flag",
			opts: Options{Table: true},
			want: FormatTable,
		},
		{
			name: "json takes precedence over table",
			opts: Options{JSON: true, Table: true},
			want: FormatJSON,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.opts.GetFormat(); got != tt.want {
				t.Errorf("GetFormat() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOptions_NewFormatter(t *testing.T) {
	var buf bytes.Buffer

	opts := Options{JSON: true}
	f := opts.NewFormatter(&buf)

	if !f.IsJSON() {
		t.Error("NewFormatter should create JSON formatter when JSON is true")
	}
}
