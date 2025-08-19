package dump

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// PostgresDumper type definition
type PostgresDumper struct {
	host        string
	port        int
	user        string
	password    string
	numCPUCores int
	environment string
	localdump   bool // in Go the default value for a bool type is `false`.
}

func NewPostgresDumper(host string, port int, user, password, environment string, localdump bool) *PostgresDumper {
	return &PostgresDumper{
		host:        host,
		port:        port,
		user:        user,
		password:    password,
		environment: environment,
		localdump:   localdump,
	}
}

type PostgresRestorer struct {
	host         string
	port         int
	user         string
	password     string
	numCPUCores  int
	fileDumpPath string
	dbName       string
}

func NewPostgresRestorer(host string, port, numCPUCores int, user, password, fileDumpPath, dbName string) *PostgresRestorer {
	return &PostgresRestorer{
		host:         host,
		port:         port,
		user:         user,
		password:     password,
		numCPUCores:  numCPUCores,
		fileDumpPath: fileDumpPath,
		dbName:       dbName,
	}
}

// Dump executes a dump of the DBs passed
func (p *PostgresDumper) Dump(dbName string, dumpFilename string) error {
	os.Setenv("PGPASSWORD", p.password)
	defer os.Unsetenv("PGPASSWORD")

	// We use `pg_dump` command here to perform the dump.
	// List and meaning of parameters used
	// `-h` -> database server host or socket directory
	// `-p` -> database server port number
	// `-U` -> connect as specified database user
	// `-d` -> database to dump
	// `-f` -> output file or directory name
	// `-F c` -> output file format (custom, directory, tar, plain text(default)). We are using the custom format here.
	// `-c` -> clean(drop) database objects before recreating
	cmd := exec.Command("pg_dump", "-h", p.host, "-p", fmt.Sprintf("%d", p.port), "-U", p.user, "-d", dbName, "-f", dumpFilename, "-F", "c", "-c")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if _, statErr := os.Stat(dumpFilename); statErr == nil {
			os.Remove(dumpFilename)
		}
		return fmt.Errorf("failed to dump '%s': %s", dbName, stderr.String())
	}
	return nil
}

func (p *PostgresDumper) ListDatabases() ([]string, error) {
	os.Setenv("PGPASSWORD", p.password)
	defer os.Unsetenv("PGPASSWORD")

	cmd := exec.Command("psql", "-h", p.host, "-p", fmt.Sprintf("%d", p.port), "-U", p.user, "-l", "-t", "-A")
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("could not list databases: %s", stderr.String())
	}

	var allDbs []string
	for _, line := range strings.Split(out.String(), "\n") {
		if len(line) > 0 {
			dbName := strings.TrimSpace(strings.Split(line, "|")[0])
			if dbName != "template0" && dbName != "template1" && dbName != "postgres" && dbName != "rdsadmin" {
				allDbs = append(allDbs, dbName)
			}
		}
	}
	return allDbs, nil
}

func (p *PostgresRestorer) Restore(dbName string, dumpFilename string) error {
	os.Setenv("PGPASSWORD", p.password)
	defer os.Unsetenv("PGPASSWORD")

	// We use `pg_restore` to restore the DBs
	// The new parameter here is `-j` which is used to speed up the restore process, passing the number of
	// parallel processes (1 per CPU Core) to use.
	cmd := exec.Command("pg_restore", "-h", p.host, "-p", fmt.Sprintf("%d", p.port), "-U", p.user, "-d", dbName, "-j", fmt.Sprintf("%d", p.numCPUCores), dumpFilename)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to restore '%s': '%s'", dbName, stderr.String())
	}
	return nil
}
