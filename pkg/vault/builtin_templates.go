package vault

// builtinTemplates defines the built-in secret templates shipped with anvil.
var builtinTemplates = []TemplateDefinition{
	{
		Name:        "postgres",
		Description: "PostgreSQL connection secrets",
		Variables: []TemplateVariable{
			{Name: "host", Description: "Database host", Required: true, Default: "localhost"},
			{Name: "port", Description: "Database port", Required: true, Default: "5432"},
			{Name: "database", Description: "Database name", Required: true},
			{Name: "username", Description: "Database user", Required: true},
			{Name: "password", Description: "Database password", Required: true, Secret: true},
		},
		Secrets: []TemplateSecret{
			{Key: "POSTGRES_HOST", Value: "{{.host}}", Description: "PostgreSQL host"},
			{Key: "POSTGRES_PORT", Value: "{{.port}}", Description: "PostgreSQL port"},
			{Key: "POSTGRES_DB", Value: "{{.database}}", Description: "PostgreSQL database"},
			{Key: "POSTGRES_USER", Value: "{{.username}}", Description: "PostgreSQL user"},
			{Key: "POSTGRES_PASSWORD", Value: "{{.password}}", Description: "PostgreSQL password"},
			{Key: "DATABASE_URL", Value: "postgres://{{.username}}:{{.password}}@{{.host}}:{{.port}}/{{.database}}", Description: "PostgreSQL connection URL"},
		},
	},
	{
		Name:        "mysql",
		Description: "MySQL connection secrets",
		Variables: []TemplateVariable{
			{Name: "host", Description: "Database host", Required: true, Default: "localhost"},
			{Name: "port", Description: "Database port", Required: true, Default: "3306"},
			{Name: "database", Description: "Database name", Required: true},
			{Name: "username", Description: "Database user", Required: true},
			{Name: "password", Description: "Database password", Required: true, Secret: true},
		},
		Secrets: []TemplateSecret{
			{Key: "MYSQL_HOST", Value: "{{.host}}", Description: "MySQL host"},
			{Key: "MYSQL_PORT", Value: "{{.port}}", Description: "MySQL port"},
			{Key: "MYSQL_DATABASE", Value: "{{.database}}", Description: "MySQL database"},
			{Key: "MYSQL_USER", Value: "{{.username}}", Description: "MySQL user"},
			{Key: "MYSQL_PASSWORD", Value: "{{.password}}", Description: "MySQL password"},
			{Key: "MYSQL_URL", Value: "{{.username}}:{{.password}}@tcp({{.host}}:{{.port}})/{{.database}}", Description: "MySQL connection DSN"},
		},
	},
	{
		Name:        "redis",
		Description: "Redis connection secrets",
		Variables: []TemplateVariable{
			{Name: "host", Description: "Redis host", Required: true, Default: "localhost"},
			{Name: "port", Description: "Redis port", Required: true, Default: "6379"},
			{Name: "password", Description: "Redis password", Secret: true},
		},
		Secrets: []TemplateSecret{
			{Key: "REDIS_HOST", Value: "{{.host}}", Description: "Redis host"},
			{Key: "REDIS_PORT", Value: "{{.port}}", Description: "Redis port"},
			{Key: "REDIS_PASSWORD", Value: "{{.password}}", Description: "Redis password"},
			{Key: "REDIS_URL", Value: "redis://:{{.password}}@{{.host}}:{{.port}}", Description: "Redis connection URL"},
		},
	},
	{
		Name:        "aws",
		Description: "AWS credential secrets",
		Variables: []TemplateVariable{
			{Name: "access_key_id", Description: "AWS access key ID", Required: true, Secret: true},
			{Name: "secret_access_key", Description: "AWS secret access key", Required: true, Secret: true},
			{Name: "region", Description: "AWS region", Required: true, Default: "us-east-1"},
		},
		Secrets: []TemplateSecret{
			{Key: "AWS_ACCESS_KEY_ID", Value: "{{.access_key_id}}", Description: "AWS access key ID"},
			{Key: "AWS_SECRET_ACCESS_KEY", Value: "{{.secret_access_key}}", Description: "AWS secret access key"},
			{Key: "AWS_DEFAULT_REGION", Value: "{{.region}}", Description: "AWS default region"},
		},
	},
	{
		Name:        "github-token",
		Description: "GitHub personal access token",
		Variables: []TemplateVariable{
			{Name: "token", Description: "GitHub personal access token", Required: true, Secret: true},
		},
		Secrets: []TemplateSecret{
			{Key: "GITHUB_TOKEN", Value: "{{.token}}", Description: "GitHub personal access token"},
		},
	},
}
