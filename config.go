package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/leaanthony/winc/w32"
	"golang.org/x/sys/windows"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Desktop struct {
		Contextmenu struct {
			DarkMode      bool `yaml:"darkMode"`
			AddDebugEntry bool `yaml:"addDebugEntry"`
		} `yaml:"contextmenu"`
	} `yaml:"desktop"`
	Taskbar struct {
		FontFamily   string `yaml:"fontFamily"`
		FontSize     int    `yaml:"fontSize"`
		Position     string `yaml:"position"`
		IconPosition string `yaml:"iconPosition"`
		Height       int    `yaml:"height"`
		Button       struct {
			Size struct {
				Width  int `yaml:"width"`
				Height int `yaml:"height"`
			} `yaml:"size"`
			Bgcolor struct {
				R int `yaml:"r"`
				G int `yaml:"g"`
				B int `yaml:"b"`
			} `yaml:"bgcolor"`
			Textcolor struct {
				R int `yaml:"r"`
				G int `yaml:"g"`
				B int `yaml:"b"`
			} `yaml:"textcolor"`
		} `yaml:"button"`
		Bgcolor struct {
			A int `yaml:"a"`
			R int `yaml:"r"`
			G int `yaml:"g"`
			B int `yaml:"b"`
		} `yaml:"bgcolor"`
	} `yaml:"taskbar"`
	Contextmenu []Contextmenu `yaml:"contextmenu"`
	Hotkey      []struct {
		Buttons       string   `yaml:"buttons"`
		ShellExecute  string   `yaml:"shellExecute,omitempty"`
		CreateProcess string   `yaml:"createProcess,omitempty"`
		OpenProcess   string   `yaml:"openProcess,omitempty"`
		Args          []string `yaml:"args,omitempty"`
		Hidden        bool     `yaml:"hidden,omitempty"`
	} `yaml:"hotkey"`
}

type Contextmenu struct {
	Name   string   `yaml:"name"`
	Args   []string `yaml:"args,omitempty"`
	Path   []string `yaml:"path,omitempty"`
	Hidden bool     `yaml:"hidden,omitempty"`
	Icon   struct {
		Filename string `yaml:"filename"`
		Index    int    `yaml:"index"`
	} `yaml:"icon,omitempty"`
	ShellExecute  string `yaml:"shellExecute,omitempty"`
	CreateProcess string `yaml:"createProcess,omitempty"`
	OpenProcess   string `yaml:"openProcess,omitempty"`
}

var (
	exPath string
	config *Config
)

func init() {
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Lshortfile) // https://ispycode.com/GO/Logging/Setting-output-flags

	ex, err := os.Executable()
	if err != nil {
		w32.MessageBox(0, "Load config.yaml", err.Error(), w32.MB_ICONERROR)
		panic(err)
	}
	exPath = filepath.Dir(ex)

	noFilesPtr := flag.Bool("nofiles", false, "do not create RegFiles")
	startUpPtr := flag.Bool("startup", false, "start Autorun")
	flag.Parse()
	if !*noFilesPtr {
		createRegFiles()
	}

	if *startUpPtr {
		go startup()
	}

	// https://devblogs.microsoft.com/oldnewthing/20230608-00/?p=108312
	// https://learn.microsoft.com/en-us/windows/win32/api/winbase/nf-winbase-registerapplicationrestart
	w32.RegisterApplicationRestart(ex, w32.RESTART_NO_PATCH|w32.RESTART_NO_REBOOT)

	config = LoadConfig()
}

var environmentVariablesRegEx = regexp.MustCompile(`%(\w+)%`)

func ResolveVariables(data string) string {
	data = strings.ToLower(data)
	for _, ev := range environmentVariablesRegEx.FindAllString(data, -1) {
		data = strings.ReplaceAll(data, ev, os.Getenv(strings.Trim(ev, "%")))
	}
	return data
}

// https://learn.microsoft.com/de-de/windows/win32/shell/knownfolderid
var FOLDERIDs = map[string]*windows.KNOWNFOLDERID{
	"FOLDERID_NetworkFolder":          windows.FOLDERID_NetworkFolder,
	"FOLDERID_ComputerFolder":         windows.FOLDERID_ComputerFolder,
	"FOLDERID_InternetFolder":         windows.FOLDERID_InternetFolder,
	"FOLDERID_ControlPanelFolder":     windows.FOLDERID_ControlPanelFolder,
	"FOLDERID_PrintersFolder":         windows.FOLDERID_PrintersFolder,
	"FOLDERID_SyncManagerFolder":      windows.FOLDERID_SyncManagerFolder,
	"FOLDERID_SyncSetupFolder":        windows.FOLDERID_SyncSetupFolder,
	"FOLDERID_ConflictFolder":         windows.FOLDERID_ConflictFolder,
	"FOLDERID_SyncResultsFolder":      windows.FOLDERID_SyncResultsFolder,
	"FOLDERID_RecycleBinFolder":       windows.FOLDERID_RecycleBinFolder,
	"FOLDERID_ConnectionsFolder":      windows.FOLDERID_ConnectionsFolder,
	"FOLDERID_Fonts":                  windows.FOLDERID_Fonts,
	"FOLDERID_Desktop":                windows.FOLDERID_Desktop,
	"FOLDERID_Startup":                windows.FOLDERID_Startup,
	"FOLDERID_Programs":               windows.FOLDERID_Programs,
	"FOLDERID_StartMenu":              windows.FOLDERID_StartMenu,
	"FOLDERID_Recent":                 windows.FOLDERID_Recent,
	"FOLDERID_SendTo":                 windows.FOLDERID_SendTo,
	"FOLDERID_Documents":              windows.FOLDERID_Documents,
	"FOLDERID_Favorites":              windows.FOLDERID_Favorites,
	"FOLDERID_NetHood":                windows.FOLDERID_NetHood,
	"FOLDERID_PrintHood":              windows.FOLDERID_PrintHood,
	"FOLDERID_Templates":              windows.FOLDERID_Templates,
	"FOLDERID_CommonStartup":          windows.FOLDERID_CommonStartup,
	"FOLDERID_CommonPrograms":         windows.FOLDERID_CommonPrograms,
	"FOLDERID_CommonStartMenu":        windows.FOLDERID_CommonStartMenu,
	"FOLDERID_PublicDesktop":          windows.FOLDERID_PublicDesktop,
	"FOLDERID_ProgramData":            windows.FOLDERID_ProgramData,
	"FOLDERID_CommonTemplates":        windows.FOLDERID_CommonTemplates,
	"FOLDERID_PublicDocuments":        windows.FOLDERID_PublicDocuments,
	"FOLDERID_RoamingAppData":         windows.FOLDERID_RoamingAppData,
	"FOLDERID_LocalAppData":           windows.FOLDERID_LocalAppData,
	"FOLDERID_LocalAppDataLow":        windows.FOLDERID_LocalAppDataLow,
	"FOLDERID_InternetCache":          windows.FOLDERID_InternetCache,
	"FOLDERID_Cookies":                windows.FOLDERID_Cookies,
	"FOLDERID_History":                windows.FOLDERID_History,
	"FOLDERID_System":                 windows.FOLDERID_System,
	"FOLDERID_SystemX86":              windows.FOLDERID_SystemX86,
	"FOLDERID_Windows":                windows.FOLDERID_Windows,
	"FOLDERID_Profile":                windows.FOLDERID_Profile,
	"FOLDERID_Pictures":               windows.FOLDERID_Pictures,
	"FOLDERID_ProgramFilesX86":        windows.FOLDERID_ProgramFilesX86,
	"FOLDERID_ProgramFilesCommonX86":  windows.FOLDERID_ProgramFilesCommonX86,
	"FOLDERID_ProgramFilesX64":        windows.FOLDERID_ProgramFilesX64,
	"FOLDERID_ProgramFilesCommonX64":  windows.FOLDERID_ProgramFilesCommonX64,
	"FOLDERID_ProgramFiles":           windows.FOLDERID_ProgramFiles,
	"FOLDERID_ProgramFilesCommon":     windows.FOLDERID_ProgramFilesCommon,
	"FOLDERID_UserProgramFiles":       windows.FOLDERID_UserProgramFiles,
	"FOLDERID_UserProgramFilesCommon": windows.FOLDERID_UserProgramFilesCommon,
	"FOLDERID_AdminTools":             windows.FOLDERID_AdminTools,
	"FOLDERID_CommonAdminTools":       windows.FOLDERID_CommonAdminTools,
	"FOLDERID_Music":                  windows.FOLDERID_Music,
	"FOLDERID_Videos":                 windows.FOLDERID_Videos,
	"FOLDERID_Ringtones":              windows.FOLDERID_Ringtones,
	"FOLDERID_PublicPictures":         windows.FOLDERID_PublicPictures,
	"FOLDERID_PublicMusic":            windows.FOLDERID_PublicMusic,
	"FOLDERID_PublicVideos":           windows.FOLDERID_PublicVideos,
	"FOLDERID_PublicRingtones":        windows.FOLDERID_PublicRingtones,
	"FOLDERID_ResourceDir":            windows.FOLDERID_ResourceDir,
	"FOLDERID_LocalizedResourcesDir":  windows.FOLDERID_LocalizedResourcesDir,
	"FOLDERID_CommonOEMLinks":         windows.FOLDERID_CommonOEMLinks,
	"FOLDERID_CDBurning":              windows.FOLDERID_CDBurning,
	"FOLDERID_UserProfiles":           windows.FOLDERID_UserProfiles,
	"FOLDERID_Playlists":              windows.FOLDERID_Playlists,
	"FOLDERID_SamplePlaylists":        windows.FOLDERID_SamplePlaylists,
	"FOLDERID_SampleMusic":            windows.FOLDERID_SampleMusic,
	"FOLDERID_SamplePictures":         windows.FOLDERID_SamplePictures,
	"FOLDERID_SampleVideos":           windows.FOLDERID_SampleVideos,
	"FOLDERID_PhotoAlbums":            windows.FOLDERID_PhotoAlbums,
	"FOLDERID_Public":                 windows.FOLDERID_Public,
	"FOLDERID_ChangeRemovePrograms":   windows.FOLDERID_ChangeRemovePrograms,
	"FOLDERID_AppUpdates":             windows.FOLDERID_AppUpdates,
	"FOLDERID_AddNewPrograms":         windows.FOLDERID_AddNewPrograms,
	"FOLDERID_Downloads":              windows.FOLDERID_Downloads,
	"FOLDERID_PublicDownloads":        windows.FOLDERID_PublicDownloads,
	"FOLDERID_SavedSearches":          windows.FOLDERID_SavedSearches,
	"FOLDERID_QuickLaunch":            windows.FOLDERID_QuickLaunch,
	"FOLDERID_Contacts":               windows.FOLDERID_Contacts,
	"FOLDERID_SidebarParts":           windows.FOLDERID_SidebarParts,
	"FOLDERID_SidebarDefaultParts":    windows.FOLDERID_SidebarDefaultParts,
	"FOLDERID_PublicGameTasks":        windows.FOLDERID_PublicGameTasks,
	"FOLDERID_GameTasks":              windows.FOLDERID_GameTasks,
	"FOLDERID_SavedGames":             windows.FOLDERID_SavedGames,
	"FOLDERID_Games":                  windows.FOLDERID_Games,
	"FOLDERID_SEARCH_MAPI":            windows.FOLDERID_SEARCH_MAPI,
	"FOLDERID_SEARCH_CSC":             windows.FOLDERID_SEARCH_CSC,
	"FOLDERID_Links":                  windows.FOLDERID_Links,
	"FOLDERID_UsersFiles":             windows.FOLDERID_UsersFiles,
	"FOLDERID_UsersLibraries":         windows.FOLDERID_UsersLibraries,
	"FOLDERID_SearchHome":             windows.FOLDERID_SearchHome,
	"FOLDERID_OriginalImages":         windows.FOLDERID_OriginalImages,
	"FOLDERID_DocumentsLibrary":       windows.FOLDERID_DocumentsLibrary,
	"FOLDERID_MusicLibrary":           windows.FOLDERID_MusicLibrary,
	"FOLDERID_PicturesLibrary":        windows.FOLDERID_PicturesLibrary,
	"FOLDERID_VideosLibrary":          windows.FOLDERID_VideosLibrary,
	"FOLDERID_RecordedTVLibrary":      windows.FOLDERID_RecordedTVLibrary,
	"FOLDERID_HomeGroup":              windows.FOLDERID_HomeGroup,
	"FOLDERID_HomeGroupCurrentUser":   windows.FOLDERID_HomeGroupCurrentUser,
	"FOLDERID_DeviceMetadataStore":    windows.FOLDERID_DeviceMetadataStore,
	"FOLDERID_Libraries":              windows.FOLDERID_Libraries,
	"FOLDERID_PublicLibraries":        windows.FOLDERID_PublicLibraries,
	"FOLDERID_UserPinned":             windows.FOLDERID_UserPinned,
	"FOLDERID_ImplicitAppShortcuts":   windows.FOLDERID_ImplicitAppShortcuts,
	"FOLDERID_AccountPictures":        windows.FOLDERID_AccountPictures,
	"FOLDERID_PublicUserTiles":        windows.FOLDERID_PublicUserTiles,
	"FOLDERID_AppsFolder":             windows.FOLDERID_AppsFolder,
	"FOLDERID_StartMenuAllPrograms":   windows.FOLDERID_StartMenuAllPrograms,
	"FOLDERID_CommonStartMenuPlaces":  windows.FOLDERID_CommonStartMenuPlaces,
	"FOLDERID_ApplicationShortcuts":   windows.FOLDERID_ApplicationShortcuts,
	"FOLDERID_RoamingTiles":           windows.FOLDERID_RoamingTiles,
	"FOLDERID_RoamedTileImages":       windows.FOLDERID_RoamedTileImages,
	"FOLDERID_Screenshots":            windows.FOLDERID_Screenshots,
	"FOLDERID_CameraRoll":             windows.FOLDERID_CameraRoll,
	"FOLDERID_SkyDrive":               windows.FOLDERID_SkyDrive,
	"FOLDERID_OneDrive":               windows.FOLDERID_OneDrive,
	"FOLDERID_SkyDriveDocuments":      windows.FOLDERID_SkyDriveDocuments,
	"FOLDERID_SkyDrivePictures":       windows.FOLDERID_SkyDrivePictures,
	"FOLDERID_SkyDriveMusic":          windows.FOLDERID_SkyDriveMusic,
	"FOLDERID_SkyDriveCameraRoll":     windows.FOLDERID_SkyDriveCameraRoll,
	"FOLDERID_SearchHistory":          windows.FOLDERID_SearchHistory,
	"FOLDERID_SearchTemplates":        windows.FOLDERID_SearchTemplates,
	"FOLDERID_CameraRollLibrary":      windows.FOLDERID_CameraRollLibrary,
	"FOLDERID_SavedPictures":          windows.FOLDERID_SavedPictures,
	"FOLDERID_SavedPicturesLibrary":   windows.FOLDERID_SavedPicturesLibrary,
	"FOLDERID_RetailDemo":             windows.FOLDERID_RetailDemo,
	"FOLDERID_Device":                 windows.FOLDERID_Device,
	"FOLDERID_DevelopmentFiles":       windows.FOLDERID_DevelopmentFiles,
	"FOLDERID_Objects3D":              windows.FOLDERID_Objects3D,
	"FOLDERID_AppCaptures":            windows.FOLDERID_AppCaptures,
	"FOLDERID_LocalDocuments":         windows.FOLDERID_LocalDocuments,
	"FOLDERID_LocalPictures":          windows.FOLDERID_LocalPictures,
	"FOLDERID_LocalVideos":            windows.FOLDERID_LocalVideos,
	"FOLDERID_LocalMusic":             windows.FOLDERID_LocalMusic,
	"FOLDERID_LocalDownloads":         windows.FOLDERID_LocalDownloads,
	"FOLDERID_RecordedCalls":          windows.FOLDERID_RecordedCalls,
	"FOLDERID_AllAppMods":             windows.FOLDERID_AllAppMods,
	"FOLDERID_CurrentAppMods":         windows.FOLDERID_CurrentAppMods,
	"FOLDERID_AppDataDesktop":         windows.FOLDERID_AppDataDesktop,
	"FOLDERID_AppDataDocuments":       windows.FOLDERID_AppDataDocuments,
	"FOLDERID_AppDataFavorites":       windows.FOLDERID_AppDataFavorites,
	"FOLDERID_AppDataProgramData":     windows.FOLDERID_AppDataProgramData,
}

func LoadConfig() *Config {
	content, err := os.ReadFile(filepath.Join(exPath, "config.yaml"))
	if err != nil {
		w32.MessageBox(0, "Load config.yaml", err.Error(), w32.MB_ICONERROR)
		log.Println(err)
	}
	c := Config{}
	err = yaml.Unmarshal(content, &c)
	if err != nil {
		w32.MessageBox(0, "Load config.yaml", err.Error(), w32.MB_ICONERROR)
		log.Printf("cannot unmarshal data: %v", err)
	}

	// Default Values
	if c.Taskbar.Height == 0 {
		c.Taskbar.Height = 30
	}

	if c.Taskbar.FontFamily == "" {
		c.Taskbar.FontFamily = "Segoe UI"
	}
	if c.Taskbar.FontSize == 0 {
		c.Taskbar.FontSize = 9
	}

	if c.Taskbar.Button.Size.Width == 0 {
		c.Taskbar.Button.Size.Width = 160
	}
	if c.Taskbar.Button.Size.Height == 0 {
		c.Taskbar.Button.Size.Height = 30
	}

	c.Taskbar.IconPosition = strings.ToLower(c.Taskbar.IconPosition)
	c.Taskbar.Position = strings.ToLower(c.Taskbar.Position)

	// resolve file paths
	for i := 0; i < len(c.Contextmenu); i++ {
		for j := 0; j < len(c.Contextmenu[i].Path); j++ {
			temp := c.Contextmenu[i].Path[j]
			val, ok := FOLDERIDs[temp]
			if ok {
				c.Contextmenu[i].Path[j] = getKnownFolderPath(val)
			} else {
				c.Contextmenu[i].Path[j] = ResolveVariables(temp)
			}
		}
	}

	return &c
}
