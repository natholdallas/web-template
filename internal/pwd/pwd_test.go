package pwd

import "testing"

func TestHashVerify(t *testing.T) {
	h, err := Hash("123456")
	if err != nil {
		t.Fatal(err)
	}
	if !IsHashed(h) {
		t.Fatalf("expected hashed prefix, got %q", h)
	}
	if !Verify("123456", h) {
		t.Fatal("expected correct password to verify")
	}
	if Verify("wrong", h) {
		t.Fatal("expected wrong password to fail")
	}
	if Verify("123456", "plaintext") {
		t.Fatal("expected plaintext to fail verify")
	}
	if Verify("123456", "") {
		t.Fatal("expected empty to fail verify")
	}
}
