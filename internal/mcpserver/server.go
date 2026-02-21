package mcpserver

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/inovacc/anvil/pkg/vault"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Server wraps an MCP server with a vault backend.
type Server struct {
	MCP   *mcp.Server
	vault *vault.Vault
}

// New creates a new MCP server backed by the given vault options.
func New(opts *vault.Options) (*Server, error) {
	v, err := vault.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("open vault: %w", err)
	}

	srv := mcp.NewServer(&mcp.Implementation{
		Name:    "anvil",
		Version: "1.0.0",
	}, nil)

	s := &Server{MCP: srv, vault: v}
	s.registerTools()
	s.registerResources()
	return s, nil
}

// Run starts the MCP server on stdio transport, blocking until context is cancelled.
func (s *Server) Run(ctx context.Context) error {
	defer func() { _ = s.vault.Close() }()
	return s.MCP.Run(ctx, &mcp.StdioTransport{})
}

// Close closes the underlying vault.
func (s *Server) Close() error {
	return s.vault.Close()
}

func (s *Server) registerTools() {
	s.registerSecretTools()
	s.registerProfileTools()
	s.registerKeyTools()
	s.registerSigningTools()
	s.registerAuditTools()
	s.registerStatusTools()
}

func (s *Server) registerResources() {
	s.MCP.AddResource(&mcp.Resource{
		URI:         "anvil://status",
		Name:        "Vault Status",
		Description: "Current vault status information",
		MIMEType:    "application/json",
	}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		st, err := s.vault.Status()
		if err != nil {
			return nil, err
		}
		data, _ := json.Marshal(st)
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{{
				URI:      "anvil://status",
				MIMEType: "application/json",
				Text:     string(data),
			}},
		}, nil
	})
}

// Secret tools

type secretGetInput struct {
	Key     string `json:"key" jsonschema:"the secret key to retrieve"`
	Profile string `json:"profile,omitempty" jsonschema:"target profile (uses default if empty)"`
}

type secretGetOutput struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type secretSetInput struct {
	Key         string `json:"key" jsonschema:"the secret key"`
	Value       string `json:"value" jsonschema:"the secret value"`
	Profile     string `json:"profile,omitempty" jsonschema:"target profile (uses default if empty)"`
	Description string `json:"description,omitempty" jsonschema:"optional description"`
}

type secretSetOutput struct {
	Key    string `json:"key"`
	Status string `json:"status"`
}

type secretDeleteInput struct {
	Key     string `json:"key" jsonschema:"the secret key to delete"`
	Profile string `json:"profile,omitempty" jsonschema:"target profile (uses default if empty)"`
}

type secretDeleteOutput struct {
	Key    string `json:"key"`
	Status string `json:"status"`
}

type secretListInput struct {
	Profile string `json:"profile,omitempty" jsonschema:"target profile (uses default if empty)"`
}

type secretListOutput struct {
	Secrets []vault.SecretInfo `json:"secrets"`
}

func (s *Server) registerSecretTools() {
	mcp.AddTool(s.MCP, &mcp.Tool{
		Name:        "secret_get",
		Description: "Get a decrypted secret value by key",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input secretGetInput) (*mcp.CallToolResult, secretGetOutput, error) {
		val, err := s.vault.Get(input.Key, input.Profile)
		if err != nil {
			return nil, secretGetOutput{}, err
		}
		return nil, secretGetOutput{Key: input.Key, Value: val}, nil
	})

	mcp.AddTool(s.MCP, &mcp.Tool{
		Name:        "secret_set",
		Description: "Set or update a secret value",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input secretSetInput) (*mcp.CallToolResult, secretSetOutput, error) {
		if err := s.vault.Set(input.Key, input.Value, input.Profile, input.Description); err != nil {
			return nil, secretSetOutput{}, err
		}
		return nil, secretSetOutput{Key: input.Key, Status: "created"}, nil
	})

	mcp.AddTool(s.MCP, &mcp.Tool{
		Name:        "secret_delete",
		Description: "Delete a secret by key",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input secretDeleteInput) (*mcp.CallToolResult, secretDeleteOutput, error) {
		if err := s.vault.Delete(input.Key, input.Profile); err != nil {
			return nil, secretDeleteOutput{}, err
		}
		return nil, secretDeleteOutput{Key: input.Key, Status: "deleted"}, nil
	})

	mcp.AddTool(s.MCP, &mcp.Tool{
		Name:        "secret_list",
		Description: "List all secrets in a profile",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input secretListInput) (*mcp.CallToolResult, secretListOutput, error) {
		secrets, err := s.vault.List(input.Profile)
		if err != nil {
			return nil, secretListOutput{}, err
		}
		return nil, secretListOutput{Secrets: secrets}, nil
	})
}

// Profile tools

type profileCreateInput struct {
	Name        string `json:"name" jsonschema:"profile name"`
	Description string `json:"description,omitempty" jsonschema:"optional description"`
	IsDefault   bool   `json:"is_default,omitempty" jsonschema:"make this the default profile"`
}

type profileCreateOutput struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

type profileDeleteInput struct {
	Name string `json:"name" jsonschema:"profile name to delete"`
}

type profileDeleteOutput struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

type profileListOutput struct {
	Profiles []vault.ProfileInfo `json:"profiles"`
}

func (s *Server) registerProfileTools() {
	mcp.AddTool(s.MCP, &mcp.Tool{
		Name:        "profile_list",
		Description: "List all profiles",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input any) (*mcp.CallToolResult, profileListOutput, error) {
		profiles, err := s.vault.ListProfiles()
		if err != nil {
			return nil, profileListOutput{}, err
		}
		return nil, profileListOutput{Profiles: profiles}, nil
	})

	mcp.AddTool(s.MCP, &mcp.Tool{
		Name:        "profile_create",
		Description: "Create a new profile",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input profileCreateInput) (*mcp.CallToolResult, profileCreateOutput, error) {
		if err := s.vault.CreateProfile(input.Name, input.Description, input.IsDefault); err != nil {
			return nil, profileCreateOutput{}, err
		}
		return nil, profileCreateOutput{Name: input.Name, Status: "created"}, nil
	})

	mcp.AddTool(s.MCP, &mcp.Tool{
		Name:        "profile_delete",
		Description: "Delete a profile",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input profileDeleteInput) (*mcp.CallToolResult, profileDeleteOutput, error) {
		if err := s.vault.DeleteProfile(input.Name); err != nil {
			return nil, profileDeleteOutput{}, err
		}
		return nil, profileDeleteOutput{Name: input.Name, Status: "deleted"}, nil
	})
}

// Key management tools

type keyGenerateInput struct {
	Name        string `json:"name" jsonschema:"key name"`
	Algorithm   string `json:"algorithm,omitempty" jsonschema:"ed25519 or ecdsa-p256 (default: ed25519)"`
	Description string `json:"description,omitempty" jsonschema:"optional description"`
}

type keyDeleteInput struct {
	Name      string `json:"name" jsonschema:"key name"`
	Algorithm string `json:"algorithm" jsonschema:"key algorithm (ed25519 or ecdsa-p256)"`
}

type keyDeleteOutput struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

type keyListOutput struct {
	Keys []vault.KeyInfo `json:"keys"`
}

type keyExportInput struct {
	Name    string `json:"name" jsonschema:"key name"`
	Private bool   `json:"private,omitempty" jsonschema:"export private key (default: public only)"`
}

type keyExportOutput struct {
	Name string `json:"name"`
	PEM  string `json:"pem"`
}

func (s *Server) registerKeyTools() {
	mcp.AddTool(s.MCP, &mcp.Tool{
		Name:        "key_generate",
		Description: "Generate an asymmetric key pair",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input keyGenerateInput) (*mcp.CallToolResult, *vault.KeyInfo, error) {
		algo := vault.AlgorithmEd25519
		if input.Algorithm != "" {
			algo = input.Algorithm
		}
		info, err := s.vault.GenerateKey(input.Name, algo, input.Description)
		if err != nil {
			return nil, nil, err
		}
		return nil, info, nil
	})

	mcp.AddTool(s.MCP, &mcp.Tool{
		Name:        "key_list",
		Description: "List all asymmetric keys",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input any) (*mcp.CallToolResult, keyListOutput, error) {
		keys, err := s.vault.ListKeys()
		if err != nil {
			return nil, keyListOutput{}, err
		}
		return nil, keyListOutput{Keys: keys}, nil
	})

	mcp.AddTool(s.MCP, &mcp.Tool{
		Name:        "key_delete",
		Description: "Delete an asymmetric key pair",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input keyDeleteInput) (*mcp.CallToolResult, keyDeleteOutput, error) {
		if err := s.vault.DeleteKey(input.Name, input.Algorithm); err != nil {
			return nil, keyDeleteOutput{}, err
		}
		return nil, keyDeleteOutput{Name: input.Name, Status: "deleted"}, nil
	})

	mcp.AddTool(s.MCP, &mcp.Tool{
		Name:        "key_export",
		Description: "Export a key in PEM format",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input keyExportInput) (*mcp.CallToolResult, keyExportOutput, error) {
		pem, err := s.vault.ExportKeyPEM(input.Name, input.Private)
		if err != nil {
			return nil, keyExportOutput{}, err
		}
		return nil, keyExportOutput{Name: input.Name, PEM: string(pem)}, nil
	})
}

// Signing tools

type signInput struct {
	KeyName string `json:"key_name" jsonschema:"name of the signing key"`
	Data    string `json:"data" jsonschema:"data to sign (UTF-8 string)"`
}

type verifyInput struct {
	KeyName   string `json:"key_name" jsonschema:"name of the verification key"`
	Data      string `json:"data" jsonschema:"original data (UTF-8 string)"`
	Signature string `json:"signature" jsonschema:"base64-encoded signature"`
}

func (s *Server) registerSigningTools() {
	mcp.AddTool(s.MCP, &mcp.Tool{
		Name:        "sign",
		Description: "Sign data with an asymmetric key",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input signInput) (*mcp.CallToolResult, *vault.SignResult, error) {
		result, err := s.vault.Sign(input.KeyName, []byte(input.Data))
		if err != nil {
			return nil, nil, err
		}
		return nil, result, nil
	})

	mcp.AddTool(s.MCP, &mcp.Tool{
		Name:        "verify",
		Description: "Verify a signature against data and key",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input verifyInput) (*mcp.CallToolResult, *vault.VerifyResult, error) {
		result, err := s.vault.Verify(input.KeyName, []byte(input.Data), input.Signature)
		if err != nil {
			return nil, nil, err
		}
		return nil, result, nil
	})
}

// Audit tools

type auditLogInput struct {
	Limit   int64  `json:"limit,omitempty" jsonschema:"max entries to return (default: 50)"`
	Profile string `json:"profile,omitempty" jsonschema:"filter by profile name"`
}

type auditLogOutput struct {
	Entries []vault.AuditEntry `json:"entries"`
}

func (s *Server) registerAuditTools() {
	mcp.AddTool(s.MCP, &mcp.Tool{
		Name:        "audit_log",
		Description: "View audit log entries",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input auditLogInput) (*mcp.CallToolResult, auditLogOutput, error) {
		limit := input.Limit
		if limit <= 0 {
			limit = 50
		}

		var entries []vault.AuditEntry
		var err error
		if input.Profile != "" {
			entries, err = s.vault.AuditLogByProfile(input.Profile, limit)
		} else {
			entries, err = s.vault.AuditLog(limit)
		}
		if err != nil {
			return nil, auditLogOutput{}, err
		}
		return nil, auditLogOutput{Entries: entries}, nil
	})
}

// Status tools

type installationIDOutput struct {
	ID string `json:"id"`
}

func (s *Server) registerStatusTools() {
	mcp.AddTool(s.MCP, &mcp.Tool{
		Name:        "vault_status",
		Description: "Get vault status information",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input any) (*mcp.CallToolResult, *vault.Status, error) {
		st, err := s.vault.Status()
		if err != nil {
			return nil, nil, err
		}
		return nil, st, nil
	})

	mcp.AddTool(s.MCP, &mcp.Tool{
		Name:        "installation_id",
		Description: "Get the vault installation ID",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input any) (*mcp.CallToolResult, installationIDOutput, error) {
		id, err := s.vault.InstallationID()
		if err != nil {
			return nil, installationIDOutput{}, err
		}
		return nil, installationIDOutput{ID: id}, nil
	})
}
