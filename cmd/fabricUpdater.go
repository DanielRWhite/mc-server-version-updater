package main

import (
	"flag"

	"github.com/DanielRWhite/fabric-mc-server-updater/internal/types"
	"github.com/DanielRWhite/fabric-mc-server-updater/internal/updaters"
)

func main() {
	directoryFlag := flag.String("dir", "", "The directory to download the server files to")
	serverFileName := flag.String("name", "server.jar", "The name to save the server .jar file as")
	allowUnstable := flag.Bool("unstable", false, "Allow unstable minecraft server builds (e.g. snapshots)")

	flag.Parse()

	options := types.Options{}
	if directoryFlag != nil {
		options.DownloadDirectory = *directoryFlag
	}

	if serverFileName != nil {
		options.ServerFileName = *serverFileName
	}

	if allowUnstable != nil {
		options.AllowUnstable = *allowUnstable
	}

	fabricUpdater := updaters.NewFabricUpdater(&options)

	if err := fabricUpdater.GetVersions(); err != nil {
		panic(err)
	}

	if err := fabricUpdater.FilterVersions(); err != nil {
		panic(err)
	}

	if err := fabricUpdater.GetLatest(); err != nil {
		panic(err)
	}

	if err := fabricUpdater.Download(); err != nil {
		panic(err)
	}

	if err := fabricUpdater.Migrate(); err != nil {
		panic(err)
	}
}
