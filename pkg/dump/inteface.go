package dump

// Dumper defines the interface for a database dump implementation.
type Dumper interface {
	Dump(dbName string, dumpFilename string) error
	ListDatabases() ([]string, error)
}

// Restorer defines the interface for a databse restore implementation.
type Restorer interface {
	Restore(dbName string, dumpFilename string) error
}

// Uploader defines the interface for a cloud storage implementaiotn.
type Uploader interface {
	Upload(localPath string, remotePath string) (string, error)
}

// Downloader defines the interace for a cloud storage implementation.
type Downloader interface {
	Download(remotePath string, localPath string) (string, error)
}
