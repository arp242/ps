// +build darwin

package ps

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

type DarwinProcess struct {
	pid    int
	ppid   int
	binary string
}

func (p DarwinProcess) String() string {
	return fmt.Sprintf("pid: %d; ppid: %d; exe: %s", p.Pid(), p.ParentPid(), p.Executable())
}
func (p *DarwinProcess) Pid() int           { return p.pid }
func (p *DarwinProcess) ParentPid() int     { return p.ppid }
func (p *DarwinProcess) Executable() string { return p.binary }

func findProcess(pid int) (Process, error) {
	ps, err := processes()
	if err != nil {
		return nil, err
	}

	for _, p := range ps {
		if p.Pid() == pid {
			return p, nil
		}
	}
	return nil, os.ErrNotExist
}

func processes() (Processes, error) {
	buf, err := darwinSyscall()
	if err != nil {
		return nil, err
	}

	procs := make(Processes, 0, 64)
	k := 0
	for i := _KINFO_STRUCT_SIZE; i < buf.Len(); i += _KINFO_STRUCT_SIZE {
		var p kinfoProc
		err = binary.Read(bytes.NewBuffer(buf.Bytes()[k:i]), binary.LittleEndian, &p)
		if err != nil {
			return nil, err
		}
		k = i

		procs = append(procs, &DarwinProcess{
			pid:    int(p.Pid),
			ppid:   int(p.PPid),
			binary: darwinCstring(p.Comm),
		})
	}
	return procs, nil
}

func darwinCstring(s [16]byte) string {
	i := 0
	for _, b := range s {
		if b == 0 {
			break
		}
		i++
	}
	return string(s[:i])
}

func darwinSyscall() (*bytes.Buffer, error) {
	mib := [4]int32{_CTRL_KERN, _KERN_PROC, _KERN_PROC_ALL, 0}
	size := uintptr(0)

	_, _, errno := syscall.Syscall6(
		syscall.SYS___SYSCTL,
		uintptr(unsafe.Pointer(&mib[0])),
		4, 0,
		uintptr(unsafe.Pointer(&size)),
		0, 0)
	if errno != 0 {
		return nil, errno
	}

	bs := make([]byte, size)
	_, _, errno = syscall.Syscall6(
		syscall.SYS___SYSCTL,
		uintptr(unsafe.Pointer(&mib[0])),
		4,
		uintptr(unsafe.Pointer(&bs[0])),
		uintptr(unsafe.Pointer(&size)),
		0, 0)
	if errno != 0 {
		return nil, errno
	}

	return bytes.NewBuffer(bs[0:size]), nil
}

const (
	_CTRL_KERN         = 1
	_KERN_PROC         = 14
	_KERN_PROC_ALL     = 0
	_KINFO_STRUCT_SIZE = 648
)

// https://github.com/apple/darwin-xnu/blob/8f02f2a044b9bb1ad951987ef5bab20ec9486310/bsd/sys/proc.h#L98
// https://github.com/apple/darwin-xnu/blob/8f02f2a044b9bb1ad951987ef5bab20ec9486310/bsd/sys/sysctl.h#L757
type kinfoProc struct {
	_    [40]byte
	Pid  int32
	_    [199]byte
	Comm [16]byte
	_    [301]byte
	PPid int32
	_    [84]byte
}
