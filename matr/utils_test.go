package matr

import "testing"

func TestShEscapedPercentIsPassedThrough(t *testing.T) {
	out, err := Sh("printf '%%s' foo").Output()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if string(out) != "foo" {
		t.Fatalf("expected foo, got %q", string(out))
	}
}
