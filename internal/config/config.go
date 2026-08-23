package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Host identifies which forge CLI backs a project.
type Host string

const (
	HostGitHub Host = "github"
	HostGitLab Host = "gitlab"
)

// Project is one tracked repository.
type Project struct {
	ID    string `yaml:"id" json:"id"`
	Label string `yaml:"label" json:"label"`
	Host  Host   `yaml:"host" json:"host"`
	Path  string `yaml:"path" json:"path"`
}

// File is the projects registry on disk.
type File struct {
	Projects []Project `yaml:"projects"`
}

// Load reads a projects YAML file.
func Load(path string) (File, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return File{}, fmt.Errorf("read projects: %w", err)
	}
	var doc File
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		return File{}, fmt.Errorf("parse projects: %w", err)
	}
	for i, p := range doc.Projects {
		if err := validateProject(p); err != nil {
			return File{}, fmt.Errorf("projects[%d]: %w", i, err)
		}
		doc.Projects[i].Host = Host(strings.ToLower(string(p.Host)))
	}
	return doc, nil
}

func validateProject(p Project) error {
	if strings.TrimSpace(p.ID) == "" {
		return fmt.Errorf("missing id")
	}
	if strings.TrimSpace(p.Label) == "" {
		return fmt.Errorf("missing label")
	}
	switch Host(strings.ToLower(string(p.Host))) {
	case HostGitHub, HostGitLab:
	default:
		return fmt.Errorf("host must be github or gitlab")
	}
	parts := strings.Split(strings.Trim(p.Path, "/"), "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return fmt.Errorf("path must be owner/repo")
	}
	return nil
}

// OpenURL returns the human web URL for a project.
func (p Project) OpenURL() string {
	owner, repo := p.OwnerRepo()
	switch p.Host {
	case HostGitHub:
		return "https://github.com/" + owner + "/" + repo
	default:
		return "https://gitlab.com/" + owner + "/" + repo
	}
}

// OwnerRepo splits path into owner and repo.
func (p Project) OwnerRepo() (string, string) {
	parts := strings.Split(strings.Trim(p.Path, "/"), "/")
	return parts[0], parts[1]
}

// DefaultProjectsPath resolves projects.yaml next to the binary or cwd.
func DefaultProjectsPath() string {
	if v := strings.TrimSpace(os.Getenv("GITBOARD_PROJECTS")); v != "" {
		return v
	}
	cwd, _ := os.Getwd()
	return filepath.Join(cwd, "projects.yaml")
}
