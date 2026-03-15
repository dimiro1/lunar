package store

import (
	"context"
	"testing"
)

func TestSQLiteDB_CreateAPIToken(t *testing.T) {
	db, sqliteDB := setupTestDB(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()

	token := APIToken{
		ID:        "token_123",
		TokenHash: "abc123hash",
		Name:      "CLI (ABCD1234)",
	}

	created, err := sqliteDB.CreateAPIToken(ctx, token)
	if err != nil {
		t.Fatalf("CreateAPIToken failed: %v", err)
	}

	if created.ID != token.ID {
		t.Errorf("expected ID %s, got %s", token.ID, created.ID)
	}
	if created.Name != token.Name {
		t.Errorf("expected Name %s, got %s", token.Name, created.Name)
	}
	if created.TokenHash != token.TokenHash {
		t.Errorf("expected TokenHash %s, got %s", token.TokenHash, created.TokenHash)
	}
	if created.CreatedAt == 0 {
		t.Error("expected CreatedAt to be set")
	}
	if created.Revoked {
		t.Error("expected Revoked to be false")
	}
	if created.LastUsed != nil {
		t.Error("expected LastUsed to be nil")
	}
}

func TestSQLiteDB_GetAPITokenByHash(t *testing.T) {
	db, sqliteDB := setupTestDB(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()

	token := APIToken{
		ID:        "token_456",
		TokenHash: "hash456",
		Name:      "CLI (EFGH5678)",
	}

	_, err := sqliteDB.CreateAPIToken(ctx, token)
	if err != nil {
		t.Fatalf("CreateAPIToken failed: %v", err)
	}

	// Retrieve by hash
	found, err := sqliteDB.GetAPITokenByHash(ctx, "hash456")
	if err != nil {
		t.Fatalf("GetAPITokenByHash failed: %v", err)
	}

	if found.ID != token.ID {
		t.Errorf("expected ID %s, got %s", token.ID, found.ID)
	}
	if found.Name != token.Name {
		t.Errorf("expected Name %s, got %s", token.Name, found.Name)
	}
}

func TestSQLiteDB_GetAPITokenByHash_NotFound(t *testing.T) {
	db, sqliteDB := setupTestDB(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()

	_, err := sqliteDB.GetAPITokenByHash(ctx, "nonexistent")
	if err != ErrAPITokenNotFound {
		t.Errorf("expected ErrAPITokenNotFound, got %v", err)
	}
}

func TestSQLiteDB_GetAPITokenByHash_Revoked(t *testing.T) {
	db, sqliteDB := setupTestDB(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()

	token := APIToken{
		ID:        "token_revoked",
		TokenHash: "revoked_hash",
		Name:      "CLI (REVOKED)",
	}

	_, err := sqliteDB.CreateAPIToken(ctx, token)
	if err != nil {
		t.Fatalf("CreateAPIToken failed: %v", err)
	}

	// Revoke the token
	err = sqliteDB.RevokeAPIToken(ctx, token.ID)
	if err != nil {
		t.Fatalf("RevokeAPIToken failed: %v", err)
	}

	// Should not find revoked token
	_, err = sqliteDB.GetAPITokenByHash(ctx, "revoked_hash")
	if err != ErrAPITokenNotFound {
		t.Errorf("expected ErrAPITokenNotFound for revoked token, got %v", err)
	}
}

func TestSQLiteDB_ListAPITokens(t *testing.T) {
	db, sqliteDB := setupTestDB(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()

	// Empty list
	tokens, err := sqliteDB.ListAPITokens(ctx)
	if err != nil {
		t.Fatalf("ListAPITokens failed: %v", err)
	}
	if len(tokens) != 0 {
		t.Errorf("expected empty list, got %d tokens", len(tokens))
	}

	// Create two tokens
	_, err = sqliteDB.CreateAPIToken(ctx, APIToken{
		ID:        "token_1",
		TokenHash: "hash_1",
		Name:      "CLI 1",
	})
	if err != nil {
		t.Fatalf("CreateAPIToken failed: %v", err)
	}

	_, err = sqliteDB.CreateAPIToken(ctx, APIToken{
		ID:        "token_2",
		TokenHash: "hash_2",
		Name:      "CLI 2",
	})
	if err != nil {
		t.Fatalf("CreateAPIToken failed: %v", err)
	}

	tokens, err = sqliteDB.ListAPITokens(ctx)
	if err != nil {
		t.Fatalf("ListAPITokens failed: %v", err)
	}
	if len(tokens) != 2 {
		t.Errorf("expected 2 tokens, got %d", len(tokens))
	}
}

func TestSQLiteDB_RevokeAPIToken(t *testing.T) {
	db, sqliteDB := setupTestDB(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()

	token := APIToken{
		ID:        "token_to_revoke",
		TokenHash: "hash_to_revoke",
		Name:      "CLI (REVOKE ME)",
	}

	_, err := sqliteDB.CreateAPIToken(ctx, token)
	if err != nil {
		t.Fatalf("CreateAPIToken failed: %v", err)
	}

	err = sqliteDB.RevokeAPIToken(ctx, token.ID)
	if err != nil {
		t.Fatalf("RevokeAPIToken failed: %v", err)
	}

	// Verify it shows as revoked in listing
	tokens, err := sqliteDB.ListAPITokens(ctx)
	if err != nil {
		t.Fatalf("ListAPITokens failed: %v", err)
	}
	if len(tokens) != 1 {
		t.Fatalf("expected 1 token, got %d", len(tokens))
	}
	if !tokens[0].Revoked {
		t.Error("expected token to be revoked")
	}
}

func TestSQLiteDB_RevokeAPIToken_NotFound(t *testing.T) {
	db, sqliteDB := setupTestDB(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()

	err := sqliteDB.RevokeAPIToken(ctx, "nonexistent")
	if err != ErrAPITokenNotFound {
		t.Errorf("expected ErrAPITokenNotFound, got %v", err)
	}
}

func TestSQLiteDB_UpdateAPITokenLastUsed(t *testing.T) {
	db, sqliteDB := setupTestDB(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()

	token := APIToken{
		ID:        "token_last_used",
		TokenHash: "hash_last_used",
		Name:      "CLI (LAST USED)",
	}

	_, err := sqliteDB.CreateAPIToken(ctx, token)
	if err != nil {
		t.Fatalf("CreateAPIToken failed: %v", err)
	}

	var timestamp int64 = 1700000000
	err = sqliteDB.UpdateAPITokenLastUsed(ctx, token.ID, timestamp)
	if err != nil {
		t.Fatalf("UpdateAPITokenLastUsed failed: %v", err)
	}

	// Verify via GetAPITokenByHash
	found, err := sqliteDB.GetAPITokenByHash(ctx, "hash_last_used")
	if err != nil {
		t.Fatalf("GetAPITokenByHash failed: %v", err)
	}
	if found.LastUsed == nil {
		t.Fatal("expected LastUsed to be set")
	}
	if *found.LastUsed != timestamp {
		t.Errorf("expected LastUsed %d, got %d", timestamp, *found.LastUsed)
	}
}
