package database

import (
	"os"
	"testing"
	"time"
)

func TestGenerateHash(t *testing.T) {
	hash1 := GenerateHash("test", "data")
	hash2 := GenerateHash("test", "data")
	hash3 := GenerateHash("different", "data")

	if hash1 != hash2 {
		t.Error("Same inputs should produce same hash")
	}

	if hash1 == hash3 {
		t.Error("Different inputs should produce different hash")
	}
}

func TestHashLength(t *testing.T) {
	hash := GenerateHash("test")
	// SHA256 produces 64 hex characters
	if len(hash) != 64 {
		t.Errorf("Expected hash length 64, got %d", len(hash))
	}
}

func TestInitDB(t *testing.T) {
	// Create temporary directory for test
	tmpDir := "./test_badger_data"
	defer os.RemoveAll(tmpDir)

	// Note: InitDB uses hardcoded path, this test is simplified
	db, err := InitDB()
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer db.Close()

	if db == nil {
		t.Fatal("InitDB returned nil")
	}
}

func TestSetGet(t *testing.T) {
	db, _ := InitDB()
	defer db.Close()

	type testData struct {
		Name  string
		Value int
		Time  time.Time
	}

	data := testData{
		Name:  "test",
		Value: 42,
		Time:  time.Now(),
	}

	err := db.Set("test_key", data)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	retrieved := testData{}
	err = db.GetJSON("test_key", &retrieved)
	if err != nil {
		t.Fatalf("GetJSON failed: %v", err)
	}

	if retrieved.Name != data.Name {
		t.Errorf("Expected name '%s', got '%s'", data.Name, retrieved.Name)
	}

	if retrieved.Value != data.Value {
		t.Errorf("Expected value %d, got %d", data.Value, retrieved.Value)
	}
}

func TestExists(t *testing.T) {
	db, _ := InitDB()
	defer db.Close()

	db.Set("test_exists", "value")

	exists, err := db.Exists("test_exists")
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}

	if !exists {
		t.Error("Key should exist")
	}

	exists, _ = db.Exists("nonexistent")
	if exists {
		t.Error("Key should not exist")
	}
}

func TestDelete(t *testing.T) {
	db, _ := InitDB()
	defer db.Close()

	db.Set("to_delete", "value")

	err := db.Delete("to_delete")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	exists, _ := db.Exists("to_delete")
	if exists {
		t.Error("Key should be deleted")
	}
}
