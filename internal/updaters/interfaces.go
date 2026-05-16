package updaters

// Updater defines the interface for Minecraft mod loader updaters.
// Implementations handle fetching version information, filtering,
// downloading server jars, and migrating existing installations.
type Updater interface {
	// GetVersions fetches available version information from the remote API.
	GetVersions() error

	// FilterVersions removes versions that don't meet stability requirements
	// based on the updater's configuration options.
	FilterVersions() error

	// GetLatest identifies and selects the latest compatible versions
	// for game, loader, and installer.
	GetLatest() error

	// Download fetches the server jar file for the selected versions
	// and saves it to the configured download directory.
	Download() error

	// Migrate backs up the existing server installation and replaces it
	// with the newly downloaded server jar.
	Migrate() error
}
