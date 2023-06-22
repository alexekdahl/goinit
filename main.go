package main

import (
	"embed"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
)

//go:embed templates/*
var templatesFS embed.FS

const (
	DefaultProjectName      = "new_project"
	GolintciTemplate        = "templates/.golangci.yml"
	GoreleaserTemplate      = "templates/.goreleaser.yml"
	GitignoreTemplate       = "templates/.gitignore"
	MakefileTemplate        = "templates/Makefile"
	ReleaserTemplate        = "templates/releaser.yml"
	PreCommitHookTemplate   = "templates/scripts/pre-commit"
	PreCommitScriptTemplate = "templates/scripts/pre-commit"
	SetupScriptTemplate     = "templates/scripts/setup.sh"
	CIBuildScriptTemplate   = "templates/scripts/cibuild.sh"
	GolintciFile            = ".golintci.yml"
	GoreleaserFile          = ".goreleaser.yml"
	GitignoreFile           = ".gitignore"
	GithubDir               = ".github"
	WorkflowsDir            = ".github/workflows"
	ReleaserFile            = ".github/workflows/releaser.yml"
	GitHooksDir             = ".git/hooks"
	ScriptsDir              = "scripts"
	PreCommitScriptFile     = "scripts/pre-commit"
	SetupScriptFile         = "scripts/setup.sh"
	CIBuildScriptFile       = "scripts/cibuild.sh"
	PreCommitHookFile       = "pre-commit"
	Makefile                = "Makefile"
	SSHConfigDir            = ".ssh"
	SSHConfigFile           = ".ssh/config"
	DefaultAlias            = "project/"
	RegexpPattern           = `Host github\.com\n\s+User (?P<user>\w+)`
)

func main() {
	if !isGoInstalled() {
		log.Fatal("Go is not installed.")
	}

	projectName := flag.String("d", DefaultProjectName, "project name")
	flag.Parse()

	if err := mkdir(*projectName); err != nil {
		log.Fatal("Error creating directory: ", err)
	}

	if err := createProjectFiles(*projectName); err != nil {
		log.Fatal("Error creating project files: ", err)
	}
}

func isGoInstalled() bool {
	_, err := exec.Command("go", "version").CombinedOutput()
	return err == nil
}

func mkdir(name string) error {
	if _, err := os.Stat(name); err == nil {
		return fmt.Errorf("folder already exists: %w", err)
	}

	pwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current working directory: %w", err)
	}

	path := filepath.Join(pwd, name)
	if err = os.Mkdir(path, os.ModePerm); err != nil {
		return fmt.Errorf("error creating folder: %w", err)
	}

	return nil
}

func createProjectFiles(projectName string) error {
	filesToCreate := []struct {
		Name     string
		Template string
	}{
		{GolintciFile, GolintciTemplate},
		{GoreleaserFile, GoreleaserTemplate},
		{GitignoreFile, GitignoreTemplate},
		{Makefile, MakefileTemplate},
	}

	if err := os.Chdir(projectName); err != nil {
		return fmt.Errorf("error changing to project directory: %w", err)
	}

	if err := runCommand("git", "init"); err != nil {
		return fmt.Errorf("error initializing repository: %w", err)
	}

	if err := goModInit(projectName); err != nil {
		return fmt.Errorf("error initializing Go module: %w", err)
	}

	for _, file := range filesToCreate {
		if err := createFile(file.Name, templatesFS, file.Template); err != nil {
			return fmt.Errorf("error creating %s: %w", file.Name, err)
		}
	}

	if err := createScripts(); err != nil {
		return fmt.Errorf("error creating scripts: %w", err)
	}

	if err := createGithubAction(); err != nil {
		return fmt.Errorf("error creating github actions: %w", err)
	}

	if err := createPreCommitHook(); err != nil {
		return fmt.Errorf("error creating pre-commit hook: %w", err)
	}

	return nil
}

func runCommand(name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	return cmd.Run()
}

func goModInit(name string) error {
	alias := getAlias()
	projectName := alias + name
	return runCommand("go", "mod", "init", projectName)
}

func getAlias() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return DefaultAlias
	}

	path := filepath.Join(home, SSHConfigFile)

	input, err := readFile(path)
	if err != nil {
		return DefaultAlias
	}

	re := regexp.MustCompile(RegexpPattern)
	match := re.FindStringSubmatch(input)

	if len(match) < 2 {
		return DefaultAlias
	}

	return fmt.Sprintf("github.com/%s/", match[1])
}

func readFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	bytes := make([]byte, 1024)
	n, err := file.Read(bytes)
	if err != nil {
		return "", err
	}

	return string(bytes[:n]), nil
}

func createFile(name string, fs embed.FS, filePath string) error {
	file, err := os.Create(name)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()

	bytes, err := fs.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error reading embedded file: %w", err)
	}

	_, err = file.Write(bytes)
	if err != nil {
		return fmt.Errorf("error writing to file: %w", err)
	}

	return nil
}

func createPreCommitHook() error {
	if err := os.Chdir(GitHooksDir); err != nil {
		return fmt.Errorf("error changing to %s directory: %w", GitHooksDir, err)
	}

	if err := createExecutableFile(PreCommitHookFile, templatesFS, PreCommitHookTemplate); err != nil {
		return fmt.Errorf("error creating %s: %w", PreCommitHookFile, err)
	}

	return nil
}

func createGithubAction() error {
	dirsToCreate := []string{GithubDir, WorkflowsDir}

	for _, dir := range dirsToCreate {
		if err := mkdir(dir); err != nil {
			return fmt.Errorf("error creating %s: %w", dir, err)
		}
	}

	if err := createFile(ReleaserFile, templatesFS, ReleaserTemplate); err != nil {
		return fmt.Errorf("error creating %s: %w", ReleaserFile, err)
	}

	return nil
}

func createScripts() error {
	if err := mkdir(ScriptsDir); err != nil {
		return err
	}

	filesToCreate := []struct {
		Name     string
		Template string
	}{
		{PreCommitScriptFile, PreCommitScriptTemplate},
		{SetupScriptFile, SetupScriptTemplate},
		{CIBuildScriptFile, CIBuildScriptTemplate},
	}

	for _, file := range filesToCreate {
		if err := createExecutableFile(file.Name, templatesFS, file.Template); err != nil {
			return fmt.Errorf("error creating %s: %w", file.Name, err)
		}
	}

	return nil
}

func createExecutableFile(name string, fs embed.FS, filePath string) error {
	if err := createFile(name, fs, filePath); err != nil {
		return err
	}

	// Make the file executable
	if err := os.Chmod(name, 0o700); err != nil {
		return fmt.Errorf("error making %s executable: %w", name, err)
	}

	return nil
}
