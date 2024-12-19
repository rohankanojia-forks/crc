package shell

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"unsafe"
)

var (
	supportedShell = []string{"cmd", "powershell", "bash", "zsh", "fish"}
)

// re-implementation of private function in https://github.com/golang/go/blob/master/src/syscall/syscall_windows.go
func getProcessEntry(pid uint32) (pe *syscall.ProcessEntry32, err error) {
	snapshot, err := syscall.CreateToolhelp32Snapshot(syscall.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = syscall.CloseHandle(syscall.Handle(snapshot))
	}()

	var processEntry syscall.ProcessEntry32
	processEntry.Size = uint32(unsafe.Sizeof(processEntry))
	err = syscall.Process32First(snapshot, &processEntry)
	if err != nil {
		return nil, err
	}

	for {
		if processEntry.ProcessID == pid {
			pe = &processEntry
			return
		}

		err = syscall.Process32Next(snapshot, &processEntry)
		if err != nil {
			return nil, err
		}
	}
}

// getNameAndItsPpid returns the exe file name its parent process id.
func getNameAndItsPpid(pid uint32) (exefile string, parentid uint32, err error) {
	pe, err := getProcessEntry(pid)
	if err != nil {
		return "", 0, err
	}
	
	name := syscall.UTF16ToString(pe.ExeFile[:])
	return name, pe.ParentProcessID, nil
}

func shellType(shell string, defaultShell string) string {
	switch {
	case strings.Contains(strings.ToLower(shell), "powershell"):
		return "powershell"
	case strings.Contains(strings.ToLower(shell), "pwsh"):
		return "powershell"
	case strings.Contains(strings.ToLower(shell), "cmd"):
		return "cmd"
	case strings.Contains(strings.ToLower(shell), "wsl"):
		return detectShellInWindowsSubsystemLinux("bash")
	case filepath.IsAbs(shell) && strings.Contains(strings.ToLower(shell), "bash"):
		return "bash"
	default:
		return defaultShell
	}
}

func detect() (string, error) {
	shell := os.Getenv("SHELL")

	if shell == "" {
		pid := os.Getppid()
		if pid < 0 || pid > math.MaxUint32 {
			return "", fmt.Errorf("integer overflow for pid: %v", pid)
		}
		shell, shellppid, err := getNameAndItsPpid(uint32(pid))
		if err != nil {
			return "cmd", err // defaulting to cmd
		}
		shell = shellType(shell, "")
		if shell == "" {
			shell, _, err := getNameAndItsPpid(shellppid)
			if err != nil {
				return "cmd", err // defaulting to cmd
			}
			return shellType(shell, "cmd"), nil
		}
		return shell, nil
	}

	if os.Getenv("__fish_bin_dir") != "" {
		return "fish", nil
	}

	return shellType(shell, "cmd"), nil
}

func detectShellInWindowsSubsystemLinux(defaultShell string) string {
	cmd := exec.Command("wsl", "-e", "bash", "-c", "ps -ao pid,comm")
	output, err := cmd.Output()
	if err != nil {
		return defaultShell
	}

	return inspectWslProcessForRecentlyUsedShell(string(output))
}

func inspectWslProcessForRecentlyUsedShell(psCommandOutput string) string {
	type ProcessOutput struct {
		processId string
		output    string
	}
	var processOutputs []ProcessOutput
	lines := strings.Split(psCommandOutput, "\n")[1:]
	for _, line := range lines {
		lineParts := strings.Split(strings.TrimSpace(line), " ")
		if len(lineParts) == 2 && (strings.Contains(lineParts[1], "zsh") ||
			strings.Contains(lineParts[1], "bash") ||
			strings.Contains(lineParts[1], "fish")) {
			processOutputs = append(processOutputs, ProcessOutput{
				processId: lineParts[0],
				output:    lineParts[1],
			})
		}
	}
	sort.Slice(processOutputs, func(i, j int) bool {
		return processOutputs[i].processId < processOutputs[j].processId
	})
	return processOutputs[0].output
}

func isWindowsSubsystemLinux() bool {
	procVersionContent, err := os.ReadFile("/proc/version")
	if err != nil {
		return false
	}
	return doesVersionFileContainsWSL(string(procVersionContent))
}

func doesVersionFileContainsWSL(procVersionContent string) bool {
	if strings.Contains(procVersionContent, "Microsoft") ||
		strings.Contains(procVersionContent, "WSL") {
		return true
	}
	return false
}

func convertToWindowsSubsystemLinuxPath(path string) string {
	cmd := exec.Command("wsl", "-e", "bash", "-c", fmt.Sprintf("wslpath -a '%s'", path))
	convertedWslPath, err := cmd.Output()
	if err != nil {
		return path
	}
	return strings.TrimSpace(string(convertedWslPath))
}
