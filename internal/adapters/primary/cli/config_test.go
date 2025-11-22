package cli

import (
	"flag"
	"os"
	"testing"
)

// helper to run ParseFlags with custom args and cleaned env
func runParseFlags(t *testing.T, args []string, env map[string]string) *Config {
	t.Helper()

	oldArgs := os.Args
	os.Args = append([]string{"cmd"}, args...)
	defer func() { os.Args = oldArgs }()

	// snapshot env and restore after
	oldEnv := make(map[string]string)
	for k := range env {
		oldEnv[k] = os.Getenv(k)
	}
	for k, v := range env {
		if v == "__UNSET__" {
			_ = os.Unsetenv(k)
		} else {
			_ = os.Setenv(k, v)
		}
	}
	defer func() {
		for k, v := range oldEnv {
			if v == "" {
				_ = os.Unsetenv(k)
			} else {
				_ = os.Setenv(k, v)
			}
		}
	}()

	// flag package keeps global state across tests; reset
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	flag.CommandLine.SetOutput(nil)

	return ParseFlags()
}

func TestResolveGitLabURL_Precedence(t *testing.T) {
	// Flag should win over ENV and default
	cfg := runParseFlags(t, []string{
		"-gitlab-url", "https://flag.example",
		"-token", "tok",
		"-out", "file",
		"-release", "v1",
		"-project", "grp/proj",
	}, map[string]string{
		"GITLAB_URL": "https://env.example",
	})
	if cfg.GitLabURL != "https://flag.example" {
		t.Fatalf("expected flag GitLab URL, got %q", cfg.GitLabURL)
	}
}

func TestResolveGitLabURL_FromEnvWhenNoFlag(t *testing.T) {
	cfg := runParseFlags(t, []string{
		"-token", "tok",
		"-out", "file",
		"-release", "v1",
		"-project", "grp/proj",
	}, map[string]string{
		"GITLAB_URL": "https://env.example",
	})
	if cfg.GitLabURL != "https://env.example" {
		t.Fatalf("expected env GitLab URL, got %q", cfg.GitLabURL)
	}
}

func TestResolveGitLabURL_DefaultWhenNone(t *testing.T) {
	cfg := runParseFlags(t, []string{
		"-token", "tok",
		"-out", "file",
		"-release", "v1",
		"-project", "grp/proj",
	}, map[string]string{
		"GITLAB_URL": "__UNSET__",
	})
	if cfg.GitLabURL != DefaultGitLabURL {
		t.Fatalf("expected default GitLab URL, got %q", cfg.GitLabURL)
	}
}

func TestTokenAndProxyFromEnv(t *testing.T) {
	cfg := runParseFlags(t, []string{
		"-gitlab-url", "https://x",
		"-out", "file",
		"-release", "v1",
		"-project", "grp/proj",
	}, map[string]string{
		"GITLAB_TOKEN": "envtoken",
		"HTTPS_PROXY":  "https://proxy",
	})
	if cfg.Token != "envtoken" {
		t.Fatalf("expected token from env, got %q", cfg.Token)
	}
	if cfg.Proxy != "https://proxy" {
		t.Fatalf("expected proxy from env, got %q", cfg.Proxy)
	}
}

func TestProxyFallsBackToHTTPProxy(t *testing.T) {
	cfg := runParseFlags(t, []string{
		"-gitlab-url", "https://x",
		"-out", "file",
		"-release", "v1",
		"-project", "grp/proj",
	}, map[string]string{
		"HTTPS_PROXY": "__UNSET__",
		"HTTP_PROXY":  "http://proxy",
	})
	if cfg.Proxy != "http://proxy" {
		t.Fatalf("expected HTTP_PROXY fallback, got %q", cfg.Proxy)
	}
}

func TestExtIndexParsing(t *testing.T) {
	cfg := runParseFlags(t, []string{
		"-gitlab-url", "https://x",
		"-token", "tok",
		"-out", "file",
		"-release", "v1",
		"-project", "grp/proj",
		"-ext", "2",
	}, nil)
	if cfg.ExtIndex != 2 {
		t.Fatalf("expected ExtIndex=2, got %d", cfg.ExtIndex)
	}
}

func TestValidateErrors(t *testing.T) {
	cases := []struct {
		cfg  Config
		want string
	}{
		{Config{}, "token is required"},
		{Config{Token: "t"}, "output path is required"},
		{Config{Token: "t", Output: "o"}, "release version is required"},
		{Config{Token: "t", Output: "o", Release: "r"}, "project name is required"},
		{Config{Token: "t", Output: "o", Release: "r", Project: "p"}, "GitLab URL is required"},
	}
	for _, tc := range cases {
		err := tc.cfg.Validate()
		if err == nil || !contains(err.Error(), tc.want) {
			t.Fatalf("expected error containing %q, got %v", tc.want, err)
		}
	}

	// Positive
	ok := Config{Token: "t", Output: "o", Release: "r", Project: "p", GitLabURL: "https://x"}
	if err := ok.Validate(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

// small helper wrapper to keep dependencies minimal
func contains(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && (func() bool {
		for i := 0; i+len(sub) <= len(s); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
		return false
	})())
}
