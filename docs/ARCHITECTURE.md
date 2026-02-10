# Architecture

Machine-bound encrypted vault CLI built in Go.

## High-Level Overview

```mermaid
graph TB
    User([User / Script])

    subgraph CLI["cmd/ — Cobra CLI"]
        Root["profile (root)"]
        VaultCmd["vault"]
        EnvCmd["env"]
        ProfileCmd["profile"]
    end

    subgraph Core["pkg/vault/ — Public API"]
        Vault["Vault"]
        EnvOps["Env Release"]
    end

    subgraph Internal["internal/"]
        Crypto["crypto/"]
        Sentinel["sentinel/"]
        Store["store/"]
        App["application/"]
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
    EnvCmd --> EnvOps
    Vault --> Crypto
    Vault --> Store
    EnvOps --> Sentinel
    Sentinel --> Crypto
    Store --> DB
    Sentinel --> FS
    App --> Store
    App --> Sentinel
```

## Command Tree

```mermaid
graph LR
    root["profile"]

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

    env --> password["password"]
    env --> release["release"]
    env --> revoke["revoke"]
    env --> estatus["status"]
    env --> eexport["export"]

    password --> pwset["set"]
    password --> pwreset["reset"]
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

    cmd --> vault
    vault --> store
    vault --> crypto
    vault --> sentinel
    vault --> app
    sentinel --> crypto
    sentinel --> app
    store --> sqlc
    store --> app
```

## Database Schema

```mermaid
erDiagram
    vault_sealed_key {
        int id PK "CHECK (id = 1)"
        blob sealed_data "encrypted master key"
        blob nonce "GCM nonce"
        blob key_salt "HKDF salt"
        int version "key version"
        blob machine_id_hash "SHA-256 of machine ID"
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

    vault_profiles ||--o{ vault_secrets : "CASCADE delete"
```

## Vault Init Flow

```mermaid
sequenceDiagram
    participant User
    participant CLI as vault init
    participant V as pkg/vault
    participant C as crypto
    participant S as store
    participant DB as SQLite

    User->>CLI: profile vault init
    CLI->>V: Init(opts)
    V->>S: Open(dbPath)
    S->>DB: CREATE TABLE IF NOT EXISTS ...
    S-->>V: Store

    V->>S: HasSealedKey()
    S-->>V: false

    V->>C: MachineID()
    C-->>V: "GUID-1234..."
    V->>C: MachineIDHash()
    C-->>V: SHA-256(machineID)

    V->>C: GenerateKey()
    C-->>V: masterKey (32 random bytes)
    V->>C: GenerateSalt()
    C-->>V: salt (32 random bytes)

    V->>C: SealMasterKey(masterKey, machineID, salt)
    Note over C: HKDF-SHA256(machineID, salt, "profile-vault-v1")<br/>AES-256-GCM encrypt masterKey
    C-->>V: sealedData, nonce

    V->>S: UpsertSealedKey(sealedData, nonce, salt, machineIDHash, 1)
    S->>DB: INSERT INTO vault_sealed_key
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
    S-->>V: VaultSealedKey{sealedData, nonce, salt, machineIDHash}

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

    V-->>CLI: &Vault{store, masterKey, dbPath}
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

## Key Derivation Chain

```mermaid
graph TD
    MID["Platform Machine ID<br/>(Windows: MachineGuid registry<br/>Linux: DMI product_id<br/>Darwin: serial number<br/>Fallback: hostname)"]
    Salt["Random Salt<br/>(32 bytes, stored in DB)"]

    MID --> HKDF["HKDF-SHA256<br/>info = 'profile-vault-v1'"]
    Salt --> HKDF
    HKDF --> DK["Derived Key<br/>(256-bit)"]

    MK["Master Key<br/>(256-bit, random)"]

    DK --> GCM["AES-256-GCM Encrypt"]
    MK --> GCM
    GCM --> Sealed["Sealed Data + Nonce<br/>(stored in vault_sealed_key)"]

    MID --> SHA["SHA-256"]
    SHA --> Hash["Machine ID Hash<br/>(stored for verification)"]

    style MK fill:#f9f,stroke:#333
    style DK fill:#bbf,stroke:#333
    style Sealed fill:#bfb,stroke:#333
```

## Env Release Flow (Password-Gated, Time-Limited)

```mermaid
sequenceDiagram
    participant User
    participant CLI
    participant V as Vault
    participant Sen as sentinel
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
        Sen->>Sen: crypto.Encrypt(masterKey, payload)
        Sen->>FS: Write nonce||ciphertext<br/>to PROFILE_ENV_RELEASE_ENABLED

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
        Sen->>Sen: Decrypt + unmarshal
        Sen->>Sen: Check expiry

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
            nonce (12B) || ciphertext
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
    participant CLI as profile CLI
    participant V as Vault
    participant Sen as sentinel

    Shell->>CLI: MY_KEY=$(profile --env-inline MY_KEY)
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
        SK["vault_sealed_key<br/>Encrypted master key + salt + hash"]
        PW["vault_password<br/>bcrypt hash"]
        PR["vault_profiles<br/>name, description, is_default"]
        SE["vault_secrets<br/>AES-256-GCM encrypted values"]
    end

    DB --> SK
    DB --> PW
    DB --> PR
    DB --> SE

    subgraph Paths["Platform Paths"]
        Win["Windows: %LOCALAPPDATA%/profile/"]
        Lin["Linux: ~/.config/profile/"]
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

All commands use `outputResult(cmd, jsonData, textFn)` for consistent dual-mode output. The `--json` flag is a persistent flag on the root command, inherited by all subcommands.

## Security Model

| Layer | Mechanism | Purpose |
|-------|-----------|---------|
| Machine binding | SHA-256(MachineID) stored at init, verified at open | Vault non-portable across machines |
| Master key sealing | HKDF-SHA256(machineID, salt) derives wrapping key | Master key never stored in plaintext |
| Secret encryption | AES-256-GCM per secret, random nonce | Each secret independently encrypted |
| Env release gate | bcrypt password + time-limited sentinel file | Secrets require explicit release |
| Sentinel integrity | Encrypted with master key, not guessable | Sentinel file is opaque without vault |
| Auto-revoke | Check() expires sessions past TTL | No background cleanup needed |
| Singleton tables | `CHECK (id = 1)` constraint | Exactly one sealed key and password |
| FK cascade | `ON DELETE CASCADE` on secrets | Deleting profile removes all its secrets |
| SQLite safety | WAL mode, busy timeout, max 1 connection, mutex | Concurrent access without corruption |
