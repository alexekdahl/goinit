package main

import (
	"embed"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
)

//go:embed templates/*
var templatesFS embed.FS

const (
	DefaultProjectName = "new_project"
)

func main() {
	cmd := exec.Command("go", "version")
	if _, err := cmd.CombinedOutput(); err != nil {
		fmt.Printf("Go is not installed: %s\n", err)
		log.Fatal(err)
	}

	projectName := flag.String("d", DefaultProjectName, "project name")
	flag.Parse()

	if err := mkdir(*projectName); err != nil {
		log.Fatal(err)
	}

	if err := createProjectFiles(*projectName); err != nil {
		log.Fatal(err)
	}
}

func mkdir(name string) error {
	if _, err := os.Stat(name); err == nil {
		return fmt.Errorf("project already exists: %w", err)
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

// createProjectFiles creates the necessary files and directories for the project.
func createProjectFiles(projectName string) error {
	if err := os.Chdir(projectName); err != nil {
		return fmt.Errorf("error changing to project directory: %w", err)
	}

	if err := initRepo(); err != nil {
		return fmt.Errorf("error initializing repository: %w", err)
	}

	if err := goModInit(projectName); err != nil {
		return fmt.Errorf("error initializing Go module: %w", err)
	}

	if err := createFile(".golintci.yml", templatesFS, "templates/.golangci.yml"); err != nil {
		return fmt.Errorf("error creating linting configuration file: %w", err)
	}

	if err := createFile(".gitignore", templatesFS, "templates/.gitignore"); err != nil {
		return fmt.Errorf("error creating .gitignore file: %w", err)
	}

	if err := createFile("Makefile", templatesFS, "templates/Makefile"); err != nil {
		return fmt.Errorf("error creating .gitignore file: %w", err)
	}

	if err := createScripts(); err != nil {
		return fmt.Errorf("error creating scripts: %w", err)
	}

	if err := createPreCommitHook(); err != nil {
		return fmt.Errorf("error creating pre-commit hook: %w", err)
	}

	return nil
}

// initRepo initializes a new Git repository in the current directory.
func initRepo() error {
	cmd := exec.Command("git", "init")
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error initializing repository: %w", err)
	}

	return nil
}

// goModInit initializes a new Go module with the given name in the current directory.
func goModInit(name string) error {
	alias := getAlias()
	projectName := alias + name
	cmd := exec.Command("go", "mod", "init", projectName)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error initializing Go module: %w", err)
	}

	return nil
}

func getAlias() string {
	const (
		alias         = "project/"
		sshConfigPath = ".ssh/config"
		req           = `Host github\.com\n\s+User (?P<user>\w+)`
	)

	home, err := os.UserHomeDir()
	if err != nil {
		return alias
	}

	path := path.Join(home, sshConfigPath)

	input, err := readFile(path)
	if err != nil {
		return alias
	}

	re := regexp.MustCompile(req)
	match := re.FindStringSubmatch(input)

	if len(match) < 2 {
		return alias
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

// createFile creates a new file with the given name and writes the contents of the specified
// embedded file to it.
func createFile(name string, fs embed.FS, filePath string) error {
	file, err := os.Create(name)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()

	// Write the contents of the embedded file to the new file
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

// createPreCommitHook creates a pre-commit hook for the repository.
func createPreCommitHook() error {
	if err := os.Chdir(".git/hooks"); err != nil {
		return fmt.Errorf("error changing to .git/hooks directory: %w", err)
	}

	file, err := os.Create("pre-commit")
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()

	// Write the contents of the preCommitHook string to the file
	bytes, err := templatesFS.ReadFile("templates/scripts/pre-commit")
	if err != nil {
		return fmt.Errorf("error reading embedded file: %w", err)
	}

	_, err = file.Write(bytes)
	if err != nil {
		return fmt.Errorf("error writing to file: %w", err)
	}

	// Make the file executable
	if err = os.Chmod("pre-commit", 0o700); err != nil {
		return fmt.Errorf("error making file executable: %w", err)
	}

	return nil
}

func createScripts() error {
	if err := mkdir("scripts"); err != nil {
		return err
	}

	if err := createFile("scripts/pre-commit", templatesFS, "templates/scripts/pre-commit"); err != nil {
		return fmt.Errorf("error creating pre-commit file: %w", err)
	}

	if err := createFile("scripts/setup.sh", templatesFS, "templates/scripts/setup.sh"); err != nil {
		return fmt.Errorf("error creating setup file: %w", err)
	}
	// Make the file executable
	if err := os.Chmod("scripts/setup.sh", 0o700); err != nil {
		return fmt.Errorf("error making file executable: %w", err)
	}

	if err := createFile("scripts/cibuild.sh", templatesFS, "templates/scripts/cibuild.sh"); err != nil {
		return fmt.Errorf("error creating cibuild file: %w", err)
	}
	// Make the file executable
	if err := os.Chmod("scripts/cibuild.sh", 0o700); err != nil {
		return fmt.Errorf("error making file executable: %w", err)
	}

	return nil
}
