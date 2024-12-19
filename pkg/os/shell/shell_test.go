package shell

import (
	"testing"
)

func TestGetPathEnvString(t *testing.T) {
	tests := []struct {
		name        string
		userShell   string
		path        string
		expectedStr string
	}{
		{"fish shell", "fish", "C:\\Users\\foo\\.crc\\bin\\oc", "contains /C/Users/foo/.crc/bin/oc $fish_user_paths; or set -U fish_user_paths /C/Users/foo/.crc/bin/oc $fish_user_paths"},
		{"powershell shell", "powershell", "C:\\Users\\foo\\oc.exe", "$Env:PATH = \"C:\\Users\\foo\\oc.exe;$Env:PATH\""},
		{"cmd shell", "cmd", "C:\\Users\\foo\\oc.exe", "SET PATH=C:\\Users\\foo\\oc.exe;%PATH%"},
		{"bash with windows path", "bash", "C:\\Users\\foo.exe", "export PATH=\"/C/Users/foo.exe:$PATH\""},
		{"unknown with windows path", "unknown", "C:\\Users\\foo.exe", "export PATH=\"C:\\Users\\foo.exe:$PATH\""},
		{"unknown shell with unix path", "unknown", "/home/foo/.crc/bin/oc", "export PATH=\"/home/foo/.crc/bin/oc:$PATH\""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetPathEnvString(tt.userShell, tt.path)
			if result != tt.expectedStr {
				t.Errorf("GetPathEnvString(%s, %s) = %s; want %s", tt.userShell, tt.path, result, tt.expectedStr)
			}
		})
	}
}

func TestConvertToLinuxStylePath(t *testing.T) {
	tests := []struct {
		name         string
		userShell    string
		path         string
		expectedPath string
	}{
		{"bash on windows, should convert", "bash", "C:\\Users\\foo\\.crc\\bin\\oc", "/C/Users/foo/.crc/bin/oc"},
		{"zsh on windows, should convert", "zsh", "C:\\Users\\foo\\.crc\\bin\\oc", "/C/Users/foo/.crc/bin/oc"},
		{"fish on windows, should convert", "fish", "C:\\Users\\foo\\.crc\\bin\\oc", "/C/Users/foo/.crc/bin/oc"},
		{"powershell on windows, should NOT convert", "powershell", "C:\\Users\\foo\\.crc\\bin\\oc", "C:\\Users\\foo\\.crc\\bin\\oc"},
		{"cmd on windows, should NOT convert", "cmd", "C:\\Users\\foo\\.crc\\bin\\oc", "C:\\Users\\foo\\.crc\\bin\\oc"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertToLinuxStylePath(tt.userShell, tt.path)
			if result != tt.expectedPath {
				t.Errorf("convertToLinuxStylePath(%s, %s) = %s; want %s", tt.userShell, tt.path, result, tt.expectedPath)
			}
		})
	}
}
