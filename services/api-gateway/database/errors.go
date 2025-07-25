package database

import (
	"errors"
)

var (
	// ErrNotFound is returned when a requested resource is not found
	ErrNotFound = errors.New("resource not found")

	// ErrAlreadyExists is returned when trying to create a resource that already exists
	ErrAlreadyExists = errors.New("resource already exists")

	// ErrInvalidInput is returned when the input data is invalid
	ErrInvalidInput = errors.New("invalid input data")

	// ErrDatabaseConnection is returned when there's a database connection issue
	ErrDatabaseConnection = errors.New("database connection error")

	// ErrTransactionFailed is returned when a database transaction fails
	ErrTransactionFailed = errors.New("database transaction failed")

	// ErrTimeout is returned when a database operation times out
	ErrTimeout = errors.New("database operation timeout")

	// ErrConstraintViolation is returned when a database constraint is violated
	ErrConstraintViolation = errors.New("database constraint violation")

	// ErrDeadlock is returned when a database deadlock occurs
	ErrDeadlock = errors.New("database deadlock")

	// ErrConnectionPoolExhausted is returned when the connection pool is exhausted
	ErrConnectionPoolExhausted = errors.New("connection pool exhausted")

	// ErrReadOnlyTransaction is returned when trying to write to a read-only transaction
	ErrReadOnlyTransaction = errors.New("read-only transaction")

	// ErrInvalidCredentials is returned when user credentials are invalid
	ErrInvalidCredentials = errors.New("invalid credentials")

	// ErrTokenExpired is returned when a token has expired
	ErrTokenExpired = errors.New("token expired")

	// ErrTokenRevoked is returned when a token has been revoked
	ErrTokenRevoked = errors.New("token revoked")

	// ErrInsufficientPermissions is returned when user lacks required permissions
	ErrInsufficientPermissions = errors.New("insufficient permissions")

	// ErrRateLimitExceeded is returned when rate limit is exceeded
	ErrRateLimitExceeded = errors.New("rate limit exceeded")

	// ErrServiceUnavailable is returned when the service is temporarily unavailable
	ErrServiceUnavailable = errors.New("service unavailable")

	// ErrDataCorruption is returned when data corruption is detected
	ErrDataCorruption = errors.New("data corruption detected")

	// ErrMigrationFailed is returned when database migration fails
	ErrMigrationFailed = errors.New("database migration failed")

	// ErrBackupFailed is returned when database backup fails
	ErrBackupFailed = errors.New("database backup failed")

	// ErrRestoreFailed is returned when database restore fails
	ErrRestoreFailed = errors.New("database restore failed")
) 