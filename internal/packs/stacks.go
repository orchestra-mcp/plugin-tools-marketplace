package packs

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// StackInfo describes a detected technology stack.
type StackInfo struct {
	Name       string
	Confidence float64
	Evidence   string
}

// DetectStacks detects technology stacks in the given workspace.
func DetectStacks(workspace string) []StackInfo {
	var stacks []StackInfo

	checks := []struct {
		name     string
		check    func(string) (bool, string)
	}{
		{"go", checkAnyFileExists("go.mod", "go.work")},
		{"rust", checkFileExists("Cargo.toml")},
		{"react", checkPackageJSONDep("react")},
		{"typescript", checkFileExists("tsconfig.json")},
		{"python", checkAnyFileExists("pyproject.toml", "requirements.txt", "setup.py")},
		{"ruby", checkFileExists("Gemfile")},
		{"java", checkAnyFileExists("pom.xml", "build.gradle")},
		{"kotlin", checkFileExists("build.gradle.kts")},
		{"swift", checkSwift},
		{"csharp", checkCSharp},
		{"php", checkFileExists("composer.json")},
		{"docker", checkAnyFileExists("Dockerfile", "docker-compose.yml", "docker-compose.yaml")},
	}

	for _, c := range checks {
		if ok, evidence := c.check(workspace); ok {
			stacks = append(stacks, StackInfo{
				Name:       c.name,
				Confidence: 1.0,
				Evidence:   evidence,
			})
		}
	}

	return stacks
}

func checkFileExists(name string) func(string) (bool, string) {
	return func(workspace string) (bool, string) {
		path := filepath.Join(workspace, name)
		if _, err := os.Stat(path); err == nil {
			return true, name + " found"
		}
		return false, ""
	}
}

func checkAnyFileExists(names ...string) func(string) (bool, string) {
	return func(workspace string) (bool, string) {
		for _, name := range names {
			path := filepath.Join(workspace, name)
			if _, err := os.Stat(path); err == nil {
				return true, name + " found"
			}
		}
		return false, ""
	}
}

func checkPackageJSONDep(dep string) func(string) (bool, string) {
	return func(workspace string) (bool, string) {
		path := filepath.Join(workspace, "package.json")
		data, err := os.ReadFile(path)
		if err != nil {
			return false, ""
		}
		var pkg struct {
			Dependencies    map[string]string `json:"dependencies"`
			DevDependencies map[string]string `json:"devDependencies"`
		}
		if err := json.Unmarshal(data, &pkg); err != nil {
			return false, ""
		}
		if _, ok := pkg.Dependencies[dep]; ok {
			return true, dep + " in dependencies"
		}
		if _, ok := pkg.DevDependencies[dep]; ok {
			return true, dep + " in devDependencies"
		}
		return false, ""
	}
}

func checkSwift(workspace string) (bool, string) {
	if _, err := os.Stat(filepath.Join(workspace, "Package.swift")); err == nil {
		return true, "Package.swift found"
	}
	matches, _ := filepath.Glob(filepath.Join(workspace, "*.xcodeproj"))
	if len(matches) > 0 {
		return true, ".xcodeproj found"
	}
	return false, ""
}

func checkCSharp(workspace string) (bool, string) {
	matches, _ := filepath.Glob(filepath.Join(workspace, "*.csproj"))
	if len(matches) > 0 {
		return true, ".csproj found"
	}
	matches, _ = filepath.Glob(filepath.Join(workspace, "*.sln"))
	if len(matches) > 0 {
		return true, ".sln found"
	}
	return false, ""
}
