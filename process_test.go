package ps

import (
	"os"
	"strings"
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
	for _, pp := range p {
		if strings.HasSuffix(pp.Executable(), "ps.test") || strings.HasSuffix(pp.Executable(), "ps.test.exe") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("should have Go\n%s", p)
	}

	t.Log(p.String())
}
