# Go Project

This is a Go project that creates a new project with the specified name. The project includes a linting configuration file, a pre-commit hook, and a .gitignore file.

## Installation
To install `goinit`, you can clone this repository and build the binary from source:

```sh
git clone https://github.com/AlexEkdahl/goinit.git
cd goinit
go build -o goinit .
```
Or build it using the makefile
```bash
make build
```

## Usage
To use this program, run the following command:

```bash
goinit  -d [project_name]
```
Replace `[project_name]` with the desired name for the new project.
