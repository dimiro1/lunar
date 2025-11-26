package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// SQLiteDB is a SQLite implementation of the DB interface
type SQLiteDB struct {
	db *sql.DB
}

// NewSQLiteDB creates a new SQLite-backed API database
func NewSQLiteDB(db *sql.DB) *SQLiteDB {
	return &SQLiteDB{db: db}
}

// Function operations

func (db *SQLiteDB) CreateFunction(ctx context.Context, fn Function) (Function, error) {
	fn.CreatedAt = time.Now().Unix()
	fn.UpdatedAt = fn.CreatedAt

	if fn.EnvVars == nil {
		fn.EnvVars = make(map[string]string)
	}

	query := `INSERT INTO functions (id, name, description, disabled, created_at, updated_at)
	          VALUES (?, ?, ?, ?, ?, ?)`

	_, err := db.db.ExecContext(ctx, query, fn.ID, fn.Name, fn.Description, fn.Disabled, fn.CreatedAt, fn.UpdatedAt)
	if err != nil {
		return Function{}, fmt.Errorf("failed to insert function: %w", err)
	}

	return fn, nil
}

func (db *SQLiteDB) GetFunction(ctx context.Context, id string) (Function, error) {
	query := `SELECT id, name, description, disabled, retention_days, created_at, updated_at
	          FROM functions WHERE id = ?`

	var fn Function
	var description sql.NullString
	var retentionDays sql.NullInt64

	err := db.db.QueryRowContext(ctx, query, id).Scan(
		&fn.ID, &fn.Name, &description, &fn.Disabled, &retentionDays, &fn.CreatedAt, &fn.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Function{}, ErrFunctionNotFound
	}
	if err != nil {
		return Function{}, fmt.Errorf("failed to query function: %w", err)
	}

	if description.Valid {
		fn.Description = &description.String
	}
	if retentionDays.Valid {
		days := int(retentionDays.Int64)
		fn.RetentionDays = &days
	}

	fn.EnvVars = make(map[string]string)

	return fn, nil
}

func (db *SQLiteDB) ListFunctions(ctx context.Context, params PaginationParams) ([]FunctionWithActiveVersion, int64, error) {
	// Get total count
	var total int64
	err := db.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM functions`).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count functions: %w", err)
	}

	// Normalize pagination parameters
	params = params.Normalize()

	query := `SELECT
		f.id, f.name, f.description, f.disabled, f.retention_days, f.created_at, f.updated_at,
		fv.id, fv.version, fv.code, fv.created_at, fv.created_by
	FROM functions f
	LEFT JOIN function_versions fv ON f.id = fv.function_id AND fv.is_active = 1
	ORDER BY f.created_at DESC
	LIMIT ? OFFSET ?`

	rows, err := db.db.QueryContext(ctx, query, params.Limit, params.Offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query functions: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var functions []FunctionWithActiveVersion
	for rows.Next() {
		var fn FunctionWithActiveVersion
		var description sql.NullString
		var retentionDays sql.NullInt64
		var versionID, versionCode sql.NullString
		var versionNum sql.NullInt64
		var versionCreatedAt sql.NullInt64
		var versionCreatedBy sql.NullString

		if err := rows.Scan(
			&fn.ID, &fn.Name, &description, &fn.Disabled, &retentionDays, &fn.CreatedAt, &fn.UpdatedAt,
			&versionID, &versionNum, &versionCode, &versionCreatedAt, &versionCreatedBy,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan function: %w", err)
		}

		if description.Valid {
			fn.Description = &description.String
		}
		if retentionDays.Valid {
			days := int(retentionDays.Int64)
			fn.RetentionDays = &days
		}

		fn.EnvVars = make(map[string]string)

		// Set active version if it exists
		if versionID.Valid {
			fn.ActiveVersion = FunctionVersion{
				ID:         versionID.String,
				FunctionID: fn.ID,
				Version:    int(versionNum.Int64),
				Code:       versionCode.String,
				CreatedAt:  versionCreatedAt.Int64,
				IsActive:   true,
			}
			if versionCreatedBy.Valid {
				fn.ActiveVersion.CreatedBy = &versionCreatedBy.String
			}
		}

		functions = append(functions, fn)
	}

	return functions, total, rows.Err()
}

func (db *SQLiteDB) UpdateFunction(ctx context.Context, id string, updates UpdateFunctionRequest) error {
	tx, err := db.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Check if function exists
	var exists bool
	err = tx.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM functions WHERE id = ?)", id).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check function existence: %w", err)
	}
	if !exists {
		return ErrFunctionNotFound
	}

	if updates.Name != nil {
		_, err = tx.ExecContext(ctx, "UPDATE functions SET name = ?, updated_at = ? WHERE id = ?",
			*updates.Name, time.Now().Unix(), id)
		if err != nil {
			return fmt.Errorf("failed to update name: %w", err)
		}
	}

	if updates.Description != nil {
		_, err = tx.ExecContext(ctx, "UPDATE functions SET description = ?, updated_at = ? WHERE id = ?",
			updates.Description, time.Now().Unix(), id)
		if err != nil {
			return fmt.Errorf("failed to update description: %w", err)
		}
	}

	if updates.Disabled != nil {
		_, err = tx.ExecContext(ctx, "UPDATE functions SET disabled = ?, updated_at = ? WHERE id = ?",
			*updates.Disabled, time.Now().Unix(), id)
		if err != nil {
			return fmt.Errorf("failed to update disabled status: %w", err)
		}
	}

	if updates.RetentionDays != nil {
		_, err = tx.ExecContext(ctx, "UPDATE functions SET retention_days = ?, updated_at = ? WHERE id = ?",
			updates.RetentionDays, time.Now().Unix(), id)
		if err != nil {
			return fmt.Errorf("failed to update retention days: %w", err)
		}
	}

	return tx.Commit()
}

func (db *SQLiteDB) DeleteFunction(ctx context.Context, id string) error {
	result, err := db.db.ExecContext(ctx, "DELETE FROM functions WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete function: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return ErrFunctionNotFound
	}

	return nil
}

// Version operations

func (db *SQLiteDB) CreateVersion(ctx context.Context, functionID string, code string, createdBy *string) (FunctionVersion, error) {
	tx, err := db.db.BeginTx(ctx, nil)
	if err != nil {
		return FunctionVersion{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Check if function exists
	var exists bool
	err = tx.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM functions WHERE id = ?)", functionID).Scan(&exists)
	if err != nil {
		return FunctionVersion{}, fmt.Errorf("failed to check function existence: %w", err)
	}
	if !exists {
		return FunctionVersion{}, ErrFunctionNotFound
	}

	// Get next version number
	var versionNum int
	err = tx.QueryRowContext(ctx,
		"SELECT COALESCE(MAX(version), 0) + 1 FROM function_versions WHERE function_id = ?",
		functionID).Scan(&versionNum)
	if err != nil {
		return FunctionVersion{}, fmt.Errorf("failed to get next version: %w", err)
	}

	// Deactivate all previous versions
	_, err = tx.ExecContext(ctx, "UPDATE function_versions SET is_active = 0 WHERE function_id = ?", functionID)
	if err != nil {
		return FunctionVersion{}, fmt.Errorf("failed to deactivate versions: %w", err)
	}

	version := FunctionVersion{
		ID:         fmt.Sprintf("ver_%s_v%d", functionID, versionNum),
		FunctionID: functionID,
		Version:    versionNum,
		Code:       code,
		CreatedAt:  time.Now().Unix(),
		CreatedBy:  createdBy,
		IsActive:   true,
	}

	query := `INSERT INTO function_versions (id, function_id, version, code, created_at, created_by, is_active)
	          VALUES (?, ?, ?, ?, ?, ?, ?)`

	_, err = tx.ExecContext(ctx, query, version.ID, version.FunctionID, version.Version,
		version.Code, version.CreatedAt, version.CreatedBy, 1)
	if err != nil {
		return FunctionVersion{}, fmt.Errorf("failed to insert version: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return FunctionVersion{}, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return version, nil
}

func (db *SQLiteDB) GetVersion(ctx context.Context, functionID string, version int) (FunctionVersion, error) {
	query := `SELECT id, function_id, version, code, created_at, created_by, is_active
	          FROM function_versions WHERE function_id = ? AND version = ?`

	var v FunctionVersion
	var createdBy sql.NullString

	err := db.db.QueryRowContext(ctx, query, functionID, version).Scan(
		&v.ID, &v.FunctionID, &v.Version, &v.Code, &v.CreatedAt, &createdBy, &v.IsActive,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return FunctionVersion{}, ErrVersionNotFound
	}
	if err != nil {
		return FunctionVersion{}, fmt.Errorf("failed to query version: %w", err)
	}

	if createdBy.Valid {
		v.CreatedBy = &createdBy.String
	}

	return v, nil
}

func (db *SQLiteDB) GetVersionByID(ctx context.Context, versionID string) (FunctionVersion, error) {
	query := `SELECT id, function_id, version, code, created_at, created_by, is_active
	          FROM function_versions WHERE id = ?`

	var v FunctionVersion
	var createdBy sql.NullString

	err := db.db.QueryRowContext(ctx, query, versionID).Scan(
		&v.ID, &v.FunctionID, &v.Version, &v.Code, &v.CreatedAt, &createdBy, &v.IsActive,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return FunctionVersion{}, ErrVersionNotFound
	}
	if err != nil {
		return FunctionVersion{}, fmt.Errorf("failed to query version: %w", err)
	}

	if createdBy.Valid {
		v.CreatedBy = &createdBy.String
	}

	return v, nil
}

func (db *SQLiteDB) ListVersions(ctx context.Context, functionID string, params PaginationParams) ([]FunctionVersion, int64, error) {
	// Get total count
	var total int64
	err := db.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM function_versions WHERE function_id = ?`, functionID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count versions: %w", err)
	}

	// Normalize pagination parameters
	params = params.Normalize()

	query := `SELECT id, function_id, version, code, created_at, created_by, is_active
	          FROM function_versions WHERE function_id = ?
	          ORDER BY version DESC
	          LIMIT ? OFFSET ?`

	rows, err := db.db.QueryContext(ctx, query, functionID, params.Limit, params.Offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query versions: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var versions []FunctionVersion
	for rows.Next() {
		var v FunctionVersion
		var createdBy sql.NullString

		if err := rows.Scan(&v.ID, &v.FunctionID, &v.Version, &v.Code, &v.CreatedAt, &createdBy, &v.IsActive); err != nil {
			return nil, 0, fmt.Errorf("failed to scan version: %w", err)
		}

		if createdBy.Valid {
			v.CreatedBy = &createdBy.String
		}

		versions = append(versions, v)
	}

	return versions, total, rows.Err()
}

func (db *SQLiteDB) GetActiveVersion(ctx context.Context, functionID string) (FunctionVersion, error) {
	query := `SELECT id, function_id, version, code, created_at, created_by, is_active
	          FROM function_versions WHERE function_id = ? AND is_active = 1`

	var v FunctionVersion
	var createdBy sql.NullString

	err := db.db.QueryRowContext(ctx, query, functionID).Scan(
		&v.ID, &v.FunctionID, &v.Version, &v.Code, &v.CreatedAt, &createdBy, &v.IsActive,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return FunctionVersion{}, ErrNoActiveVersion
	}
	if err != nil {
		return FunctionVersion{}, fmt.Errorf("failed to query active version: %w", err)
	}

	if createdBy.Valid {
		v.CreatedBy = &createdBy.String
	}

	return v, nil
}

func (db *SQLiteDB) ActivateVersion(ctx context.Context, functionID string, version int) error {
	tx, err := db.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Check if version exists
	var exists bool
	err = tx.QueryRowContext(ctx,
		"SELECT EXISTS(SELECT 1 FROM function_versions WHERE function_id = ? AND version = ?)",
		functionID, version).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check version existence: %w", err)
	}
	if !exists {
		return ErrVersionNotFound
	}

	// Deactivate all versions
	_, err = tx.ExecContext(ctx, "UPDATE function_versions SET is_active = 0 WHERE function_id = ?", functionID)
	if err != nil {
		return fmt.Errorf("failed to deactivate versions: %w", err)
	}

	// Activate the specified version
	_, err = tx.ExecContext(ctx,
		"UPDATE function_versions SET is_active = 1 WHERE function_id = ? AND version = ?",
		functionID, version)
	if err != nil {
		return fmt.Errorf("failed to activate version: %w", err)
	}

	return tx.Commit()
}

// Execution operations

func (db *SQLiteDB) CreateExecution(ctx context.Context, exec Execution) (Execution, error) {
	exec.CreatedAt = time.Now().Unix()

	query := `INSERT INTO executions (id, function_id, function_version_id, status, duration_ms, error_message, event_json, created_at)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := db.db.ExecContext(ctx, query, exec.ID, exec.FunctionID, exec.FunctionVersionID,
		exec.Status, exec.DurationMs, exec.ErrorMessage, exec.EventJSON, exec.CreatedAt)
	if err != nil {
		return Execution{}, fmt.Errorf("failed to insert execution: %w", err)
	}

	return exec, nil
}

func (db *SQLiteDB) GetExecution(ctx context.Context, executionID string) (Execution, error) {
	query := `SELECT id, function_id, function_version_id, status, duration_ms, error_message, event_json, created_at
	          FROM executions WHERE id = ?`

	var exec Execution
	var durationMs sql.NullInt64
	var errorMessage sql.NullString
	var eventJSON sql.NullString

	err := db.db.QueryRowContext(ctx, query, executionID).Scan(
		&exec.ID, &exec.FunctionID, &exec.FunctionVersionID,
		&exec.Status, &durationMs, &errorMessage, &eventJSON, &exec.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Execution{}, ErrExecutionNotFound
	}
	if err != nil {
		return Execution{}, fmt.Errorf("failed to query execution: %w", err)
	}

	if durationMs.Valid {
		exec.DurationMs = &durationMs.Int64
	}
	if errorMessage.Valid {
		exec.ErrorMessage = &errorMessage.String
	}
	if eventJSON.Valid {
		exec.EventJSON = &eventJSON.String
	}

	return exec, nil
}

func (db *SQLiteDB) UpdateExecution(ctx context.Context, executionID string, status ExecutionStatus, durationMs *int64, errorMsg *string) error {
	query := `UPDATE executions SET status = ?, duration_ms = ?, error_message = ? WHERE id = ?`

	result, err := db.db.ExecContext(ctx, query, status, durationMs, errorMsg, executionID)
	if err != nil {
		return fmt.Errorf("failed to update execution: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return ErrExecutionNotFound
	}

	return nil
}

func (db *SQLiteDB) ListExecutions(ctx context.Context, functionID string, params PaginationParams) ([]Execution, int64, error) {
	// Get total count
	var total int64
	err := db.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM executions WHERE function_id = ?`, functionID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count executions: %w", err)
	}

	// Normalize pagination parameters
	params = params.Normalize()

	query := `
		SELECT e.id, e.function_id, e.function_version_id, e.status,
		       e.duration_ms, e.error_message, e.event_json, e.created_at
		FROM executions e
		WHERE e.function_id = ?
		ORDER BY e.created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := db.db.QueryContext(ctx, query, functionID, params.Limit, params.Offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query executions: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var executions []Execution
	for rows.Next() {
		var exec Execution
		var durationMs sql.NullInt64
		var errorMessage sql.NullString
		var eventJSON sql.NullString

		if err := rows.Scan(&exec.ID, &exec.FunctionID, &exec.FunctionVersionID,
			&exec.Status, &durationMs, &errorMessage, &eventJSON, &exec.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan execution: %w", err)
		}

		if durationMs.Valid {
			exec.DurationMs = &durationMs.Int64
		}
		if errorMessage.Valid {
			exec.ErrorMessage = &errorMessage.String
		}
		if eventJSON.Valid {
			exec.EventJSON = &eventJSON.String
		}

		executions = append(executions, exec)
	}

	return executions, total, rows.Err()
}

func (db *SQLiteDB) DeleteOldExecutions(ctx context.Context, beforeTimestamp int64) (int64, error) {
	query := `DELETE FROM executions WHERE created_at < ?`

	result, err := db.db.ExecContext(ctx, query, beforeTimestamp)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old executions: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}

// Health check

func (db *SQLiteDB) Ping(ctx context.Context) error {
	return db.db.PingContext(ctx)
}
