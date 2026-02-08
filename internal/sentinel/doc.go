// Package sentinel manages time-limited env release sessions using encrypted
// file-based state. Sessions are bound to the machine identity and automatically
// expire after the configured TTL.
package sentinel
