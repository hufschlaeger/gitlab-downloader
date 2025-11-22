package cli

import (
	"flag"
	"fmt"
)

type Config struct {
	Token    string
	Proxy    string
	ExtIndex int
	Output   string
	Release  string
	Project  string
}

func ParseFlags() *Config {
	config := &Config{}

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

	return config
}

func (c *Config) Validate() error {
	if c.Token == "" {
		return fmt.Errorf("token is required")
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
	return nil
}
