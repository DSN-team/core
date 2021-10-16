package core

import "testing"

func TestStartDB(t *testing.T) {
	StartDB()
	_, err := db.Exec("SELECT name FROM sqlite_master WHERE type='table' AND name=$0", "Profiles")
	if err != nil {
		t.Error("Profiles does not exist")
	}
	_, err = db.Exec("SELECT name FROM sqlite_master WHERE type='table' AND name=$0", "Users")
	if err != nil {
		t.Error("Users does not exist")
	}
}
