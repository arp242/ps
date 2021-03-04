package ps

import (
	"os"
	"testing"
)

func TestFind(t *testing.T) {
	p, err := Find(os.Getpid())
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if p == nil {
		t.Fatal("should have process")
	}

	if p.Pid() != os.Getpid() {
		t.Fatalf("bad pid: %#v", p.Pid())
	}

	t.Log(p.String())
}

func TestList(t *testing.T) {
	// This test works because there will always be SOME processes running.
	p, err := List()
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if len(p) <= 0 {
		t.Fatal("should have processes")
	}

	found := false
	for _, p1 := range p {
		if p1.Executable() == "go" || p1.Executable() == "go.exe" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("should have Go")
	}

	t.Log(p.String())
}
