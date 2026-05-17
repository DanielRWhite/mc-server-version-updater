# Fabric MC Updater

A command-line tool for automatically updating Fabric Minecraft servers. This tool fetches the latest available versions from the [Fabric Meta API](https://meta.fabricmc.net/), downloads the appropriate server loader jar, and safely migrates your existing server installation with automatic backups.

## Features

- **Automatic version detection**: Fetches available Minecraft game versions, Fabric loaders, and installers from the official Fabric Meta API
- **Smart filtering**: Option to include or exclude unstable/snapshot versions
- **Safe migration**: Automatically backs up your existing server files before updating
- **Duplicate detection**: Compares SHA256 hashes to avoid unnecessary updates
- **Timestamped backups**: Previous server versions are archived in `migrated_versions/` with timestamps

## Prerequisites

- [Go 1.26+](https://go.dev/dl/)
- Access to a Fabric Minecraft server directory

## Installation

```bash
# Clone the repository
git clone https://github.com/DanielRWhite/fabric-mc-server-updater.git
cd fabric-mc-server-updater

# Build the binary
go build -o fabric-updater ./cmd/fabricUpdater.go

# Or run directly
go run ./cmd/fabricUpdater.go
```

## Usage

Run the updater with the following command:

```bash
./fabric-updater [flags]
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-dir` | Current working directory | The directory where server files will be downloaded and managed |
| `-name` | `server.jar` | The name to save the server jar file as |
| `-unstable` | `false` | Allow unstable Minecraft server builds (e.g. snapshots, betas) |

### Examples

Update a server in the current directory:
```bash
./fabric-updater
```

Update a server in a specific directory:
```bash
./fabric-updater -dir /path/to/your/server
```

Allow snapshot/beta versions:
```bash
./fabric-updater -dir /path/to/your/server -unstable
```

Custom server filename:
```bash
./fabric-updater -dir /path/to/your/server -name fabric-server.jar
```

## How It Works

The updater performs the following steps in order:

1. **GetVersions**: Fetches available version information from the Fabric Meta API (`https://meta.fabricmc.net/v1/versions`)
2. **FilterVersions**: Removes unstable versions based on your configuration (unless `-unstable` is set)
3. **GetLatest**: Selects the latest compatible game, loader, and installer versions
4. **Download**: Downloads the Fabric server loader jar to your specified directory
5. **Migrate**: Safely replaces your existing server with the new version, creating a backup of your old files

### Migration Process

During migration, the tool:
- Compares SHA256 hashes of the existing and new server jars
- If identical, skips migration and removes the downloaded jar
- Creates a timestamped backup directory under `migrated_versions/YYYYMMDD-HHMMSS/`
- Moves your existing `server.jar` and `mods/` folder to the backup
- Places the new server jar and creates a fresh `mods/` folder

## Project Structure

```
.
├── cmd/
│   └── fabricUpdater.go      # CLI entry point
├── internal/
│   ├── types/
│   │   └── types.go          # Shared types and data structures
│   └── updaters/
│       ├── interfaces.go     # Updater interface definition
│       └── fabric.go         # Fabric-specific implementation
├── test/                     # Tests (coming soon)
├── go.mod
└── go.sum
```

## Dependencies

- [hashicorp/go-version](https://github.com/hashicorp/go-version) - Semantic version comparison

## License

This project is open source and available under the [MIT License](LICENSE).

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Acknowledgments

- [FabricMC](https://fabricmc.net/) - The mod loader this tool is built for
- [Fabric Meta API](https://meta.fabricmc.net/) - Provides version information
