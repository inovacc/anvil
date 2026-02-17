package main

import (
	"testing"

	"github.com/inovacc/anvil/pkg/vault"
)

func TestParseEnvData(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []vault.SecretEntry
		wantErr bool
	}{
		{
			name:  "basic key=value",
			input: "FOO=bar\nBAZ=qux\n",
			want: []vault.SecretEntry{
				{Key: "FOO", Value: "bar"},
				{Key: "BAZ", Value: "qux"},
			},
		},
		{
			name:  "export prefix",
			input: "export FOO=bar\nexport BAZ=qux\n",
			want: []vault.SecretEntry{
				{Key: "FOO", Value: "bar"},
				{Key: "BAZ", Value: "qux"},
			},
		},
		{
			name:  "double-quoted with escapes",
			input: `KEY="hello\nworld\t!"` + "\n",
			want: []vault.SecretEntry{
				{Key: "KEY", Value: "hello\nworld\t!"},
			},
		},
		{
			name:  "single-quoted literal",
			input: "KEY='hello\\nworld'\n",
			want: []vault.SecretEntry{
				{Key: "KEY", Value: `hello\nworld`},
			},
		},
		{
			name:  "multiline double-quoted",
			input: "KEY=\"line1\nline2\nline3\"\n",
			want: []vault.SecretEntry{
				{Key: "KEY", Value: "line1\nline2\nline3"},
			},
		},
		{
			name:  "variable substitution braces",
			input: "HOST=localhost\nURL=http://${HOST}:8080\n",
			want: []vault.SecretEntry{
				{Key: "HOST", Value: "localhost"},
				{Key: "URL", Value: "http://localhost:8080"},
			},
		},
		{
			name:  "variable substitution no braces",
			input: "HOST=localhost\nURL=http://$HOST:8080\n",
			want: []vault.SecretEntry{
				{Key: "HOST", Value: "localhost"},
				{Key: "URL", Value: "http://localhost:8080"},
			},
		},
		{
			name:  "variable substitution in double quotes",
			input: "NAME=world\nGREET=\"hello ${NAME}\"\n",
			want: []vault.SecretEntry{
				{Key: "NAME", Value: "world"},
				{Key: "GREET", Value: "hello world"},
			},
		},
		{
			name:  "no substitution in single quotes",
			input: "NAME=world\nGREET='hello ${NAME}'\n",
			want: []vault.SecretEntry{
				{Key: "NAME", Value: "world"},
				{Key: "GREET", Value: "hello ${NAME}"},
			},
		},
		{
			name:  "comments and empty lines",
			input: "# comment\n\nFOO=bar\n  # indented comment\nBAZ=qux\n",
			want: []vault.SecretEntry{
				{Key: "FOO", Value: "bar"},
				{Key: "BAZ", Value: "qux"},
			},
		},
		{
			name:  "inline comment unquoted",
			input: "FOO=bar # this is a comment\n",
			want: []vault.SecretEntry{
				{Key: "FOO", Value: "bar"},
			},
		},
		{
			name:  "malformed line no equals",
			input: "NOEQUALS\nFOO=bar\n",
			want: []vault.SecretEntry{
				{Key: "FOO", Value: "bar"},
			},
		},
		{
			name:  "empty value",
			input: "EMPTY=\n",
			want: []vault.SecretEntry{
				{Key: "EMPTY", Value: ""},
			},
		},
		{
			name:  "value with equals sign",
			input: "DSN=postgres://user:pass@host/db?sslmode=disable\n",
			want: []vault.SecretEntry{
				{Key: "DSN", Value: "postgres://user:pass@host/db?sslmode=disable"},
			},
		},
		{
			name:  "escaped quote in double quotes",
			input: `KEY="say \"hello\""` + "\n",
			want: []vault.SecretEntry{
				{Key: "KEY", Value: `say "hello"`},
			},
		},
		{
			name:  "undefined variable kept as-is",
			input: "FOO=${UNDEFINED}\n",
			want: []vault.SecretEntry{
				{Key: "FOO", Value: "${UNDEFINED}"},
			},
		},
		{
			name:  "mixed scenario",
			input: "# App config\nexport APP_NAME=myapp\nDATABASE_URL=\"postgres://localhost/${APP_NAME}\"\nSECRET='s3cr3t'\nDEBUG=true # enable debug\n",
			want: []vault.SecretEntry{
				{Key: "APP_NAME", Value: "myapp"},
				{Key: "DATABASE_URL", Value: "postgres://localhost/myapp"},
				{Key: "SECRET", Value: "s3cr3t"},
				{Key: "DEBUG", Value: "true"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseEnvData([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseEnvData() error = %v, wantErr %v", err, tt.wantErr)
			}

			if len(got) != len(tt.want) {
				t.Fatalf("parseEnvData() got %d entries, want %d\ngot:  %+v\nwant: %+v", len(got), len(tt.want), got, tt.want)
			}

			for i := range got {
				if got[i].Key != tt.want[i].Key || got[i].Value != tt.want[i].Value {
					t.Errorf("entry[%d] = {%q, %q}, want {%q, %q}", i, got[i].Key, got[i].Value, tt.want[i].Key, tt.want[i].Value)
				}
			}
		})
	}
}
