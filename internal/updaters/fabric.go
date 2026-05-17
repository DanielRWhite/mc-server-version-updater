package updaters

import (
	"crypto/sha256"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/DanielRWhite/fabric-mc-server-updater/internal/types"
	"github.com/hashicorp/go-version"
)

const (
	// API URL of the fabric server downloads, responds with JSON in the format of types.Versions
	API_URL string = "https://meta.fabricmc.net/v1/versions"

	// First argument is the game version (e.g. 1.26.1)
	// Second argument is the loader version (e.g. 0.19.2)
	// Third argument is the installer version (e.g. 1.1.1)
	EXECUTABLE_URL string = "https://meta.fabricmc.net/v2/versions/loader/%s/%s/%s/server/jar"

	// First argument is the game version (e.g. 1.26.1)
	// Second argument is the loader version (e.g. 0.19.2)
	// Third argument is the installer version (e.g. 1.1.1)
	DOWNLOAD_FILENAME string = "fabric-server-%s-%s-%s.jar"
)

// FabricUpdater handles downloading and updating Fabric Minecraft server installations.
// It interacts with the Fabric Meta API to fetch version information and download
// the appropriate server loader jar.
type FabricUpdater struct {
	// Versions contains the available version information fetched from the Fabric API.
	Versions types.Versions
	// options contains the configuration for the updater.
	options types.Options

	// GameVersion is the selected Minecraft game version (e.g. "1.21.4").
	GameVersion *string
	// LoaderVersion is the selected Fabric loader version (e.g. "0.16.9").
	LoaderVersion *string
	// InstallerVersion is the selected Fabric installer version (e.g. "1.1.1").
	InstallerVersion *string

	// client is the HTTP client used for API requests.
	client http.Client
}

// NewFabricUpdater creates a new FabricUpdater configured with the provided options.
// If options is nil, a default empty Options struct will be used.
func NewFabricUpdater(options *types.Options) Updater {
	if options == nil {
		options = &types.Options{}
	}

	return &FabricUpdater{
		options:  *options,
		Versions: types.Versions{},
		client: http.Client{
			Transport: &http.Transport{
				TLSHandshakeTimeout: 30 * time.Second,
				DisableKeepAlives:   false,

				TLSClientConfig: &tls.Config{
					CipherSuites: []uint16{
						tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
						tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
						tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
						tls.TLS_AES_128_GCM_SHA256,
						tls.VersionTLS13,
						tls.VersionTLS10,
					},
				},
				DialTLS: func(network, addr string) (net.Conn, error) {
					return tls.Dial(network, addr, http.DefaultTransport.(*http.Transport).TLSClientConfig)
				},
			},
		},
	}
}

// GetVersions fetches available Fabric versions from the Fabric Meta API.
// It populates the internal versions field with game, loader, installer,
// mapping, and intermediary version information.
func (u *FabricUpdater) GetVersions() error {
	req, err := http.NewRequest(http.MethodGet, API_URL, nil)
	if err != nil {
		return fmt.Errorf("Failed to create new request: %v", err)
	}

	// Set required headers
	req.Header.Set("Accept", "*/*")

	// Add timestamp header (Cloudflare?)
	gmtTZ, err := time.LoadLocation("GMT")
	if err != nil {
		return fmt.Errorf("failed to load gmt timezone: %v", err)
	}

	tsGMT := time.Now().Add(-time.Minute).In(gmtTZ).Format(time.RFC1123)
	req.Header.Set("If-Modified-Since", tsGMT)

	// Actually perform request now
	resp, err := u.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed making request with DefaultClient: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("Error closing response body: %v", err)
		}
	}()

	if resp.StatusCode != 200 {
		return errors.New("Got non-200 status code response")
	}

	if resp.Header.Get("Content-Type") != "application/json" {
		return errors.New("Returned Content-Type header is not application/json")
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed reading resp body: %v", err)
	}

	if err := json.Unmarshal(b, &u.Versions); err != nil {
		fmt.Printf("%s", string(b))
		return fmt.Errorf("Failed to unmarshal response body: %s", err)
	}

	return nil
}

// GetLatest selects the latest available versions for game, loader, and installer.
// It requires that GetVersions() (and optionally FilterVersions()) has been called first.
// The selected versions are stored in the updater's gameVersion, loaderVersion,
// and installerVersion fields.
func (u *FabricUpdater) GetLatest() error {
	if len(u.Versions.Games) == 0 {
		return errors.New("game versions are empty")
	}

	if len(u.Versions.Mappings) == 0 {
		return errors.New("mapping verisons are empty")
	}

	if len(u.Versions.Intermediaries) == 0 {
		return errors.New("intermediary versions are empty")
	}

	if len(u.Versions.Loaders) == 0 {
		return errors.New("loader versions are empty")
	}

	if len(u.Versions.Installers) == 0 {
		return errors.New("installer versions are empty")
	}

	gameVersion := slices.MaxFunc(u.Versions.Games, func(a, b types.Game) int {
		return compareVersions(a.Version, b.Version)
	})

	installerVersion := slices.MaxFunc(u.Versions.Installers, func(a, b types.Installer) int {
		return compareVersions(a.Version, b.Version)
	})

	loaderVersion := slices.MaxFunc(u.Versions.Loaders, func(a, b types.Loader) int {
		return compareVersions(a.Version, b.Version)
	})

	u.GameVersion = &gameVersion.Version
	u.InstallerVersion = &installerVersion.Version
	u.LoaderVersion = &loaderVersion.Version

	return nil
}

// FilterVersions removes unstable versions from the available versions list.
// Whether unstable versions are kept depends on the AllowUnstable option.
func (u *FabricUpdater) FilterVersions() error {
	if u.Versions.Games == nil {
		return errors.New("game versions are empty")
	}

	if u.Versions.Mappings == nil {
		return errors.New("mapping verisons are empty")
	}

	if u.Versions.Intermediaries == nil {
		return errors.New("intermediary versions are empty")
	}

	if u.Versions.Loaders == nil {
		return errors.New("loader versions are empty")
	}

	if u.Versions.Installers == nil {
		return errors.New("installer versions are empty")
	}

	// Remove all unstable versions
	u.Versions.Games = slices.DeleteFunc(u.Versions.Games, func(g types.Game) bool {
		return !g.Stable && !u.options.AllowUnstable
	})

	u.Versions.Mappings = slices.DeleteFunc(u.Versions.Mappings, func(m types.Mapping) bool {
		return !m.Stable && !u.options.AllowUnstable
	})

	u.Versions.Intermediaries = slices.DeleteFunc(u.Versions.Intermediaries, func(i types.Intermediary) bool {
		return !i.Stable && !u.options.AllowUnstable
	})

	u.Versions.Loaders = slices.DeleteFunc(u.Versions.Loaders, func(l types.Loader) bool {
		return !l.Stable && !u.options.AllowUnstable
	})

	// Don't filter installers since none of them for fabric are stable?
	// u.Versions.Installers = slices.DeleteFunc(u.Versions.Installers, func(i types.Installer) bool {
	// 	return !i.Stable && !u.options.AllowUnstable
	// })

	return nil
}

// Download initiates the download of the Fabric server jar for the selected
// game, loader, and installer versions. The file is saved to the configured
// download directory with a filename indicating the version combination.
// It returns an error if no versions are selected or if the selected versions
// are invalid.
func (u *FabricUpdater) Download() error {
	if u.GameVersion == nil {
		return errors.New("no game version set")
	}

	if u.LoaderVersion == nil {
		return errors.New("no loader version set")
	}

	if u.InstallerVersion == nil {
		return errors.New("no installer version set")
	}

	if !slices.ContainsFunc(u.Versions.Games, func(g types.Game) bool {
		return g.Version == *u.GameVersion
	}) {
		return errors.New("invalid game version selected")
	}

	if !slices.ContainsFunc(u.Versions.Loaders, func(l types.Loader) bool {
		return l.Version == *u.LoaderVersion
	}) {
		return errors.New("invalid loader version selected")
	}

	if !slices.ContainsFunc(u.Versions.Installers, func(i types.Installer) bool {
		return i.Version == *u.InstallerVersion
	}) {
		return errors.New("invalid installer version selected")
	}

	downloadURL := fmt.Sprintf(EXECUTABLE_URL, *u.GameVersion, *u.LoaderVersion, *u.InstallerVersion)
	req, err := http.NewRequest(http.MethodGet, downloadURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create download request: %v", err)
	}

	// Set required header
	req.Header.Set("Accept", "*/*")

	resp, err := u.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to do request: %v", err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("Failed to close request body: %v", err)
		}
	}()

	// All is ok, create file on disk now
	downloadFilename := fmt.Sprintf(DOWNLOAD_FILENAME, *u.GameVersion, *u.LoaderVersion, *u.InstallerVersion)
	serverFile, err := os.Create(filepath.Join(u.options.DownloadDirectory, downloadFilename))
	if err != nil {
		return fmt.Errorf("failed to create file on disk: %v", err)
	}

	if _, err := io.Copy(serverFile, resp.Body); err != nil {
		return fmt.Errorf("failed to write response body to file contents: %v", err)
	}

	return nil
}

// Migrate handles the transition from the existing server installation to the
// newly downloaded server jar. It performs the following steps:
//   - Compares SHA256 hashes of the existing and new server jars.
//   - If identical, removes the downloaded jar and returns early.
//   - Creates a timestamped backup directory under "migrated_versions/".
//   - Moves the existing server.jar and mods/ folder to the backup directory.
//   - Renames the new server jar to server.jar and creates a fresh mods/ folder.
func (u *FabricUpdater) Migrate() error {
	existingServerFilePath := filepath.Join(u.options.DownloadDirectory, "server.jar")
	newServerFilePath := filepath.Join(u.options.DownloadDirectory, fmt.Sprintf(DOWNLOAD_FILENAME, *u.GameVersion, *u.LoaderVersion, *u.InstallerVersion))

	// Get existing server file
	existingFile, err := os.Open(existingServerFilePath)
	if err != nil {
		// Soft fail here, log it
		fmt.Printf("Failed to open previous server.jar: %s", err)
	}

	// Open new server file too, so that we can check it exists before migrating anything
	newServerFile, err := os.Open(newServerFilePath)
	if err != nil || newServerFile == nil {
		if existingFile != nil {
			existingFile.Close()
		}
		return fmt.Errorf("failed to open new server.jar file: %v", err)
	}

	// Compare hash of both server files, if they are the same exit early and delete downloaded file :)
	existingFileHash, err := getFileHash(existingFile)
	if err != nil || existingFileHash == nil {
		existingFile.Close()
		newServerFile.Close()
		return fmt.Errorf("failed to get existing server file hash: %v", err)
	}

	newFileHash, err := getFileHash(newServerFile)
	if err != nil || newFileHash == nil {
		existingFile.Close()
		newServerFile.Close()
		return fmt.Errorf("failed to get new server file hash: %v", err)
	}

	// Close files before any rename/remove operations (required on Windows).
	existingFile.Close()
	newServerFile.Close()

	if *existingFileHash == *newFileHash {
		fmt.Printf("Server version are the same, deleting downloaded one\n")
		if err := os.Remove(newServerFilePath); err != nil {
			fmt.Printf("Failed to cleanup downloaded server file: %v", err)
		}

		// return early, we don't want to migrate
		return nil
	}

	// Make backup directory
	loc, err := time.LoadLocation("Australia/Sydney")
	if err != nil {
		return errors.New("Failed to load timezone location")
	}

	migratedDirVersion := fmt.Sprintf("migrated_versions/%s", time.Now().In(loc).Format("20060102-150405"))
	migratedFilepath := filepath.Join(u.options.DownloadDirectory, migratedDirVersion)
	if err := os.MkdirAll(migratedFilepath, os.ModeDir); err != nil {
		return fmt.Errorf("failed to make migrated versions dir: %v", err)
	}

	if existingFile != nil {
		newServerPath := filepath.Join(migratedFilepath, "server.jar")
		if err := os.Rename(existingServerFilePath, newServerPath); err != nil {
			return fmt.Errorf("failed to migrate old server.jar to migrations folder")
		}

		oldModsPath := filepath.Join(u.options.DownloadDirectory, "mods")
		newModsPath := filepath.Join(migratedFilepath, "mods")
		if err := os.Rename(oldModsPath, newModsPath); err != nil {
			fmt.Printf("failed to migrate old server mods: %v", err)
		}
	}

	if err := os.Rename(newServerFilePath, filepath.Join(u.options.DownloadDirectory, "server.jar")); err != nil {
		return fmt.Errorf("failed to move old server.jar: %v", err)
	}

	if err := os.MkdirAll(filepath.Join(u.options.DownloadDirectory, "mods"), os.ModeDir); err != nil {
		fmt.Printf("Failed to create new mods folder")
	}

	return nil
}

// compareVersions compares two semantic version strings using hashicorp/go-version.
// It returns -1 if a < b, 1 if a >= b. If parsing fails, it returns a fallback
// value (-1 if a fails to parse, 1 if b fails to parse).
func compareVersions(a, b string) int {
	aVersion, err := version.NewVersion(a)
	if err != nil {
		return -1
	}

	bVersion, err := version.NewVersion(b)
	if err != nil {
		return 1
	}

	if aVersion.LessThan(bVersion) {
		return -1
	} else {
		return 1
	}
}

// getFileHash computes the SHA256 hash of the given file's contents.
// It returns a pointer to the hex-encoded hash string, or an error if
// reading the file fails.
func getFileHash(f *os.File) (*string, error) {
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return nil, err
	}

	hash := fmt.Sprintf("%x", h.Sum(nil))
	return &hash, nil
}
