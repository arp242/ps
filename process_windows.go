// +build windows

package ps

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

// WindowsProcess is an implementation of Process for Windows.
type WindowsProcess struct {
	pid     int
	ppid    int
	exe     string
	cmdline []string
}

func (p WindowsProcess) String() string {
	return fmt.Sprintf("pid: %d; ppid: %d; state: %c; cmdline: %s",
		p.pid, p.ppid, p.exe, p.cmdline)
}
func (p *WindowsProcess) Pid() int              { return p.pid }
func (p *WindowsProcess) ParentPid() int        { return p.ppid }
func (p *WindowsProcess) Executable() string    { return p.exe }
func (p *WindowsProcess) Commandline() []string { return p.cmdline }

// Windows API functions
var (
	modKernel32                  = syscall.NewLazyDLL("kernel32.dll")
	procCloseHandle              = modKernel32.NewProc("CloseHandle")
	procCreateToolhelp32Snapshot = modKernel32.NewProc("CreateToolhelp32Snapshot")
	procProcess32First           = modKernel32.NewProc("Process32FirstW")
	procProcess32Next            = modKernel32.NewProc("Process32NextW")
)

// Some constants from the Windows API
const (
	ERROR_NO_MORE_FILES = 0x12
	MAX_PATH            = 260
)

// PROCESSENTRY32 is the Windows API structure that contains a process's
// information.
//
// https://docs.microsoft.com/en-us/windows/win32/api/tlhelp32/ns-tlhelp32-processentry32
type PROCESSENTRY32 struct {
	Size              uint32           // Size of the structure, in bytes
	CntUsage          uint32           // Unused; always 0.
	ProcessID         uint32           // The process identifier.
	DefaultHeapID     uintptr          // Unused; always 0.
	ModuleID          uint32           // Unused; always 0.
	CntThreads        uint32           // Number of execution threads started by the process.
	ParentProcessID   uint32           // The identifier of the process that created this process (its parent process).
	PriorityClassBase int32            // The base priority of any threads created by this process.
	Flags             uint32           // Unused: always 0.
	ExeFile           [MAX_PATH]uint16 // The name of the executable file for the process.
}

func newWindowsProcess(e *PROCESSENTRY32) *WindowsProcess {
	// Find when the string ends for decoding
	end := 0
	for {
		if e.ExeFile[end] == 0 {
			break
		}
		end++
	}

	// To retrieve the full path to the executable file, call the Module32First
	// function and check the szExePath member of the MODULEENTRY32 structure
	// that is returned. However, if the calling process is a 32-bit process,
	// you must call the QueryFullProcessImageName function to retrieve the full
	// path of the executable file for a 64-bit process.

	// TODO: get full executable name.
	// TODO: get cmdline.
	return &WindowsProcess{
		pid:  int(e.ProcessID),
		ppid: int(e.ParentProcessID),
		exe:  syscall.UTF16ToString(e.ExeFile[:end]),
	}
}

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
	handle, _, _ := procCreateToolhelp32Snapshot.Call(0x00000002, 0)
	if handle < 0 {
		return nil, syscall.GetLastError()
	}
	defer procCloseHandle.Call(handle)

	var entry PROCESSENTRY32
	entry.Size = uint32(unsafe.Sizeof(entry))
	ret, _, _ := procProcess32First.Call(handle, uintptr(unsafe.Pointer(&entry)))
	if ret == 0 {
		return nil, fmt.Errorf("could not get process list")
	}

	procs := make(Processes, 0, 64)
	for {
		procs = append(procs, newWindowsProcess(&entry))

		ret, _, _ := procProcess32Next.Call(handle, uintptr(unsafe.Pointer(&entry)))
		if ret == 0 {
			break
		}
	}
	return procs, nil
}
