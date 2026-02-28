// Package skill handles loading and parsing SKILL.md files
package skill

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// Skill represents a loaded skill definition
type Skill struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Tags        []string `yaml:"tags"`
	Content     string   `yaml:"-"` // The markdown content after front-matter
	Path        string   `yaml:"-"` // Source file path
}

// Loader handles skill discovery and loading
type Loader struct {
	paths  []string
	skills map[string]*Skill
}

// NewLoader creates a new skill loader
func NewLoader(paths []string) *Loader {
	return &Loader{
		paths:  paths,
		skills: make(map[string]*Skill),
	}
}

// frontMatterRegex matches YAML front-matter between --- delimiters
var frontMatterRegex = regexp.MustCompile(`(?s)^---\n(.+?)\n---\n(.*)$`)

// Load discovers and loads all skills from configured paths
func (l *Loader) Load() error {
	for _, basePath := range l.paths {
		// Expand ~ in path
		if strings.HasPrefix(basePath, "~") {
			if home, err := os.UserHomeDir(); err == nil {
				basePath = filepath.Join(home, basePath[1:])
			}
		}

		// Check if path exists
		info, err := os.Stat(basePath)
		if err != nil {
			continue // Skip non-existent paths
		}

		if info.IsDir() {
			// Load all SKILL.md files in directory
			if err := l.loadDir(basePath); err != nil {
				return err
			}
		} else if strings.HasSuffix(basePath, ".md") {
			// Load single file
			if err := l.loadFile(basePath); err != nil {
				return err
			}
		}
	}

	return nil
}

func (l *Loader) loadDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read skills dir %s: %w", dir, err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			// Check for SKILL.md in subdirectory
			skillPath := filepath.Join(dir, entry.Name(), "SKILL.md")
			if _, err := os.Stat(skillPath); err == nil {
				if err := l.loadFile(skillPath); err != nil {
					return err
				}
			}
		} else if strings.HasSuffix(entry.Name(), ".md") {
			if err := l.loadFile(filepath.Join(dir, entry.Name())); err != nil {
				return err
			}
		}
	}

	return nil
}

func (l *Loader) loadFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read skill %s: %w", path, err)
	}

	skill, err := Parse(string(data))
	if err != nil {
		return fmt.Errorf("parse skill %s: %w", path, err)
	}

	skill.Path = path
	l.skills[skill.Name] = skill
	return nil
}

// Parse parses a skill from markdown content with YAML front-matter
func Parse(content string) (*Skill, error) {
	matches := frontMatterRegex.FindStringSubmatch(content)
	if matches == nil {
		// No front-matter, use filename or default
		return &Skill{
			Name:    "unnamed",
			Content: content,
		}, nil
	}

	var skill Skill
	if err := yaml.Unmarshal([]byte(matches[1]), &skill); err != nil {
		return nil, fmt.Errorf("parse front-matter: %w", err)
	}

	skill.Content = strings.TrimSpace(matches[2])
	return &skill, nil
}

// Get retrieves a skill by name
func (l *Loader) Get(name string) (*Skill, bool) {
	skill, ok := l.skills[name]
	return skill, ok
}

// List returns all loaded skills
func (l *Loader) List() []*Skill {
	skills := make([]*Skill, 0, len(l.skills))
	for _, s := range l.skills {
		skills = append(skills, s)
	}
	return skills
}

// Match finds skills matching a description using simple keyword matching
func (l *Loader) Match(description string) []*Skill {
	description = strings.ToLower(description)
	words := strings.Fields(description)

	var matches []*Skill
	for _, skill := range l.skills {
		score := 0
		skillText := strings.ToLower(skill.Name + " " + skill.Description + " " + strings.Join(skill.Tags, " "))
		
		for _, word := range words {
			if strings.Contains(skillText, word) {
				score++
			}
		}
		
		if score > 0 {
			matches = append(matches, skill)
		}
	}

	return matches
}

// Names returns all skill names
func (l *Loader) Names() []string {
	names := make([]string, 0, len(l.skills))
	for name := range l.skills {
		names = append(names, name)
	}
	return names
}
