package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"syscall"

	"github.com/leaanthony/winc"
	"github.com/leaanthony/winc/w32"
	lnk "github.com/parsiya/golnk"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

var FolderIconhBmp *winc.Bitmap

func init() {
	ic, err := winc.ExtractIcon(ResolveVariables("%SystemRoot%\\system32\\imageres.dll"), -4)
	if err != nil {
		log.Println(err)
	}
	hBmp, err := winc.NewBitmapFromIconForDPI(ic, w32.Size{Width: 16, Height: 16}, 96)
	if err != nil {
		log.Println(err)
	}
	FolderIconhBmp = hBmp
}

var ContextmenuFunc = func(arg *winc.Event) {
	data := arg.Data.(*winc.MouseContextData)
	createCommandCall(data.Item.Command.Workdir, data.Item.Command.Filename)
}

var TaskmenuFunc = func(arg *winc.Event) {
	data := arg.Data.(*winc.MouseContextData)
	createCommandCall(data.Item.Command.Workdir, data.Item.Command.Filename)
}

func (s *shell) ContextMenu() *winc.MenuItem {
	contextmenu := winc.NewContextMenu()

	for i := 0; i < len(config.Contextmenu); i++ {
		menu := config.Contextmenu[i]
		switch menu.Name {
		case "separator", "_", ".":
			contextmenu.AddSeparator()
			continue
		case "CLSID_Desktop":
			val := getDefaultValue(registry.CLASSES_ROOT, `CLSID\{20D04FE0-3AEA-1069-A2D8-08002B30309D}\DefaultIcon`)
			commaSep := strings.Index(val, ",")
			menu.Name = getMUINames(registry.CLASSES_ROOT, `CLSID\{20D04FE0-3AEA-1069-A2D8-08002B30309D}`, "Desktop")
			menu.Icon.Filename = ResolveVariables(val[:commaSep])
			menu.Icon.Index = StringToInt(val[commaSep+1:])
		case "CLSID_Run":
			val := getDefaultValue(registry.CLASSES_ROOT, `CLSID\{2559a1f3-21d7-11d4-bdaf-00c04f60b9f0}\DefaultIcon`)
			commaSep := strings.Index(val, ",")
			menu.Name = getMUINames(registry.CLASSES_ROOT, `CLSID\{2559a1f3-21d7-11d4-bdaf-00c04f60b9f0}`, "Run...")
			menu.Icon.Filename = ResolveVariables(val[:commaSep])
			menu.Icon.Index = StringToInt(val[commaSep+1:])
		case "CLSID_RecycleBin":
			val := getDefaultValue(registry.CLASSES_ROOT, `CLSID\{645FF040-5081-101B-9F08-00AA002F954E}\DefaultIcon`)
			commaSep := strings.Index(val, ",")
			menu.Name = getMUINames(registry.CLASSES_ROOT, `CLSID\{645FF040-5081-101B-9F08-00AA002F954E}`, "Recycle Bin")
			menu.Icon.Filename = ResolveVariables(val[:commaSep])
			menu.Icon.Index = StringToInt(val[commaSep+1:])
		}

		if len(menu.Path) != 0 {
			AddSubMenu(contextmenu, menu.Name, menu.Path)
		} else {
			AddItem(contextmenu, &menu)
		}
	}

	if config.Desktop.Contextmenu.AddDebugEntry {
		contextmenu.AddSeparator()

		exitContextMenu := contextmenu.AddItem("Exit GoShell", winc.NoShortcut)
		exitContextMenu.OnClick().Bind(func(_ *winc.Event) {
			winc.Exit()
		})
	}

	return contextmenu
}

func (s *shell) MiddleMenu() *winc.MenuItem {
	middleMenu := winc.NewContextMenu()

	if s.TaskbarWindow == nil {
		return middleMenu
	}

	for _, btn := range s.TaskbarWindow.tl.PushButtonList {
		var m *winc.MenuItem
		windowTitle := WindowTitle(btn.hWnd)
		if ico := GetAppIcon(btn.hWnd); ico != nil {
			hBmp, err := winc.NewBitmapFromIconForDPI(ico, w32.Size{Width: 16, Height: 16}, 96)
			if err != nil {
				// log.Println("NewBitmapFromIconForDPI, Error:", err)
				m = middleMenu.AddItem(windowTitle, winc.NoShortcut)
			} else {
				m = middleMenu.AddItemWithBitmap(windowTitle, winc.NoShortcut, hBmp)
			}
		} else {
			m = middleMenu.AddItem(windowTitle, winc.NoShortcut)
		}
		m.Command.Hwnd = btn.hWnd
		m.OnMClick().Bind(func(arg *winc.Event) {
			hWnd := arg.Data.(*winc.MouseContextData).Item.Command.Hwnd
			if !w32.IsWindowVisible(hWnd) {
				w32.ShowWindow(w32.HWND(hWnd), w32.SW_SHOW)
			}
			if w32.IsIconic(hWnd) {
				w32.ShowWindow(w32.HWND(hWnd), w32.SW_RESTORE)
			}
			w32.SetForegroundWindow(w32.HWND(w32.GetLastActivePopup(hWnd)))
		})
	}

	return middleMenu
}

func (s *shell) Refresh() {
	s.mainWindow.SetContextMenu(s.ContextMenu())
}

func createCommandCall(dir, file string) {
	var prog string
	var args []string
	switch filepath.Ext(file) {
	case ".ps1": // https://docs.microsoft.com/de-de/powershell/module/microsoft.powershell.core/about/about_powershell_exe
		prog = "PowerShell.exe"
		args = []string{"-NoLogo", "-NoProfile", "-ExecutionPolicy Bypass", "-File", filepath.Join(dir, file)}

	default: // https://stackoverflow.com/a/12076082
		prog = "rundll32.exe"
		args = []string{"url.dll,FileProtocolHandler", filepath.Join(dir, file)}
	}
	exec.Command(prog, args...).Start()
}

func shellExecute(argv0 string, args []string, hidden bool) syscall.Handle {
	var lpFile *uint16
	if len(args) != 0 {
		lpFile, _ = syscall.UTF16PtrFromString(" " + strings.Join(args, " "))
	} else {
		lpFile, _ = syscall.UTF16PtrFromString(argv0)
	}

	var showCmd int32
	if hidden {
		showCmd = windows.SW_HIDE
	} else {
		showCmd = windows.SW_SHOWDEFAULT
	}

	err := windows.ShellExecute(
		0,
		nil, // windows.StringToUTF16Ptr("open"), // windows.StringToUTF16Ptr("runas"),
		lpFile,
		nil,
		nil,
		showCmd,
	)
	if err != nil {
		log.Printf("Return: %d, %v\n", err, err)
	}
	return syscall.Handle(0)
}

// https://docs.microsoft.com/de-de/windows/win32/procthread/creating-processes
// https://docs.microsoft.com/en-us/windows/win32/procthread/process-creation-flags
func createProcess(argv0 string, args []string, hidden bool) syscall.ProcessInformation {
	var lpCommandLine *uint16
	if len(args) != 0 {
		lpCommandLine, _ = syscall.UTF16PtrFromString(" " + strings.Join(args, " "))
	}
	lpApplicationName, _ := syscall.UTF16PtrFromString(argv0)

	var startupInfo syscall.StartupInfo
	startupInfo.Flags = w32.SW_SHOWNORMAL
	var processInformation syscall.ProcessInformation

	var creationFlags uint32
	if hidden {
		creationFlags = windows.NORMAL_PRIORITY_CLASS | windows.CREATE_NO_WINDOW | windows.CREATE_NEW_PROCESS_GROUP
	} else {
		creationFlags = windows.NORMAL_PRIORITY_CLASS | windows.CREATE_NEW_CONSOLE | windows.CREATE_NEW_PROCESS_GROUP
	}

	err := syscall.CreateProcess(
		lpApplicationName,   // No module name (use command line)
		lpCommandLine,       // Command line
		nil,                 // Process handle not inheritable
		nil,                 // Thread handle not inheritable
		false,               // Set handle inheritance to FALSE
		creationFlags,       // creation flags
		nil,                 // Use parent's environment block
		nil,                 // Use parent's starting directory
		&startupInfo,        // Pointer to STARTUPINFO structure
		&processInformation) // Pointer to PROCESS_INFORMATION structure

	if err != nil {
		log.Printf("Return: %d, %v\n", err, err)
	}

	return processInformation
	// WaitForSingleObject(processInfo.hProcess, INFINITE);
}

// Start starts the specified command but does not wait for it to complete.
func openProcess(file string, args []string, hidden bool) int {
	cmd := exec.Command(file, args...)
	if hidden {
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	}
	err := cmd.Start()
	if err != nil {
		log.Println(err)
		w32.MessageBox(0, "openProcess", err.Error(), w32.MB_ICONERROR)
	}
	return cmd.Process.Pid
}

// https://docs.microsoft.com/en-us/windows/win32/api/shlobj_core/nf-shlobj_core-shgetfolderpatha
func getKnownFolderPath(guid *windows.KNOWNFOLDERID) string {
	flags := []uint32{windows.KF_FLAG_DEFAULT, windows.KF_FLAG_DEFAULT_PATH}
	for _, flag := range flags {
		p, _ := windows.KnownFolderPath(guid, flag|windows.KF_FLAG_DONT_VERIFY)
		if p != "" {
			return p
		}
	}
	return ""
}

type DirValues struct {
	Root  string
	Value string
}

func osReadDir(dirname string) (files, folders []string) {
	fileinfos, err := os.ReadDir(dirname)
	if err != nil {
		log.Println(err)
		w32.MessageBox(0, err.Error(), "ReadDir Error", w32.MB_ICONERROR)
	}

	for _, file := range fileinfos {
		name := file.Name()
		if file.IsDir() {
			folders = append(folders, name)
		} else if name != "Desktop.ini" && name != "desktop.ini" && name != "thumbs.db" {
			files = append(files, name)
		}
	}

	return
}

type IconContainer struct {
	Pfad string
	ID   int
	Name string
}

var (
	RecycleBinIcon IconContainer
	DesktopIcon    IconContainer
	FolderIcon     IconContainer
	RunIcon        IconContainer
)

func getDefaultValue(k registry.Key, path string) string {
	key, err := registry.OpenKey(k, path, registry.QUERY_VALUE)
	if err != nil {
		log.Fatal(err)
	}
	value, _, err := key.GetStringValue("")
	if err != nil {
		log.Fatal(err)
	}

	if err := key.Close(); err != nil {
		log.Fatal(err)
	}

	return value
}

func getMUINames(k registry.Key, path, alternativeName string) string {
	key, err := registry.OpenKey(k, path, registry.QUERY_VALUE)
	if err != nil {
		log.Fatal(err)
	}
	defer key.Close()

	err = registry.LoadRegLoadMUIString()
	if err == nil {
		value, err := key.GetMUIStringValue("LocalizedString")
		if err != nil {
			return alternativeName
		}
		return value
	}

	return alternativeName
}

func AddSubMenu(contextmenu *winc.MenuItem, name string, targetfolder []string) {
	var files [][]string
	var folders = map[string][]string{}
	for _, folder := range targetfolder {
		subfiles, subfolder := osReadDir(folder)
		for _, f := range subfiles {
			files = append(files, []string{folder, f})
		}
		for _, f := range subfolder {
			folders[f] = append(folders[f], filepath.Join(folder, f))

		}
	}

	submenu := contextmenu.AddSubMenu(name)
	submenu.SetImage(FolderIconhBmp)

	for name, paths := range folders {
		AddSubMenu(submenu, name, paths)
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i][1] < files[j][1]
	})

	for _, file := range files {
		var (
			iconPath  string
			iconIndex int32
		)

		command := winc.Command{
			Workdir:  file[0],
			Filename: file[1],
		}

		if strings.ToLower(filepath.Ext(file[1])) == ".lnk" {
			Lnk, err := lnk.File(filepath.Join(file[0], file[1]))
			if err != nil {
				panic(err)
			}

			if Lnk.StringData.IconLocation != "" {
				iconPath = Lnk.StringData.IconLocation
			}
			if Lnk.Header.IconIndex != 0 {
				iconIndex = Lnk.Header.IconIndex
			}
			if Lnk.LinkInfo.LocalBasePath != "" {
				command.Filename = Lnk.LinkInfo.LocalBasePath
			}
			if Lnk.StringData.WorkingDir != "" {
				command.Workdir = Lnk.StringData.WorkingDir
			}
			if Lnk.StringData.CommandLineArguments != "" {
				command.Arguments = []string{Lnk.StringData.CommandLineArguments}
			}
		}

		var item *winc.MenuItem
		if iconIndex == 0 && iconPath == "" {
			item = submenu.AddItem(fileNameWithoutExt(file[1]), winc.NoShortcut)
			hI := winc.GetIcon(filepath.Join(file[0], file[1]))
			if hI != 0 {
				ic, err := winc.NewIconFromHICONForDPI(hI, 96)
				if err == nil {
					hBmp, err := winc.NewBitmapFromIconForDPI(ic, w32.Size{Width: 16, Height: 16}, 96)
					if err != nil {
						log.Println(err)
					}
					item.SetImage(hBmp)
				} else {
					panic(err)
				}
			}
		} else {
			item = submenu.AddItemWithBitmap(fileNameWithoutExt(file[1]), winc.NoShortcut, winc.GetBitmap(iconPath, int(iconIndex)))
		}
		item.Command = command
		item.OnClick().Bind(ContextmenuFunc)
	}
}

func AddItem(contextmenu *winc.MenuItem, menu *Contextmenu) {
	var newMenu *winc.MenuItem

	if menu.Icon.Filename != "" {
		newMenu = contextmenu.AddItemWithBitmap(menu.Name, winc.NoShortcut, winc.GetBitmap(ResolveVariables(menu.Icon.Filename), menu.Icon.Index))
	} else {
		var path string
		switch {
		case menu.ShellExecute != "":
			path = menu.ShellExecute
		case menu.CreateProcess != "":
			path = menu.CreateProcess
		case menu.OpenProcess != "":
			path = menu.OpenProcess
		}
		hIcon := winc.GetIcon(ResolveVariables(path))
		if hIcon == 0 {
			newMenu = contextmenu.AddItem(menu.Name, winc.NoShortcut)
		} else {
			hBmp := winc.GetBitmapFromIcon(hIcon, w32.Size{Width: 16, Height: 16}, 96)
			if hBmp.GetHBITMAP() == 0 {
				newMenu = contextmenu.AddItem(menu.Name, winc.NoShortcut)
			} else {
				newMenu = contextmenu.AddItemWithBitmap(menu.Name, winc.NoShortcut, hBmp)
			}
		}
	}

	switch {
	case menu.ShellExecute != "":
		newMenu.OnClick().Bind(func(_ *winc.Event) {
			shellExecute(ResolveVariables(menu.ShellExecute), menu.Args, menu.Hidden)
		})
	case menu.CreateProcess != "":
		newMenu.OnClick().Bind(func(_ *winc.Event) {
			createProcess(ResolveVariables(menu.CreateProcess), menu.Args, menu.Hidden)
		})
	case menu.OpenProcess != "":
		newMenu.OnClick().Bind(func(_ *winc.Event) {
			openProcess(ResolveVariables(menu.OpenProcess), menu.Args, menu.Hidden)
		})
	}
}

// func hIconForFilePath(filePath string) w32.HICON {
// 	fPptr, _ := syscall.UTF16PtrFromString(filePath)
// 	var shfi w32.SHFILEINFO
// 	hIml := w32.HIMAGELIST(w32.SHGetFileInfo( // https://docs.microsoft.com/en-us/windows/win32/api/shellapi/nf-shellapi-shgetfileinfow
// 		fPptr,
// 		0,
// 		&shfi,
// 		uint32(unsafe.Sizeof(shfi)),
// 		w32.SHGFI_ICON|w32.SHGFI_SMALLICON))
// 	if hIml != 0 {
// 		return shfi.HIcon
// 	}
// 	return 0
// }
