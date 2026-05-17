package main

import (
	"flag"
	"fmt"
	"os"

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
	} else {
		// If this option isn't set, try to get the current working directory, if not default to a temp dir
		wd, err := os.Getwd()
		if err != nil {
			options.DownloadDirectory = os.TempDir()
		}

		options.DownloadDirectory = wd
	}

	if serverFileName != nil {
		options.ServerFileName = *serverFileName
	}

	if allowUnstable != nil {
		options.AllowUnstable = *allowUnstable
	}

	fmt.Printf("[Updater] Downloading to: %s\n", options.DownloadDirectory)
	fmt.Printf("[Updater] Server Filename: %s\n", options.ServerFileName)
	fmt.Printf("[Updater] Allow Unstable: %t\n", options.AllowUnstable)

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
