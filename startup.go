package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	lnk "github.com/parsiya/golnk"
	"golang.org/x/sys/windows/registry"
)

// https://github.com/lsdev/litestep-/blob/bbc1182d8abcd660e1c91c2207aa27b940898c9b/litestep/StartupRunner.cpp#L143
// https://github.com/cairoshell/ManagedShell/blob/5e7e0ed524c6d196032161eec11888c59c6175b4/src/ManagedShell.Common/SupportingClasses/StartupRunner.cs#L33
func startup() {
	startFromRegistry(registry.LOCAL_MACHINE, `Software\Microsoft\Windows\CurrentVersion\Run`, false)
	startFromRegistry(registry.LOCAL_MACHINE, `Software\Microsoft\Windows\CurrentVersion\RunOnce`, true)
	startFromRegistry(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Run`, false)
	startFromRegistry(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\RunOnce`, true)

	startFromFolder(getKnownFolderPath(FOLDERIDs["FOLDERID_CommonStartup"]))
	startFromFolder(getKnownFolderPath(FOLDERIDs["FOLDERID_Startup"]))
}

func startFromFolder(path string) {
	files, err := os.ReadDir(path)
	if err != nil {
		log.Println(err)
		return
	}
	for _, f := range files {
		go RunProgram(filepath.Join(path, f.Name()))
	}
}

func startFromRegistry(loc registry.Key, path string, deleteValue bool) {
	k, err := registry.OpenKey(loc, path, registry.SET_VALUE|registry.QUERY_VALUE)
	if err != nil {
		log.Println(err)
		return
	}
	defer k.Close()

	params, err := k.ReadValueNames(0)
	if err != nil {
		return
	}
	for _, param := range params {
		s, _, err := k.GetStringValue(param)
		if err != nil {
			continue
		}

		if deleteValue {
			k.DeleteValue(param)
		}

		// https://stackoverflow.com/a/12076082
		go exec.Command("rundll32.exe", "url.dll,FileProtocolHandler", s).Start()
	}
}

func RunProgram(file string) {
	switch strings.ToLower(filepath.Ext(file)) {
	case ".lnk":
		Lnk, err := lnk.File(file)
		if err != nil {
			return
		}
		exec.Command(Lnk.LinkInfo.LocalBasePath, Lnk.StringData.CommandLineArguments).Start()
	case ".ini": // desktop.ini
	default:
		exec.Command(file).Start()
	}
}
