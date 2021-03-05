// +build linux

package ps

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

// ProcFS is the path to the procfs filesystem on Linux.
var ProcFS = "/proc"

// UnixProcess is an implementation of Process that contains Unix-specific
// fields and information.
type UnixProcess struct {
	pid    int
	ppid   int
	state  rune
	pgrp   int
	sid    int
	binary string
}

func (p UnixProcess) String() string {
	return fmt.Sprintf("pid: %d; ppid: %d; exe: %s", p.Pid(), p.ParentPid(), p.Executable())
}
func (p *UnixProcess) Pid() int           { return p.pid }
func (p *UnixProcess) ParentPid() int     { return p.ppid }
func (p *UnixProcess) Executable() string { return p.binary }

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
	d, err := ioutil.ReadFile(fmt.Sprintf("/%s/%d/stat", ProcFS, pid))
	if err != nil {
		return nil, err
	}

	// First, parse out the image name
	data := string(d)
	binStart := strings.IndexRune(data, '(') + 1
	binEnd := strings.IndexRune(data[binStart:], ')')

	p := UnixProcess{
		pid:    pid,
		binary: data[binStart : binStart+binEnd],
	}

	// Move past the image name and start parsing the rest
	data = data[binStart+binEnd+2:]
	_, err = fmt.Sscanf(data,
		"%c %d %d %d",
		&p.state,
		&p.ppid,
		&p.pgrp,
		&p.sid)
	if err != nil {
		return nil, err
	}
	return &p, nil
}
