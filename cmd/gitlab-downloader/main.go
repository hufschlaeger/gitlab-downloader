// cmd/gitlab-downloader/main.go
package main

import (
	"flag"
	"fmt"
	"os"

	"hufschlaeger.net/gitlab-downloader/internal/adapters/primary/cli"
	"hufschlaeger.net/gitlab-downloader/internal/adapters/secondary/gitlab"
	"hufschlaeger.net/gitlab-downloader/internal/adapters/secondary/http"
	"hufschlaeger.net/gitlab-downloader/internal/core/services"
)

func main() {
	// Parse CLI flags
	config := cli.ParseFlags()

	if err := config.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		flag.Usage()
		os.Exit(1)
	}

	// Secondary Adapters (Driven)
	httpClient := http.NewInsecureClient(config.Proxy)
	gitlabAdapter := gitlab.NewAdapter(config.GitLabURL, config.Token, httpClient)
	downloadAdapter := http.NewDownloadAdapter(httpClient)
	fileAdapter := http.NewFileAdapter()

	// Core Service
	releaseService := services.NewReleaseService(gitlabAdapter, downloadAdapter, fileAdapter)

	// Primary Adapter (Driver)
	cliAdapter := cli.NewAdapter(releaseService)

	// Execute
	if err := cliAdapter.DownloadRelease(config); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Download completed successfully")
}
