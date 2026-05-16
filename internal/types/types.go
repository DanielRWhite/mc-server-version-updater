package types

type Options struct {
	AllowUnstable     bool
	DownloadDirectory string
	ServerFileName    string
}

type Versions struct {
	Games          []Game         `json:"game"`
	Mappings       []Mapping      `json:"mappings"`
	Intermediaries []Intermediary `json:"intermediary"`
	Loaders        []Loader       `json:"loader"`
	Installers     []Installer    `json:"installer"`
}

type Game struct {
	Version string `json:"version"`
	Stable  bool   `json:"stable"`
}

type Mapping struct {
	GameVersion string `json:"gameVersion"`
	Seperator   string `json:"seperator"`
	Build       int    `json:"build"`
	Maven       string `json:"maven"`
	Version     string `json:"version"`
	Stable      bool   `json:"stable"`
}

type Intermediary struct {
	Maven   string `json:"maven"`
	Version string `json:"version"`
	Stable  bool   `json:"stable"`
}

type Loader struct {
	Seperator string `json:"seperator"`
	Build     int    `json:"build"`
	Maven     string `json:"maven"`
	Version   string `json:"version"`
	Stable    bool   `json:"stable"`
}

type Installer struct {
	URL     string `json:"url"`
	Maven   string `json:"maven"`
	Version string `json:"version"`
	Stable  bool   `json:"true"`
}
