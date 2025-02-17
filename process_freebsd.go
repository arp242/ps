// +build freebsd

package ps

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"syscall"
	"unsafe"
)

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

// https://github.com/freebsd/freebsd-src/blob/147eea3/sys/sys/sysctl.h
const (
	CTL_KERN           = 1  // "high kernel": proc, limits
	KERN_PROC          = 14 // struct: process entries
	KERN_PROC_PID      = 1  // by process id
	KERN_PROC_PROC     = 8  // only return procs
	KERN_PROC_PATHNAME = 12 // path to executable
)

// https://github.com/freebsd/freebsd-src/blob/25c6318/sys/sys/user.h#L121
type KinfoProc struct {
	Ki_structsize   int32
	Ki_layout       int32
	Ki_args         int64 // struct	pargs *ki_args;		/* address of command arguments */
	Ki_paddr        int64
	Ki_addr         int64
	Ki_tracep       int64
	Ki_textvp       int64 // struct	vnode *ki_textvp;	/* pointer to executable file */
	Ki_fd           int64
	Ki_vmspace      int64
	Ki_wchan        int64
	Ki_pid          int32
	Ki_ppid         int32
	Ki_pgid         int32
	Ki_tpgid        int32
	Ki_sid          int32
	Ki_tsid         int32
	Ki_jobc         [2]byte
	Ki_spare_short1 [2]byte
	Ki_tdev         int32
	Ki_siglist      [16]byte
	Ki_sigmask      [16]byte
	Ki_sigignore    [16]byte
	Ki_sigcatch     [16]byte
	Ki_uid          int32
	Ki_ruid         int32
	Ki_svuid        int32
	Ki_rgid         int32
	Ki_svgid        int32
	Ki_ngroups      [2]byte
	Ki_spare_short2 [2]byte
	Ki_groups       [64]byte
	Ki_size         int64
	Ki_rssize       int64
	Ki_swrss        int64
	Ki_tsize        int64
	Ki_dsize        int64
	Ki_ssize        int64
	Ki_xstat        [2]byte
	Ki_acflag       [2]byte
	Ki_pctcpu       int32
	Ki_estcpu       int32
	Ki_slptime      int32
	Ki_swtime       int32
	Ki_cow          int32
	Ki_runtime      int64
	Ki_start        [16]byte
	Ki_childtime    [16]byte
	Ki_flag         int64
	Ki_kiflag       int64
	Ki_traceflag    int32
	Ki_stat         [1]byte
	Ki_nice         [1]byte
	Ki_lock         [1]byte
	Ki_rqindex      [1]byte
	Ki_oncpu        [1]byte
	Ki_lastcpu      [1]byte
	Ki_ocomm        [17]byte
	Ki_wmesg        [9]byte
	Ki_login        [18]byte
	Ki_lockname     [9]byte
	Ki_comm         [20]byte // command name
	Ki_emul         [17]byte
	Ki_sparestrings [68]byte
	Ki_spareints    [36]byte
	Ki_cr_flags     int32
	Ki_jid          int32
	Ki_numthreads   int32
	Ki_tid          int32
	Ki_pri          int32
	Ki_rusage       [144]byte
	Ki_rusage_ch    [144]byte
	Ki_pcb          int64
	Ki_kstack       int64
	Ki_udata        int64
	Ki_tdaddr       int64
	Ki_spareptrs    [48]byte
	Ki_spareint64s  [96]byte
	Ki_sflag        int64
	Ki_tdflags      int64
}

// TODO: get full executable name.
// TODO: get cmdline.
func setParams(k *KinfoProc, p *UnixProcess) {
	p.ppid = int(k.Ki_ppid)
	for i, b := range k.Ki_comm {
		if b == 0 {
			p.exe = string(k.Ki_comm[:i+1])
			break
		}
	}
}

func findProcess(pid int) (Process, error) {
	_, _, err := callSyscall([]int32{CTL_KERN, KERN_PROC, KERN_PROC_PATHNAME, int32(pid)})
	if err != nil {
		return nil, err
	}
	return newUnixProcess(pid)
}

func processes() (Processes, error) {
	buf, bufLen, err := callSyscall([]int32{CTL_KERN, KERN_PROC, KERN_PROC_PROC, 0})
	if err != nil {
		return nil, err
	}

	var k KinfoProc
	kLen := int(unsafe.Sizeof(k))
	count := int(bufLen / uint64(kLen))

	procs := make(Processes, 0, 64)
	for i := 0; i < count; i++ {
		b := buf[i*kLen : i*kLen+kLen]
		k, err := parseKinfoProc(b)
		if err != nil {
			continue
		}
		p, err := newUnixProcess(int(k.Ki_pid))
		if err != nil {
			continue
		}

		setParams(&k, p)
		procs = append(procs, p)
	}

	return procs, nil
}

func parseKinfoProc(buf []byte) (KinfoProc, error) {
	var k KinfoProc
	err := binary.Read(bytes.NewReader(buf), binary.LittleEndian, &k)
	if err != nil {
		return k, err
	}

	return k, nil
}

func callSyscall(mib []int32) ([]byte, uint64, error) {
	miblen := uint64(len(mib))

	// get required buffer size
	length := uint64(0)
	_, _, err := syscall.RawSyscall6(
		syscall.SYS___SYSCTL,
		uintptr(unsafe.Pointer(&mib[0])),
		uintptr(miblen),
		0,
		uintptr(unsafe.Pointer(&length)),
		0, 0)
	if err != 0 {
		return nil, 0, err
	}
	if length == 0 {
		return nil, 0, errors.New("length == 0")
	}

	// get proc info itself
	buf := make([]byte, length)
	_, _, err = syscall.RawSyscall6(
		syscall.SYS___SYSCTL,
		uintptr(unsafe.Pointer(&mib[0])),
		uintptr(miblen),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&length)),
		0, 0)
	if err != 0 {
		return buf, length, err
	}

	return buf, length, nil
}

func newUnixProcess(pid int) (*UnixProcess, error) {
	buf, length, err := callSyscall([]int32{CTL_KERN, KERN_PROC, KERN_PROC_PID, int32(pid)})
	if err != nil {
		return nil, err
	}
	if length != uint64(unsafe.Sizeof(KinfoProc{})) {
		return nil, errors.New("sysctl call failed: wrong length")
	}

	k, err := parseKinfoProc(buf)
	if err != nil {
		return nil, err
	}

	var p UnixProcess
	setParams(&k, &p)
	return &p, nil
}
