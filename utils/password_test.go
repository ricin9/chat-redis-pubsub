package utils

import (
	"testing"
)

func TestSamePassword(t *testing.T) {
	pass := "12345678"
	hash, err := HashPassword(pass)

	if err != nil {
		t.Fatalf("Hashing password failed, value: %s, error: %v", pass, err)
	}

	if ok, err := ComparePassword(hash, pass); !ok || err != nil {
		t.Fatalf("Password comparison failed, value: %s, error: %v", pass, err)
	}

}

func TestDifferentPassword(t *testing.T) {
	pass := "12345678"
	pass2 := "different"
	hash, err := HashPassword(pass)

	if err != nil {
		t.Fatalf("Hashing password failed, value: %s, error: %v", pass, err)
	}

	if ok, _ := ComparePassword(hash, pass2); ok {
		t.Fatalf("Password comparison succeeded when should have failed, %s != %s", pass, pass2)
	}

}
