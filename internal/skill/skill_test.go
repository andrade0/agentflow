package skill

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParse_WithFrontMatter(t *testing.T) {
	content := `---
name: brainstorming
description: Design thinking before coding
tags:
  - design
  - planning
---

# Brainstorming Skill

Always think before you code.
`

	skill, err := Parse(content)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if skill.Name != "brainstorming" {
		t.Errorf("name = %q, want 'brainstorming'", skill.Name)
	}
	if skill.Description != "Design thinking before coding" {
		t.Errorf("description = %q, want 'Design thinking before coding'", skill.Description)
	}
	if len(skill.Tags) != 2 {
		t.Errorf("tags len = %d, want 2", len(skill.Tags))
	}
	if skill.Content == "" {
		t.Error("content should not be empty")
	}
	if skill.Content != "# Brainstorming Skill\n\nAlways think before you code." {
		t.Errorf("content = %q", skill.Content)
	}
}

func TestParse_WithoutFrontMatter(t *testing.T) {
	content := `# Simple Skill

Just some content without front-matter.
`

	skill, err := Parse(content)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if skill.Name != "unnamed" {
		t.Errorf("name = %q, want 'unnamed'", skill.Name)
	}
	if skill.Content != content {
		t.Errorf("content should equal original")
	}
}

func TestParse_EmptyFrontMatter(t *testing.T) {
	content := `---
name: empty-skill
---

Content here.
`

	skill, err := Parse(content)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if skill.Name != "empty-skill" {
		t.Errorf("name = %q, want 'empty-skill'", skill.Name)
	}
	if skill.Description != "" {
		t.Errorf("description should be empty")
	}
}

func TestLoader_LoadDir(t *testing.T) {
	// Create temp directory structure
	tmpDir, err := os.MkdirTemp("", "skills-test")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a skill file
	skillContent := `---
name: test-skill
description: A test skill
---

Test content.
`
	skillPath := filepath.Join(tmpDir, "test.md")
	if err := os.WriteFile(skillPath, []byte(skillContent), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	// Create subdirectory with SKILL.md
	subDir := filepath.Join(tmpDir, "sub-skill")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	subSkillContent := `---
name: sub-skill
description: A subdirectory skill
---

Sub content.
`
	if err := os.WriteFile(filepath.Join(subDir, "SKILL.md"), []byte(subSkillContent), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	// Load skills
	loader := NewLoader([]string{tmpDir})
	if err := loader.Load(); err != nil {
		t.Fatalf("Load: %v", err)
	}

	// Check loaded skills
	names := loader.Names()
	if len(names) != 2 {
		t.Errorf("expected 2 skills, got %d: %v", len(names), names)
	}

	skill, ok := loader.Get("test-skill")
	if !ok {
		t.Error("test-skill not found")
	} else if skill.Description != "A test skill" {
		t.Errorf("description = %q", skill.Description)
	}

	skill, ok = loader.Get("sub-skill")
	if !ok {
		t.Error("sub-skill not found")
	}
}

func TestLoader_Match(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "skills-match")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	skills := []struct {
		name    string
		content string
	}{
		{"brainstorming.md", `---
name: brainstorming
description: Design thinking and ideation
tags:
  - design
  - planning
---
Content.
`},
		{"tdd.md", `---
name: tdd
description: Test-driven development
tags:
  - testing
  - development
---
Content.
`},
		{"debugging.md", `---
name: debugging
description: Systematic debugging approach
tags:
  - debugging
  - analysis
---
Content.
`},
	}

	for _, s := range skills {
		if err := os.WriteFile(filepath.Join(tmpDir, s.name), []byte(s.content), 0644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}
	}

	loader := NewLoader([]string{tmpDir})
	if err := loader.Load(); err != nil {
		t.Fatalf("Load: %v", err)
	}

	// Test matching
	matches := loader.Match("design thinking")
	if len(matches) != 1 {
		t.Errorf("expected 1 match for 'design thinking', got %d", len(matches))
	}
	if len(matches) > 0 && matches[0].Name != "brainstorming" {
		t.Errorf("expected 'brainstorming', got %q", matches[0].Name)
	}

	matches = loader.Match("testing development")
	if len(matches) != 1 {
		t.Errorf("expected 1 match for 'testing development', got %d", len(matches))
	}

	matches = loader.Match("unknown")
	if len(matches) != 0 {
		t.Errorf("expected 0 matches for 'unknown', got %d", len(matches))
	}
}

func TestLoader_NonExistentPath(t *testing.T) {
	loader := NewLoader([]string{"/nonexistent/path"})
	if err := loader.Load(); err != nil {
		t.Errorf("Load should not error on nonexistent paths: %v", err)
	}
	if len(loader.List()) != 0 {
		t.Error("expected no skills")
	}
}
