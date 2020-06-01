package common

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// WriteTarArchive creates Tar archive of specified path
// using specified writer
func WriteTarArchive(srcDirPath string, compressWriter io.Writer) error {
	tarWriter := tar.NewWriter(compressWriter)
	defer tarWriter.Close()

	err := filepath.Walk(srcDirPath, func(filePath string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		tarHeader, err := tar.FileInfoHeader(fileInfo, fileInfo.Name())
		if err != nil {
			return err
		}

		tarHeader.Name, err = filepath.Rel(srcDirPath, filePath)
		if err != nil {
			return err
		}

		if err := tarWriter.WriteHeader(tarHeader); err != nil {
			return err
		}

		if fileInfo.Mode().IsRegular() {
			if err := writeFileToWriter(filePath, tarWriter); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

// WriteTgzArchive creates TGZ archive of specified path
func WriteTgzArchive(srcDirPath string, destFilePath string) error {
	destFile, err := os.Create(destFilePath)
	if err != nil {
		return fmt.Errorf("Failed to create result TGZ file %s: %s", destFilePath, err)
	}

	gzipWriter := gzip.NewWriter(destFile)
	defer gzipWriter.Close()

	err = WriteTarArchive(srcDirPath, gzipWriter)
	if err != nil {
		return err
	}

	return nil
}

// CompressGzip compresses specified file  with gzip.BestCompression level
func CompressGzip(srcFilePath string, destFilePath string) error {
	var err error

	// src file reader
	srcFileReader, err := os.Open(srcFilePath)
	if err != nil {
		return fmt.Errorf("Failed to open source file %s: %s", srcFilePath, err)
	}
	defer srcFileReader.Close()

	// dest file writer
	destFile, err := os.Create(destFilePath)
	if err != nil {
		return fmt.Errorf("Failed to create result GZIP file %s: %s", destFilePath, err)
	}
	defer destFile.Close()

	// dest file GZIP writer
	gzipWriter, err := gzip.NewWriterLevel(destFile, gzip.BestCompression)
	if err != nil {
		return fmt.Errorf("Failed to create GZIP writer %s: %s", destFilePath, err)
	}
	defer gzipWriter.Flush()

	// compressing itself
	if _, err := io.Copy(gzipWriter, srcFileReader); err != nil {
		return err
	}

	return nil
}
