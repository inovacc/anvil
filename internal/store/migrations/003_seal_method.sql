-- Add seal method discriminator to vault_sealed_key
-- Values: "software" (default, HKDF-based) or "tpm" (TPM 2.0 hardware-backed)
ALTER TABLE vault_sealed_key
    ADD COLUMN seal_method TEXT NOT NULL DEFAULT 'software';
