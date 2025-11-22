// internal/adapters/primary/cli/config.go
package cli

import (
	"flag"
	"fmt"
	"os"
)

const (
	DefaultGitLabURL = "https://gitlab.com"
)

type Config struct {
	GitLabURL string
	Token     string
	Proxy     string
	ExtIndex  int
	Output    string
	Release   string
	Project   string
}

func ParseFlags() *Config {
	config := &Config{}

	// GitLab URL - Priority: CLI flag -> ENV -> Default
	gitlabURL := flag.String("gitlab-url", "", "GitLab instance URL")
	flag.StringVar(&config.Token, "token", "", "Your private GitLab token (required)")
	flag.StringVar(&config.Token, "t", "", "Your private GitLab token (short)")
	flag.StringVar(&config.Proxy, "proxy", "", "Proxy URL")
	flag.IntVar(&config.ExtIndex, "ext", 0, "Source extension index (0=zip, 1=tar.gz, 2=tar.bz2, 3=tar)")
	flag.StringVar(&config.Output, "out", "", "Path to store the release (required)")
	flag.StringVar(&config.Output, "o", "", "Path to store the release (short)")
	flag.StringVar(&config.Release, "release", "", "Version string of release (required)")
	flag.StringVar(&config.Release, "r", "", "Version string of release (short)")
	flag.StringVar(&config.Project, "project", "", "Project name with namespace/group (required)")
	flag.StringVar(&config.Project, "p", "", "Project name with namespace/group (short)")

	flag.Parse()

	// Resolve GitLab URL: CLI flag -> ENV -> Default
	config.GitLabURL = resolveGitLabURL(*gitlabURL)

	// Token kann auch aus ENV kommen, wenn nicht via Flag gesetzt
	if config.Token == "" {
		config.Token = os.Getenv("GITLAB_TOKEN")
	}

	// Proxy kann auch aus ENV kommen
	if config.Proxy == "" {
		config.Proxy = os.Getenv("HTTPS_PROXY")
		if config.Proxy == "" {
			config.Proxy = os.Getenv("HTTP_PROXY")
		}
	}

	return config
}

func resolveGitLabURL(flagValue string) string {
	// 1. Priority: CLI flag
	if flagValue != "" {
		return flagValue
	}

	// 2. Priority: Environment variable
	if envURL := os.Getenv("GITLAB_URL"); envURL != "" {
		return envURL
	}

	// 3. Priority: Default
	return DefaultGitLabURL
}

func (c *Config) Validate() error {
	if c.Token == "" {
		return fmt.Errorf("token is required (use -token flag or GITLAB_TOKEN env)")
	}
	if c.Output == "" {
		return fmt.Errorf("output path is required")
	}
	if c.Release == "" {
		return fmt.Errorf("release version is required")
	}
	if c.Project == "" {
		return fmt.Errorf("project name is required")
	}
	if c.GitLabURL == "" {
		return fmt.Errorf("GitLab URL is required")
	}
	return nil
}
