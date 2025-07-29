package database

import (
	"errors"
	"strings"

	"github.com/lib/pq"
)

// IsDuplicateError checks if the error is a Postgres duplicate entry error (unique_violation).
// Returns the constraint/field name if duplicate, otherwise nil.
func IsDuplicateError(err error) *string {
	// For github.com/lib/pq
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		if pqErr.Code == "23505" {
			label := extractColumnLabel(pqErr.Constraint) + " already exist"
			return &label
		}
	}

	return nil
}

// ExtractColumnLabel converts "users_email_address_key" to "email address"
func extractColumnLabel(constraint string) string {
	constraint = strings.TrimSuffix(constraint, "_key")
	parts := strings.Split(constraint, "_")

	if len(parts) <= 1 {
		return constraint
	}

	// Remove the table name (assumed to be the first part)
	columnParts := parts[1:]
	return strings.Join(columnParts, " ")
}
