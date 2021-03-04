// +build linux

package ps

import (
	"fmt"
	"io"
	"os"
	"strconv"
)

// UnixProcess is an implementation of Process that contains Unix-specific
// fields and information.
type UnixProcess struct {
	pid   int
	ppid  int
	state rune
	pgrp  int
	sid   int

	binary string
}

func (p UnixProcess) String() string {
	return fmt.Sprintf("pid: %d; ppid: %d; exe: %s", p.Pid(), p.PPid(), p.Executable())
}
func (p *UnixProcess) Pid() int           { return p.pid }
func (p *UnixProcess) PPid() int          { return p.ppid }
func (p *UnixProcess) Executable() string { return p.binary }

func findProcess(pid int) (Process, error) {
	_, err := os.Stat(fmt.Sprintf("/proc/%d", pid))
	if err != nil {
		return nil, err
	}
	return newUnixProcess(pid)
}

func processes() (Processes, error) {
	d, err := os.Open("/proc")
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
	p := &UnixProcess{pid: pid}
	return p, p.Refresh()
}
