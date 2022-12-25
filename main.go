package main

import (
	"embed"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

//go:embed .golangci.yml
var linting embed.FS

//go:embed pre-commit.sh
var hook embed.FS

//go:embed .gitignore
var gitignore embed.FS

func main() {
	projectName := flag.String("d", "new_project", "project name")
	flag.Parse()

	if err := mkdir(*projectName); err != nil {
		log.Fatal(err)
	}

	err := os.Chdir(*projectName)
	if err != nil {
		log.Fatal(err)
	}

	if err = initRepo(); err != nil {
		log.Fatal(err)
	}

	if err = goModInit(*projectName); err != nil {
		log.Fatal(err)
	}

	if err = createLintConfig(); err != nil {
		log.Fatal(err)
	}

	if err = createPreCommitHook(); err != nil {
		log.Fatal(err)
	}

	if err = createGitIgnore(); err != nil {
		log.Fatal(err)
	}
}

func mkdir(name string) error {
	if _, err := os.Stat(name); err == nil {
		return fmt.Errorf("project already exist: %w", err)
	}

	pwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	path := filepath.Join(pwd, name)
	if err = os.Mkdir(path, os.ModePerm); err != nil {
		return fmt.Errorf("error creating folder: %w", err)
	}

	return nil
}

func initRepo() error {
	cmd := exec.Command("git", "init")
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error git init: %w", err)
	}

	return nil
}

func goModInit(name string) error {
	cmd := exec.Command("go", "mod", "init", name)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error init go module: %w", err)
	}

	return nil
}

func createLintConfig() error {
	file, err := os.Create(".golintci.yml")
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()

	// Write the contents of the golinterciYml string to the file
	bytes, err := linting.ReadFile(".golangci.yml")
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}

	_, err = file.Write(bytes)
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}

	return nil
}

func createGitIgnore() error {
	file, err := os.Create(".gitignore")
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()

	// Write the contents of the golinterciYml string to the file
	bytes, err := gitignore.ReadFile(".gitignore")
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}

	_, err = file.Write(bytes)
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}

	return nil
}

func createPreCommitHook() error {
	err := os.Chdir(".git/hooks")
	if err != nil {
		return fmt.Errorf("error cd into git: %w", err)
	}
	file, err := os.Create("pre-commit")
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()

	bytes, err := hook.ReadFile("pre-commit.sh")
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}

	_, err = file.Write(bytes)
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}

	return nil
}
