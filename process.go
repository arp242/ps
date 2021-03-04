// ps lists ystem processes.
//
// NOTE: If you're reading these docs online via GoDocs or some other system,
// you might only see the Unix docs. This project makes heavy use of
// platform-specific implementations. We recommend reading the source if you
// are interested.
package ps

import (
	"fmt"
	"strconv"
	"strings"
)

// Process is a single process.
type Process interface {
	fmt.Stringer
	Pid() int           // Process ID.
	PPid() int          // Parent process ID.
	Executable() string // Executable name running this process, i.e. "go" or "go.exe".
}

type Processes []Process

func (p Processes) String() string {
	// TODO: align
	var b strings.Builder
	b.WriteString(strconv.Itoa(len(p)))
	b.WriteString(" processes:\n")
	for _, pp := range p {
		b.WriteString("    ")
		b.WriteString(pp.String())
		b.WriteRune('\n')
	}
	return b.String()
}

// List all processes.
//
// This is a point-in-time snapshot of when this method was called. Some
// operating systems don't provide snapshot capability of the process table, in
// which case the process table returned might contain ephemeral entities that
// happened to be running when this was called.
func List() (Processes, error) {
	return processes()
}

// Find looks up a single process by pid.
//
// Process and error will be nil if a matching process is not found.
func Find(pid int) (Process, error) {
	return findProcess(pid)
}
