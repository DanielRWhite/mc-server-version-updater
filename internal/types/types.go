package types

// Options configures the behavior of a Minecraft server updater.
type Options struct {
	// AllowUnstable determines whether unstable (beta/snapshot) versions
	// should be included when filtering available versions.
	AllowUnstable bool
	// DownloadDirectory is the file system path where server files will be
	// downloaded and managed.
	DownloadDirectory string
	// ServerFileName is the name of the server jar file (typically "server.jar").
	ServerFileName string
}

// Versions represents the response from the Fabric Meta API, containing
// all available version information for games, mappings, intermediaries,
// loaders, and installers.
type Versions struct {
	Games          []Game         `json:"game"`
	Mappings       []Mapping      `json:"mappings"`
	Intermediaries []Intermediary `json:"intermediary"`
	Loaders        []Loader       `json:"loader"`
	Installers     []Installer    `json:"installer"`
}

// Game represents a Minecraft game version available for mod loading.
type Game struct {
	// Version is the Minecraft version string (e.g. "1.21.4").
	Version string `json:"version"`
	// Stable indicates whether this is a release version (true) or a snapshot/preview (false).
	Stable bool `json:"stable"`
}

// Mapping represents a version of Minecraft mapping files (e.g. Mojang -> intermediary names).
type Mapping struct {
	// GameVersion is the Minecraft version this mapping supports.
	GameVersion string `json:"gameVersion"`
	// Seperator is the character used to separate version components (e.g. ".").
	Seperator string `json:"seperator"`
	// Build is the build number of this mapping version.
	Build int `json:"build"`
	// Maven is the Maven coordinate string for this mapping.
	Maven string `json:"maven"`
	// Version is the human-readable version string (e.g. "1.21.4-1.1.2").
	Version string `json:"version"`
	// Stable indicates whether this mapping version is stable.
	Stable bool `json:"stable"`
}

// Intermediary represents an intermediary mapping version for Minecraft.
type Intermediary struct {
	// Maven is the Maven coordinate string for this intermediary.
	Maven string `json:"maven"`
	// Version is the version string of the intermediary mapping.
	Version string `json:"version"`
	// Stable indicates whether this intermediary version is stable.
	Stable bool `json:"stable"`
}

// Loader represents a Fabric loader version.
type Loader struct {
	// Seperator is the character used to separate version components.
	Seperator string `json:"seperator"`
	// Build is the build number of this loader version.
	Build int `json:"build"`
	// Maven is the Maven coordinate string for this loader.
	Maven string `json:"maven"`
	// Version is the human-readable loader version (e.g. "0.16.9").
	Version string `json:"version"`
	// Stable indicates whether this loader version is stable (true) or beta (false).
	Stable bool `json:"stable"`
}

// Installer represents a Fabric installer version.
type Installer struct {
	// URL is the download URL for the installer.
	URL string `json:"url"`
	// Maven is the Maven coordinate string for this installer.
	Maven string `json:"maven"`
	// Version is the human-readable installer version (e.g. "1.1.1").
	Version string `json:"version"`
	// Stable indicates whether this installer version is stable.
	Stable bool `json:"true"`
}
