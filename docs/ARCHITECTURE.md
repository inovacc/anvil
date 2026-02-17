# Architecture

Machine-bound encrypted vault CLI built in Go.

## High-Level Overview

```mermaid
graph TB
    User([User / Script])

    subgraph CLI["cmd/ — Cobra CLI"]
        Root["anvil (root)"]
        VaultCmd["vault"]
        EnvCmd["env"]
        ProfileCmd["profile"]
    end

    subgraph Core["pkg/vault/ — Public API"]
        Vault["Vault"]
        EnvOps["Env Release"]
        Ifaces["Interfaces<br/>VaultReader / VaultWriter<br/>VaultEnv / VaultPassword"]
    end

    subgraph Internal["internal/"]
        Crypto["crypto/"]
        Sentinel["sentinel/"]
        Store["store/"]
        App["application/"]
        TUI["tui/"]
    end

    subgraph External["External"]
        Sealbox["sealbox<br/>(TPM 2.0 + AES-GCM)"]
        TPM["TPM 2.0 Hardware"]
    end

    subgraph Storage["Storage"]
        DB[(SQLite)]
        FS["Sentinel File"]
    end

    User --> Root
    Root --> VaultCmd
    VaultCmd --> EnvCmd
    VaultCmd --> ProfileCmd
    VaultCmd --> Vault
    VaultCmd --> TUI
    TUI --> Vault
    EnvCmd --> EnvOps
    Vault --> Crypto
    Vault --> Store
    EnvOps --> Sentinel
    Sentinel --> Sealbox
    Crypto --> Sealbox
    Sealbox -.-> TPM
    Store --> DB
    Sentinel --> FS
```

## Command Tree

```mermaid
graph LR
    root["anvil"]

    root --> envInline["--env-inline KEY"]
    root --> vault["vault"]

    vault --> init["init"]
    vault --> status["status"]
    vault --> set["set"]
    vault --> get["get"]
    vault --> list["list"]
    vault --> delete["delete"]
    vault --> export["export"]
    vault --> import_["import"]
    vault --> prof["profile"]
    vault --> env["env"]

    prof --> pcreate["create"]
    prof --> plist["list"]
    prof --> pdelete["delete"]
    prof --> puse["use"]

    vault --> tmpl["template"]
    vault --> plugin["plugin"]
    vault --> rotate["rotate-key"]
    vault --> audit["audit"]
    vault --> history["history"]
    vault --> backup["backup"]
    vault --> share["share"]
    vault --> docker["docker"]
    vault --> tui["tui"]
    vault --> gather["gather"]
    vault --> rollback["rollback"]
    vault --> seal["seal"]
    vault --> unseal["unseal"]

    env --> password["password"]
    env --> release["release"]
    env --> revoke["revoke"]
    env --> estatus["status"]
    env --> eexport["export"]

    password --> pwset["set"]
    password --> pwreset["reset"]

    tmpl --> tcreate["create"]
    tmpl --> tlist["list"]
    tmpl --> tshow["show"]
    tmpl --> tdelete["delete"]
    tmpl --> tapply["apply"]

    plugin --> pllist["list"]
    plugin --> plhookadd["hook-add"]
    plugin --> plhookrm["hook-remove"]
    plugin --> plprovadd["provider-add"]
    plugin --> plprovrm["provider-remove"]

    share --> sexport["export"]
    share --> simport["import"]

    docker --> dexport["export"]
    docker --> dclean["clean"]
    docker --> dcompose["compose"]
```

## Package Dependency Graph

```mermaid
graph TD
    cmd["cmd/"]
    vault["pkg/vault/"]
    store["internal/store/"]
    crypto["internal/crypto/"]
    sentinel["internal/sentinel/"]
    app["internal/application/"]
    sqlc["internal/store/sqlc/"]
    sealbox["sealbox<br/>(external)"]

    tui_pkg["internal/tui/"]
    cmd --> vault
    cmd --> tui_pkg
    tui_pkg --> vault
    vault --> store
    vault --> crypto
    vault --> sentinel
    vault --> app
    sentinel --> crypto
    sentinel --> sealbox
    sentinel --> app
    crypto --> sealbox
    store --> sqlc
```

## Database Schema

```mermaid
erDiagram
    vault_sealed_key {
        int id PK "CHECK (id = 1)"
        blob sealed_data "encrypted master key or TPM SealedData JSON"
        blob nonce "GCM nonce (software) or NULL (TPM)"
        blob key_salt "HKDF salt (software) or NULL (TPM)"
        int version "key version"
        blob machine_id_hash "SHA-256 of machine ID"
        text seal_method "tpm or software (default)"
        datetime created_at
        datetime updated_at
    }

    vault_password {
        int id PK "CHECK (id = 1)"
        blob password_hash "bcrypt hash"
        datetime created_at
        datetime updated_at
    }

    vault_profiles {
        int id PK
        text name UK "unique name"
        text description
        int is_default "0 or 1"
        datetime created_at
        datetime updated_at
    }

    vault_secrets {
        int id PK
        text profile_name FK
        text key
        blob encrypted_value "AES-256-GCM"
        blob nonce "GCM nonce"
        text description
        datetime created_at
        datetime updated_at
    }

    vault_secret_versions {
        int id PK
        text profile_name FK
        text key
        int version
        blob encrypted_value "AES-256-GCM"
        blob nonce "GCM nonce"
        datetime created_at
    }

    vault_audit_log {
        int id PK
        text action
        text profile_name
        text secret_key
        text detail
        datetime created_at
    }

    vault_templates {
        int id PK
        text name UK "unique name"
        text description
        text template_data "JSON TemplateDefinition"
        int builtin "0 or 1"
        datetime created_at
        datetime updated_at
    }

    vault_profiles ||--o{ vault_secrets : "CASCADE delete"
    vault_profiles ||--o{ vault_secret_versions : "CASCADE delete"
```

## Vault Init Flow

```mermaid
sequenceDiagram
    participant User
    participant CLI as vault init
    participant V as pkg/vault
    participant C as crypto
    participant SB as sealbox
    participant TPM as TPM 2.0
    participant S as store
    participant DB as SQLite

    User->>CLI: anvil vault init
    CLI->>V: Init(opts)
    V->>S: Open(dbPath)
    S->>DB: CREATE TABLE IF NOT EXISTS ...
    S-->>V: Store

    V->>S: HasSealedKey()
    S-->>V: false

    V->>C: MachineIDHash()
    C-->>V: SHA-256(machineID)

    V->>C: GenerateKey()
    C-->>V: masterKey (32 random bytes)

    V->>C: IsTPMAvailable()

    alt TPM available
        V->>C: SealMasterKeyTPM(masterKey)
        C->>SB: NewKeyManager()
        SB->>TPM: Open TPM device
        TPM-->>SB: handle
        C->>SB: km.SealKey(masterKey)
        SB->>TPM: Create sealed object
        TPM-->>SB: SealedData
        SB-->>C: *SealedData
        C->>C: json.Marshal(SealedData)
        C->>SB: km.Close()
        C-->>V: sealedJSON

        V->>S: UpsertSealedKey(sealedJSON, nil, nil, hash, 1, "tpm")
        S->>DB: INSERT INTO vault_sealed_key
    else TPM unavailable or seal failed
        V->>C: MachineID()
        C-->>V: "GUID-1234..."
        V->>C: GenerateSalt()
        C-->>V: salt (32 random bytes)
        V->>C: SealMasterKey(masterKey, machineID, salt)
        Note over C: HKDF-SHA256(machineID, salt)<br/>AES-256-GCM encrypt masterKey
        C-->>V: sealedData, nonce

        V->>S: UpsertSealedKey(sealedData, nonce, salt, hash, 1, "software")
        S->>DB: INSERT INTO vault_sealed_key
    end

    V->>C: ZeroBytes(masterKey)
    V->>S: Close()
    V-->>CLI: nil
    CLI-->>User: Vault initialized
```

## Vault Open Flow

```mermaid
sequenceDiagram
    participant CLI
    participant V as pkg/vault
    participant C as crypto
    participant SB as sealbox
    participant TPM as TPM 2.0
    participant S as store
    participant DB as SQLite

    CLI->>V: Open(opts)
    V->>S: Open(dbPath)
    S->>DB: Connect (WAL mode)
    S-->>V: Store

    V->>S: HasSealedKey()
    S-->>V: true

    V->>S: GetSealedKey()
    S->>DB: SELECT * FROM vault_sealed_key WHERE id = 1
    DB-->>S: row
    S-->>V: VaultSealedKey{sealedData, nonce, salt, machineIDHash, sealMethod}

    alt sealMethod == "tpm"
        V->>C: UnsealMasterKeyTPM(sealedData)
        C->>C: json.Unmarshal → SealedData
        C->>SB: NewKeyManager()
        SB->>TPM: Open TPM device
        C->>SB: km.UnsealKey(&sealed)
        SB->>TPM: Unseal object
        TPM-->>SB: masterKey
        SB-->>C: masterKey
        C->>SB: km.Close()
        C-->>V: masterKey
    else sealMethod == "software" (default)
        V->>C: MachineIDHash()
        C-->>V: currentHash
        Note over V: Compare currentHash == stored machineIDHash
        alt mismatch
            V-->>CLI: ErrMachineMismatch
        end
        V->>C: MachineID()
        C-->>V: machineID
        V->>C: UnsealMasterKey(sealedData, nonce, machineID, salt)
        Note over C: DeriveKey(machineID, salt) via HKDF<br/>AES-256-GCM decrypt sealedData
        C-->>V: masterKey
    end

    V-->>CLI: &Vault{store, masterKey, dbPath}
```

## Vault Close Flow

```mermaid
sequenceDiagram
    participant Caller
    participant V as Vault
    participant C as crypto
    participant SB as sealbox
    participant S as store

    Caller->>V: Close()
    V->>C: ZeroBytes(masterKey)
    C->>SB: SecureZero(masterKey)
    Note over SB: Overwrite all bytes with 0x00
    V->>S: Close()
    S-->>V: nil
    V-->>Caller: nil
```

## Secret Encrypt / Decrypt

```mermaid
sequenceDiagram
    participant User
    participant V as Vault
    participant C as crypto
    participant S as store

    rect rgb(235, 245, 255)
        Note over User,S: Set (encrypt + store)
        User->>V: Set("API_KEY", "sk-123", "prod", "")
        V->>V: resolveProfile("prod")
        V->>C: Encrypt(masterKey, "sk-123")
        Note over C: AES-256-GCM<br/>random 12-byte nonce
        C-->>V: ciphertext, nonce
        V->>S: UpsertSecret("prod", "API_KEY", ciphertext, nonce)
        V-->>User: nil
    end

    rect rgb(255, 245, 235)
        Note over User,S: Get (fetch + decrypt)
        User->>V: Get("API_KEY", "prod")
        V->>V: resolveProfile("prod")
        V->>S: GetSecret("prod", "API_KEY")
        S-->>V: VaultSecret{encryptedValue, nonce}
        V->>C: Decrypt(masterKey, encryptedValue, nonce)
        C-->>V: "sk-123"
        V-->>User: "sk-123"
    end
```

## Master Key Sealing

```mermaid
graph TD
    MK["Master Key<br/>(256-bit, random)"]

    MK --> Decision{"TPM 2.0<br/>available?"}

    Decision -->|Yes| TPMPath
    Decision -->|No / Failed| SWPath

    subgraph TPMPath["TPM Path (seal_method = 'tpm')"]
        direction TB
        KM["sealbox.NewKeyManager()"]
        KM --> Seal["km.SealKey(masterKey)"]
        Seal --> SD["SealedData struct<br/>(public_area, private_area,<br/>sealed_blob_public)"]
        SD --> JSON["JSON marshal"]
        JSON --> Store1["sealed_data = JSON blob<br/>nonce = NULL<br/>key_salt = NULL"]
    end

    subgraph SWPath["Software Path (seal_method = 'software')"]
        direction TB
        MID["Platform Machine ID"]
        Salt["Random Salt (32 bytes)"]
        MID --> HKDF["HKDF-SHA256<br/>info = 'profile-vault-v1'"]
        Salt --> HKDF
        HKDF --> DK["Derived Key (256-bit)"]
        DK --> GCM["AES-256-GCM Encrypt"]
        GCM --> Store2["sealed_data = ciphertext<br/>nonce = GCM nonce<br/>key_salt = salt"]
    end

    MID2["Machine ID"] --> SHA["SHA-256"]
    SHA --> Hash["machine_id_hash<br/>(stored for verification)"]

    Store1 --> DB[(vault_sealed_key)]
    Store2 --> DB
    Hash --> DB

    style MK fill:#f9f,stroke:#333
    style TPMPath fill:#e8f5e9,stroke:#4caf50
    style SWPath fill:#e3f2fd,stroke:#2196f3
```

## Env Release Flow (Password-Gated, Time-Limited)

```mermaid
sequenceDiagram
    participant User
    participant CLI
    participant V as Vault
    participant Sen as sentinel
    participant SB as sealbox
    participant FS as Filesystem

    rect rgb(235, 255, 235)
        Note over User,FS: 1. Set password (one-time)
        User->>CLI: vault env password set
        CLI->>CLI: Prompt for password (hidden)
        CLI->>V: SetPassword(password)
        Note over V: bcrypt(password, cost=10)
        V-->>CLI: OK
    end

    rect rgb(235, 245, 255)
        Note over User,FS: 2. Release secrets (time-limited)
        User->>CLI: vault env release --ttl 15m
        CLI->>CLI: Prompt for password
        CLI->>V: EnvRelease(password, opts)
        V->>V: VerifyPassword(password) via bcrypt
        V->>V: resolveProfile(name)
        V->>Sen: Release(masterKey, profile, 15m)

        Sen->>Sen: Build SessionPayload JSON
        Note over Sen: {profile, expires_at, session_id,<br/>machine_hash, created_at}
        Sen->>SB: sealbox.Encrypt(masterKey, payload)
        Note over SB: AES-256-GCM, returns nonce||ciphertext
        SB-->>Sen: packed blob
        Sen->>FS: Write packed blob<br/>to PROFILE_ENV_RELEASE_ENABLED

        Sen-->>V: ReleaseState{Active: true}
        V-->>CLI: state
        CLI-->>User: Released for 15m
    end

    rect rgb(255, 245, 235)
        Note over User,FS: 3. Use released secrets
        User->>CLI: vault env export
        CLI->>V: EnvExport(profile)
        V->>Sen: Check(masterKey)
        Sen->>FS: Read PROFILE_ENV_RELEASE_ENABLED
        Sen->>SB: sealbox.Decrypt(masterKey, fileData)
        SB-->>Sen: plaintext JSON
        Sen->>Sen: Unmarshal + check expiry

        alt expired
            Sen->>FS: Rename to DISABLED (auto-revoke)
            Sen-->>V: Active: false
            V-->>CLI: ErrNotReleased
        else valid
            Sen-->>V: Active: true
            V->>V: Export(profile) — decrypt all secrets
            V-->>CLI: []SecretEntry
            CLI-->>User: KEY=value (shell format)
        end
    end

    rect rgb(255, 235, 235)
        Note over User,FS: 4. Manual revoke
        User->>CLI: vault env revoke
        CLI->>V: EnvRevoke()
        V->>Sen: Revoke()
        Sen->>FS: Rename ENABLED to DISABLED (atomic)
        Sen-->>V: nil
        V-->>CLI: OK
        CLI-->>User: Release revoked
    end
```

## Sentinel File Lifecycle

```mermaid
stateDiagram-v2
    [*] --> NoFile: initial state

    NoFile --> Enabled: Release()<br/>write encrypted session
    Enabled --> Disabled: Revoke()<br/>rename (atomic)
    Enabled --> Disabled: Check() auto-revoke<br/>(TTL expired)
    Disabled --> Enabled: Release()<br/>overwrite + remove disabled
    Enabled --> Enabled: Release()<br/>overwrite (new session)

    state Enabled {
        note right of Enabled
            PROFILE_ENV_RELEASE_ENABLED
            sealbox packed: nonce (12B) || ciphertext
            Encrypted SessionPayload JSON
        end note
    }

    state Disabled {
        note right of Disabled
            PROFILE_ENV_RELEASE_DISABLED
            (same encrypted content, now inert)
        end note
    }
```

## Inline Secret Retrieval

```mermaid
sequenceDiagram
    participant Shell as Shell Script
    participant CLI as anvil CLI
    participant V as Vault
    participant Sen as sentinel

    Shell->>CLI: MY_KEY=$(anvil --env-inline MY_KEY)
    CLI->>V: Open(nil)
    V-->>CLI: vault

    CLI->>V: EnvInlineGet("MY_KEY", "")
    V->>Sen: Check(masterKey)

    alt not released
        Sen-->>V: Active: false
        V-->>CLI: ErrNotReleased
        CLI-->>Shell: exit 1
    else active
        Sen-->>V: Active: true, ProfileName: "prod"
        V->>V: Get("MY_KEY", "prod")
        V-->>CLI: "sk-123"
        CLI-->>Shell: sk-123 (raw, no newline wrapping)
    end
```

## Data At Rest

```mermaid
graph LR
    subgraph AppDir["Application Directory"]
        direction TB
        DB["vault.db<br/>(SQLite, WAL mode)"]
        EN["PROFILE_ENV_RELEASE_ENABLED"]
        DIS["PROFILE_ENV_RELEASE_DISABLED"]
    end

    subgraph Tables["SQLite Tables"]
        direction TB
        SK["vault_sealed_key<br/>TPM SealedData JSON or software-sealed blob<br/>+ seal_method discriminator"]
        PW["vault_password<br/>bcrypt hash"]
        PR["vault_profiles<br/>name, description, is_default"]
        SE["vault_secrets<br/>AES-256-GCM encrypted values"]
        SV["vault_secret_versions<br/>Archived secret values"]
        AL["vault_audit_log<br/>Action history"]
        TM["vault_templates<br/>Secret templates (JSON)"]
    end

    DB --> SK
    DB --> PW
    DB --> PR
    DB --> SE

    subgraph Paths["Platform Paths"]
        Win["Windows: %LOCALAPPDATA%/anvil/"]
        Lin["Linux: ~/.config/anvil/"]
    end

    Paths --> AppDir
```

## Output Modes

```mermaid
graph TD
    Cmd["Any CLI Command"] --> Check{"--json flag?"}

    Check -->|yes| JSON["outputResult: json.Encode(data)"]
    Check -->|no| Text["outputResult: textFn()"]

    JSON --> Stdout["stdout"]
    Text --> Stdout
```

All commands use `outputResult(cmd, jsonData, textFn)` for consistent dual-mode output. The `--json` flag is a
persistent flag on the root command, inherited by all subcommands.

## Error Handling

```mermaid
graph TD
    Err["Command returns error"] --> Extract{"errors.As<br/>*vault.UserError?"}

    Extract -->|yes| UE["UserError<br/>{Message, Hint}"]
    Extract -->|no| Raw["Raw error string"]

    UE --> Mode1{"--json flag?"}
    Raw --> Mode2{"--json flag?"}

    Mode1 -->|yes| UEJSON["{\"error\": \"...\", \"hint\": \"...\"}"]
    Mode1 -->|no| UEText["Error: message\nHint:  hint"]

    Raw --> Mode2
    Mode2 -->|yes| RawJSON["{\"error\": \"...\"}"]
    Mode2 -->|no| RawText["Error: message"]

    UEJSON --> Stderr["stderr + exit 1"]
    UEText --> Stderr
    RawJSON --> Stderr
    RawText --> Stderr
```

- `SilenceErrors` and `SilenceUsage` are set on rootCmd — Cobra never dumps usage text on errors
- `handleError` in `cmd/errors.go` extracts `UserError` via `errors.As` and formats output
- All sentinel errors in `pkg/vault/errors.go` are `*UserError` values with message and optional hint
- Non-`UserError` errors (unexpected/system errors) pass through with the raw error message

## Security Model

| Layer                     | Mechanism                                                           | Purpose                                                                   |
|---------------------------|---------------------------------------------------------------------|---------------------------------------------------------------------------|
| TPM key sealing           | `sealbox.NewKeyManager().SealKey()` via TPM 2.0                     | Master key hardware-bound; cannot be extracted even with full disk access |
| Software fallback         | HKDF-SHA256(machineID, salt) derives wrapping key                   | Fallback for machines without TPM (macOS, VMs)                            |
| Seal method discriminator | `seal_method` column (`"tpm"` or `"software"`)                      | Vault knows which unseal path to use                                      |
| Machine binding           | SHA-256(MachineID) stored at init, verified at open (software path) | Vault non-portable across machines                                        |
| Secret encryption         | AES-256-GCM per secret, random nonce                                | Each secret independently encrypted                                       |
| Sentinel encryption       | `sealbox.Encrypt` (packed nonce\|\|ciphertext)                      | Sentinel file is opaque without vault master key                          |
| Env release gate          | bcrypt password + time-limited sentinel file                        | Secrets require explicit release                                          |
| Auto-revoke               | Check() expires sessions past TTL                                   | No background cleanup needed                                              |
| Memory zeroing            | `sealbox.SecureZero()` on vault Close and after Init                | Master key does not linger in process memory                              |
| Singleton tables          | `CHECK (id = 1)` constraint                                         | Exactly one sealed key and password                                       |
| FK cascade                | `ON DELETE CASCADE` on secrets                                      | Deleting profile removes all its secrets                                  |
| SQLite safety             | WAL mode, busy timeout, max 1 connection, mutex                     | Concurrent access without corruption                                      |
