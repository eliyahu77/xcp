package backends

import (
	"fmt"
	"github.com/nuclio/logger"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type FileSearcher struct {
	SourcePath string
	TargetPath string
	Since      time.Time
	MinSize    int64
	MaxSize    int64
	Filter     string
	Recursive  bool
	Hidden     bool
}

type CopyTask struct {
	Source    *PathParams
	Target    *PathParams
	Since     time.Time
	MinSize   int64
	MaxSize   int64
	Filter    string
	Recursive bool
	Hidden    bool
}

type FileDetails struct {
	Key   string
	Mtime time.Time
	Size  int64
}

type ListSummary struct {
	TotalFiles int
	TotalBytes int64
}

type PathParams struct {
	Kind     string `json:"kind"`
	Endpoint string `json:"endpoint,omitempty"`
	Bucket   string `json:"bucket,omitempty"`
	Path     string `json:"path"`
	Tag      string `json:"tag,omitempty"`
	Secure   bool   `json:"secure,omitempty"`
	UserKey  string `json:"userKey,omitempty"`
	Secret   string `json:"secret,omitempty"`
	Token    string `json:"token,omitempty"`
}

type FSClient interface {
	ListDir(fileChan chan *FileDetails, task *CopyTask, summary *ListSummary) error
	Reader(path string) (io.ReadCloser, error)
	Writer(path string) (io.WriteCloser, error)
}

func GetNewClient(logger logger.Logger, params *PathParams) (FSClient, error) {
	switch strings.ToLower(params.Kind) {
	case "v3io":
		return NewV3ioClient(logger, params)
	case "s3":
		return NewS3Client(logger, params)
	case "", "file":
		return NewLocalClient(logger, params)
	default:
		return nil, fmt.Errorf("Unknown backend %s use s3, v3io or local", params.Kind)
	}
}

func ValidFSTarget(filePath string) error {
	// Verify if destination already exists.
	st, err := os.Stat(filePath)
	if err == nil {
		// If the destination exists and is a directory.
		if st.IsDir() {
			return fmt.Errorf("fileName %s is a directory.", filePath)
		}
	}

	// Proceed if file does not exist. return for all other errors.
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	}

	// Extract top level directory.
	objectDir, _ := filepath.Split(filePath)
	if objectDir != "" {
		// Create any missing top level directories.
		if err := os.MkdirAll(objectDir, 0700); err != nil {
			return err
		}
	}

	return nil
}

func defaultFromEnv(param string, envvar string) string {
	if param == "" {
		param = os.Getenv(envvar)
	}
	return param
}