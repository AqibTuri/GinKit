package password

import "testing"

func TestHashAndVerify(t *testing.T) {
	h, err := Hash("correct-horse-battery-staple-phrase")
	if err != nil {
		t.Fatal(err)
	}
	if !Verify(h, "correct-horse-battery-staple-phrase") {
		t.Fatal("expected verify ok")
	}
	if Verify(h, "wrong") {
		t.Fatal("expected verify fail")
	}
}
