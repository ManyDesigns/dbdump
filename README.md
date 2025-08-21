# Dbdump

This tool is a wrapper for PostgreSQL's native `pg_dump` and `pg_restore` commands (in the future, others DB engines will
be added) and basically add the option to upload a DB dump to an S3 bucket.

The **main goal** of this tool, is to allow Developers or whoever need it, to be able to generate a DB dump in a remote
system, which does have a working connection to the Internet network, but, users are denied to download/upload files from it
to their local machines directly.

The secondary goal is to have a standard dump format (a custom compressed archive) for everyone. You can not change
(add/remove) parameters.

We use `pg_dump` command here to perform the dump. Followin is a list and meaning of the parameters used:
- `-h` -> database server host or socket directory
- `-p` -> database server port number
- `-U` -> connect as specified database user
- `-d` -> database to dump
- `-O` -> prevents `ALTER OWNER` commands in the dump file.
- `-f` -> output file or directory name
- `-F c` -> output file format (custom, directory, tar, plain text(default)). We are using the custom format here.
- `-c` -> clean(drop) database objects before recreating

The third goal, is to execute the dump once and have it at disposal for others in AWS S3.

## Download the tool and install on your local machine or the remote server

Go on the release page of the repository: https://github.com/ManyDesigns/dbdump/releases and select the correct architecture
for your Operating System.

Download it, decompress and save it in your Path.

## How to use this cli

### DUMP

To create a new dump you have to execute the following:

1. Create a new environment variable `export AWS_BUCKET=<name-of-the-bucket>` where the dumps will be uploaded.
2. The default AWS Region, if not specified is set to `eu-south-1`. So if the bucket has been created in another AWS
region, you have to create a new env variable: `export AWS_REGION=<your-region>`
3. Launch the command:

```bash
dbdump dump -W <your db password> \
      -U <your db user> \
			-h <your DB server hostname/IP>
			-p <your DB server port>
			-d <the db name you want to dump>
```
For more options have a look at the help menu `dbdump dump --help`:

```bash
dbdump dump --help
Usage of dump:
-U string
PostgreSQL user (default "postgres")
-W string
PostgreSQL password
-a    Back up all non-template databases
-d value
Specify a database to back up (can be used multiple times)
-e string
Specify the target environment (e.g. prod|staging)getenv) (default "staging")
-h string
PostgreSQL host (default "127.0.0.1")
-l    Avoid uploading the dump to a S3 bucket. Default is 'false'
-p int
PostgreSQL port (default 5432)
-t string
The type of database to back up (e.g., postgres) (default "postgres")
```

### Restore

The `restore` sub-command works both using an S3 URI or a local file path

It works the same way as the `dump` sub-command but it has an option to speed-up the restore passing the number of CPU
cores you want to use.

```bash
dbdump restore --help
Usage of restore:
-U string
PostgreSQL user (default "postgres")
-W string
PostgreSQL password
-d string
Name of the DB to restore (default "eclaim")
-f string
The absolute path of the dump file to restore the DB from or the S3 URI
-h string
PostgreSQL host (default "127.0.0.1")
-n int
Number of parallel processes (1 per CPU Core) to use (default 2)
-p int
PostgreSQL port (default 5432)
-s    Download the dump from AWS S3 (default true)
-t string
The type of database to restore (postgres,mysql,etc...) (default "postgres")
```
