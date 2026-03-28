package config

import (
	"strings"
	"testing"
)

func validDatabase() Database {
	return Database{Host: "localhost", Port: 5432, Name: "argus", User: "argus", Password: "secret", SSLMode: "disable"}
}

func TestDatabase_Validate_Valid(t *testing.T) {
	if err := validDatabase().Validate(); err != nil {
		t.Errorf("expected valid, got: %v", err)
	}
}

func TestDatabase_Validate_MissingHost(t *testing.T) {
	c := validDatabase()
	c.Host = ""
	if err := c.Validate(); err == nil {
		t.Error("expected error for missing host")
	}
}

func TestDatabase_Validate_InvalidPort(t *testing.T) {
	for _, port := range []int{0, -1} {
		c := validDatabase()
		c.Port = port
		if err := c.Validate(); err == nil {
			t.Errorf("expected error for port %d", port)
		}
	}
}

func TestDatabase_Validate_MissingName(t *testing.T) {
	c := validDatabase()
	c.Name = ""
	if err := c.Validate(); err == nil {
		t.Error("expected error for missing name")
	}
}

func TestDatabase_Validate_MissingUser(t *testing.T) {
	c := validDatabase()
	c.User = ""
	if err := c.Validate(); err == nil {
		t.Error("expected error for missing user")
	}
}

func TestDatabase_DSN(t *testing.T) {
	c := Database{Host: "db.example.com", Port: 5432, User: "admin", Password: "pw", Name: "mydb", SSLMode: "require"}
	dsn := c.DSN()

	for _, want := range []string{"host=db.example.com", "port=5432", "user=admin", "password=pw", "dbname=mydb", "sslmode=require"} {
		if !strings.Contains(dsn, want) {
			t.Errorf("DSN %q missing %q", dsn, want)
		}
	}
}
