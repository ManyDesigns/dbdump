package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"manydesigns/dbdump/pkg/dump"
)

const AppVersion = "0.0.2"

type dbList []string

func (d *dbList) String() string         { return fmt.Sprintf("%v", *d) }
func (d *dbList) Set(value string) error { *d = append(*d, value); return nil }

func main() {
	// Check if a valid command (dump or restore or -v) was provided.
	var appVersion = flag.Bool("v", false, "Print the version number and exit")
	flag.Parse()

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	} else if *appVersion {
		fmt.Printf("dbdump version %s\n", AppVersion)
		os.Exit(0)
	}

	// Set the first argument as the main command.
	switch os.Args[1] {
	case "dump":
		executeDumpCommand()
	case "restore":
		executeRestoreCommand()
	default:
		fmt.Printf("Error: unsupported command '%s'\n\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

// printUsage
func printUsage() {
	fmt.Println("Usage: dbdump <command> [arguments]")
	fmt.Println("\nAvailable commands:")
	fmt.Println("  dump       Dump one or more databases")
	fmt.Println("  restore    Restore a database from a dump file")
	fmt.Println("  -v         Print the version number and exit")
	fmt.Println("\nUse `dbdump <command> --help` for more information about specific command.")
}

// executeDumpCommand
func executeDumpCommand() {
	dumpCmd := flag.NewFlagSet("dump", flag.ExitOnError)
	// Flags to connect to the DB Server
	var pgHost, pgUser, pgPassword, dbType, targetEnv string
	var pgPort int
	var backupAll, localDump bool

	dumpCmd.StringVar(&pgHost, "h", "127.0.0.1", "PostgreSQL host")

	dumpCmd.IntVar(&pgPort, "p", 5432, "PostgreSQL port")

	dumpCmd.StringVar(&pgUser, "U", "postgres", "PostgreSQL user")

	dumpCmd.StringVar(&pgPassword, "W", "", "PostgreSQL password")

	var dbsToBackup dbList
	dumpCmd.BoolVar(&backupAll, "a", false, "Back up all non-template databases")

	dumpCmd.StringVar(&dbType, "t", "postgres", "The type of database to back up (e.g., postgres)")

	dumpCmd.Var(&dbsToBackup, "d", "Specify a database to back up (can be used multiple times)")

	dumpCmd.StringVar(&targetEnv, "e", "staging", "Specify the target environment (e.g. prod|staging)getenv)")

	dumpCmd.BoolVar(&localDump, "l", false, "Avoid uploading the dump to a S3 bucket. Default is 'false'")

	err := dumpCmd.Parse(os.Args[2:])
	if err != nil {
		return
	}

	// Check if the password for the PostgreSQL user has been passed
	if pgPassword == "" {
		fmt.Println("Error: -W flag is required.")
		os.Exit(1)
	}

	var dbdumper dump.Dumper
	switch dbType {
	case "postgres":
		dbdumper = dump.NewPostgresDumper(pgHost, pgPort, pgUser, pgPassword, targetEnv, localDump)
	default:
		fmt.Println("Error: Unsupported database type:", dbType)
		os.Exit(1)
	}

	uploader, err := dump.NewS3Uploader(localDump)
	if err != nil {
		fmt.Println("Error initializing S3 uploader:", err)
		os.Exit(1)
	}

	var finalDbList []string
	if backupAll {
		finalDbList, err = dbdumper.ListDatabases()
		if err != nil {
			fmt.Println("Error listing databases:", err)
			os.Exit(1)
		}
	} else if len(dbsToBackup) > 0 {
		finalDbList = dbsToBackup
	} else {
		fmt.Println("Error: No databases specified. Use '-d' or '-a'.")
		os.Exit(1)
	}

	// Create a WaitGroup
	var wg sync.WaitGroup

	fmt.Println("\nInitializing backup processs in background...")
	for _, dbName := range finalDbList {
		fmt.Printf("-> Spawning backup task for '%s'.\n", dbName)
		wg.Add(1)
		go processDatabaseDump(&wg, targetEnv, dbName, dbdumper, uploader, localDump)
	}
	fmt.Println("\nAll backup processes have been started. Waiting for them to complete....")
	wg.Wait()
	fmt.Println("Check the 'backup_log_*.log' files for progress and results.")
}

// executeRestoreCommand
func executeRestoreCommand() {
	restoreCmd := flag.NewFlagSet("restore", flag.ExitOnError)
	// Flags to connect to the DB Server and load the dump file
	// `dumpFile` here can have two values, if the `s3Dump` flag is true (default) the value is the S3 bucket URI
	// otherwise is the local system path.
	var pgHost, pgUser, pgPassword, dbName, dbType, dumpFile string
	var pgPort, numCPUCores int
	var s3Dump bool

	restoreCmd.StringVar(&pgHost, "h", "127.0.0.1", "PostgreSQL host")

	restoreCmd.StringVar(&pgUser, "U", "postgres", "PostgreSQL user")

	restoreCmd.StringVar(&pgPassword, "W", "", "PostgreSQL password")

	restoreCmd.StringVar(&dbName, "d", "eclaim", "Name of the DB to restore")

	restoreCmd.StringVar(&dbType, "t", "postgres", "The type of database to restore (postgres,mysql,etc...)")

	restoreCmd.StringVar(&dumpFile, "f", "", "The absolute path of the dump file to restore the DB from or the S3 URI")

	restoreCmd.IntVar(&pgPort, "p", 5432, "PostgreSQL port")

	restoreCmd.IntVar(&numCPUCores, "n", 2, "Number of parallel processes (1 per CPU Core) to use")

	restoreCmd.BoolVar(&s3Dump, "s", true, "Download the dump from AWS S3")

	err := restoreCmd.Parse(os.Args[2:])
	if err != nil {
		return
	}

	if pgPassword == "" || dbName == "" || dumpFile == "" {
		fmt.Println("Error: -W, -d, -f flags are required for restore.")
		restoreCmd.Usage()
		os.Exit(1)
	}

	downloader, err := dump.NewS3Downloader(s3Dump)
	if err != nil {
		fmt.Println("Error initializing S3 downloader:", err)
		os.Exit(1)
	}

	var dbrestorer dump.Restorer = dump.NewPostgresRestorer(pgHost, pgPort, numCPUCores, pgUser, pgPassword, dumpFile, dbName)

	var wg sync.WaitGroup

	fmt.Println("\nInitalizing restore process in background...")
	fmt.Printf("-> Spawning restore task for '%s'.\n", *&dbName)
	wg.Add(1)
	go processDatabaseRestore(&wg, *&dbName, dbrestorer, dumpFile, s3Dump, downloader)
	fmt.Println("\nAll restore processes have been started. Waiting for them to complete.")
	wg.Wait()
	fmt.Println("Check the 'restore_log_*.log' file for progress and results.")
}

// processDatabaseRestore is used to restore a DB using the dump file passed as parameter.
// It can also take the number of processes to execute in parallel based on the number of available CPUs core
func processDatabaseRestore(wg *sync.WaitGroup, dbName string, dbrestorer dump.Restorer, dumpFileName string, s3Download bool, downloader dump.Downloader) {
	defer wg.Done()

	logFilename := fmt.Sprintf("restore_log_%s_%s.log", dbName, time.Now().UTC().Format("20060102_150405"))
	logFile, err := os.OpenFile(logFilename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)

	if err != nil {
		log.Printf("Fatal: Could not open log file '%s': %v", logFilename, err)
	}

	defer func(logFile *os.File) {
		err := logFile.Close()
		if err != nil {
			log.Printf("Error closing log file: %v", err)
		}
	}(logFile)

	logger := log.New(logFile, "", log.LstdFlags)
	logger.Printf("Starting restore process for '%s' using '%s' dump...", dbName, dumpFileName)

	// We need to manage the restore based on the remote S3 bucket or the local file path
	var local_dump_file string
	if s3Download {
		local_dump_file = filepath.Base(dumpFileName)
		logger.Println(" Downloading dump from S3...")
		s3Uri, err := downloader.Download(dumpFileName, local_dump_file)
		if err != nil {
			logger.Printf("Error downloading dump from S3: %v", err)
			return
		}
		logger.Println("Download successful.")
		logger.Printf("File downloaded to: %s", s3Uri)
	}

	logger.Println("Restoring database...")
	if s3Download {
		dumpFileName = local_dump_file
	} else {
		dumpFileName = *&dumpFileName
	}
	if err := dbrestorer.Restore(dbName, dumpFileName); err != nil {
		logger.Printf("Error during restoring the DB '%s': %v", dbName, err)
		return
	}
	logger.Println("DB(s) restored succesful.")
}

func processDatabaseDump(wg *sync.WaitGroup, tgEnv string, dbName string, dbdumper dump.Dumper, uploader dump.Uploader, localDump bool) {
	defer wg.Done()
	logFilename := fmt.Sprintf("backup_log_%s_%s_%s.log", tgEnv, dbName, time.Now().UTC().Format("20060102_150405"))
	logFile, err := os.OpenFile(logFilename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)

	if err != nil {
		log.Printf("Fatal: Could not open log file %s: %v", logFilename, err)
		return
	}
	defer func(logFile *os.File) {
		err := logFile.Close()
		if err != nil {
			log.Printf("Error closing log file: %v", err)
		}
	}(logFile)

	logger := log.New(logFile, "", log.LstdFlags)
	logger.Printf("Starting dump process for '%s'...", dbName)

	backupFilename := fmt.Sprintf("%s_%s_%s.dump", tgEnv, dbName, time.Now().UTC().Format("20060102_150405"))

	logger.Println("1. Dumping database...")
	if err := dbdumper.Dump(dbName, backupFilename); err != nil {
		logger.Printf("Error during dump: %v", err)
		return
	}
	logger.Println("Dump successful.")

	switch localDump {
	case false:
		logger.Println("2. Uploading to S3...")
		s3Uri, err := uploader.Upload(backupFilename, backupFilename)
		if err != nil {
			logger.Printf("Error during upload: %v", err)
			logger.Printf("The local file '%s' has been kept for manual inspection.", backupFilename)
			return
		}
		logger.Println("Upload successful.")
		logger.Printf("3. Dumpm complete. File uploaded to: %s", s3Uri)
	case true:
		logger.Println("Skipping S3 uploading as requested.")
	default:
		fmt.Println("No target location specified:", localDump)
		os.Exit(1)
	}

	// if the user requested to have the dump locally, we do not remove it
	if *&localDump {
		logger.Printf("Process finished successfully. Your dump file is: %s", backupFilename)
	} else {
		if err := os.Remove(backupFilename); err != nil {
			logger.Printf("Warning: Could not delete local file '%s'. Error: %v", backupFilename, err)
		} else {
			logger.Printf("Local file '%s' has been deleted.", backupFilename)
			logger.Println("Process finished successfully.")
		}
	}
}
