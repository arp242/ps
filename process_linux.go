// +build linux

package ps

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// ProcFS is the path to the procfs filesystem on Linux.
var ProcFS = "/proc"

// UnixProcess is an implementation of Process that contains Unix-specific
// fields and information.
type UnixProcess struct {
	pid, ppid int
	exe       string
	cmdline   []string
	state     rune
}

func (p UnixProcess) String() string {
	return fmt.Sprintf("pid: %d; ppid: %d; state: %c; exe: %s; cmdline: %s",
		p.pid, p.ppid, p.state, p.exe, p.cmdline)
}
func (p *UnixProcess) Pid() int              { return p.pid }
func (p *UnixProcess) ParentPid() int        { return p.ppid }
func (p *UnixProcess) Executable() string    { return p.exe }
func (p *UnixProcess) Commandline() []string { return p.cmdline }

func findProcess(pid int) (Process, error) {
	_, err := os.Stat(fmt.Sprintf("%s/%d", ProcFS, pid))
	if err != nil {
		return nil, err
	}
	return newUnixProcess(pid)
}

func processes() (Processes, error) {
	d, err := os.Open(ProcFS)
	if err != nil {
		return nil, err
	}
	defer d.Close()

	procs := make(Processes, 0, 64)
	for {
		names, err := d.Readdirnames(10)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		for _, name := range names {
			// We only care if the name starts with a numeric
			if name[0] < '0' || name[0] > '9' {
				continue
			}

			// From this point forward, any errors we just ignore, because
			// it might simply be that the process doesn't exist anymore.
			pid, err := strconv.ParseInt(name, 10, 0)
			if err != nil {
				continue
			}

			p, err := newUnixProcess(int(pid))
			if err != nil {
				continue
			}

			procs = append(procs, p)
		}
	}
	return procs, nil
}

func newUnixProcess(pid int) (*UnixProcess, error) {
	p := UnixProcess{pid: pid}
	proc := fmt.Sprintf("/%s/%d", ProcFS, pid)

	d, err := ioutil.ReadFile(proc + "/status")
	if err != nil {
		return nil, err
	}
	for _, line := range bytes.Split(d, []byte("\n")) {
		s := bytes.SplitN(line, []byte(":"), 2)
		if len(s) < 2 {
			continue
		}
		s[1] = bytes.TrimSpace(s[1])
		switch string(s[0]) { // TODO: also list other fields.
		case "Name":
			p.exe = string(s[1])
		case "State":
			if len(s[1]) > 0 {
				p.state = rune(s[1][0])
			}
		case "PPid":
			p.ppid, _ = strconv.Atoi(string(s[1]))
		}
	}

	d, err = ioutil.ReadFile(proc + "/cmdline")
	if err == nil {
		p.cmdline = strings.Split(string(bytes.TrimRight(d, "\x00")), "\x00")
	}

	// The enry in the stat and status files are limited to 16 characters, so
	// try to get the full path from the exe symlink.
	exe, err := os.Readlink(proc + "/exe")
	if err == nil {
		if !filepath.IsAbs(exe) {
			abs, err := filepath.Abs(exe)
			if err == nil {
				exe = abs
			}
		}
		p.exe = exe
	} else if len(p.cmdline) > 0 && len(p.cmdline[0]) > 0 && p.cmdline[0][0] == '/' {
		p.exe = p.cmdline[0]
	}

	return &p, nil
}
