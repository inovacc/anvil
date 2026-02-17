package vault

import (
	"log/slog"
	"time"
)

func (v *Vault) logAudit(action, profileName, secretKey, detail string) {
	if err := v.store.LogAudit(action, profileName, secretKey, detail); err != nil {
		slog.Error("audit log failed", "action", action, "error", err)
	}
}

// AuditLog returns the most recent audit log entries.
func (v *Vault) AuditLog(limit int64) ([]AuditEntry, error) {
	rows, err := v.store.ListAuditLog(limit)
	if err != nil {
		return nil, err
	}

	entries := make([]AuditEntry, 0, len(rows))
	for _, r := range rows {
		e := AuditEntry{
			Action:      r.Action,
			ProfileName: r.ProfileName,
			CreatedAt:   r.CreatedAt,
		}
		if r.SecretKey != nil {
			e.SecretKey = *r.SecretKey
		}

		if r.Detail != nil {
			e.Detail = *r.Detail
		}

		entries = append(entries, e)
	}

	return entries, nil
}

// AuditLogByProfile returns audit log entries filtered by profile name.
func (v *Vault) AuditLogByProfile(profileName string, limit int64) ([]AuditEntry, error) {
	rows, err := v.store.ListAuditLogByProfile(profileName, limit)
	if err != nil {
		return nil, err
	}

	entries := make([]AuditEntry, 0, len(rows))
	for _, r := range rows {
		e := AuditEntry{
			Action:      r.Action,
			ProfileName: r.ProfileName,
			CreatedAt:   r.CreatedAt,
		}
		if r.SecretKey != nil {
			e.SecretKey = *r.SecretKey
		}

		if r.Detail != nil {
			e.Detail = *r.Detail
		}

		entries = append(entries, e)
	}

	return entries, nil
}

// PurgeAuditLog deletes audit log entries older than the given time.
func (v *Vault) PurgeAuditLog(before time.Time) error {
	return v.store.PurgeAuditLog(before)
}
