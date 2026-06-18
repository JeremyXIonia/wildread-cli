package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

const (
	DefaultDataDirName = ".cli-read"
	DefaultDBFileName  = "novel-reader.db"
	DefaultBookDirName = ".book"
)

type Paths struct {
	DataDir        string
	DBPath         string
	DefaultBookDir string
	TempBookDir    string
}

func ResolvePaths(dataDirFlag, dbFlag, tempDirFlag string) (Paths, error) {
	dataDir := dataDirFlag
	if strings.TrimSpace(dataDir) == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return Paths{}, err
		}
		dataDir = filepath.Join(home, DefaultDataDirName)
	}

	dataDir, err := NormalizePath(dataDir)
	if err != nil {
		return Paths{}, err
	}

	dbPath := filepath.Join(dataDir, DefaultDBFileName)
	if strings.TrimSpace(dbFlag) != "" {
		dbPath, err = NormalizePath(dbFlag)
		if err != nil {
			return Paths{}, err
		}
	}

	var tempDir string
	if strings.TrimSpace(tempDirFlag) != "" {
		tempDir, err = NormalizePath(tempDirFlag)
		if err != nil {
			return Paths{}, err
		}
	}

	return Paths{
		DataDir:        dataDir,
		DBPath:         dbPath,
		DefaultBookDir: filepath.Join(dataDir, DefaultBookDirName),
		TempBookDir:    tempDir,
	}, nil
}

func NormalizePath(path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", errors.New("empty path")
	}
	if path == "~" || strings.HasPrefix(path, "~"+string(filepath.Separator)) || strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		if path == "~" {
			path = home
		} else {
			path = filepath.Join(home, path[2:])
		}
	}
	if !filepath.IsAbs(path) {
		abs, err := filepath.Abs(path)
		if err != nil {
			return "", err
		}
		path = abs
	}
	return filepath.Clean(path), nil
}
