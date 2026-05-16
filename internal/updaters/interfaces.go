package updaters

type Updater interface {
	GetVersions() error

	// General
	FilterVersions() error
	GetLatest() error

	Download() error
	Migrate() error
}
