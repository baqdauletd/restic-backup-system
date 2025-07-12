package logic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/minio/minio-go/v7"
)

func Restore(minioClient *minio.Client, snapshotName string, destDir string) error {
	// downlaod snapshot
	obj, err := minioClient.GetObject(context.Background(), bucketSnapshots, snapshotName, minio.GetObjectOptions{})
	if err != nil {
		return fmt.Errorf("error getting snapshot: %w", err)
	}
	defer obj.Close()

	var snapshot Snapshot
	err = json.NewDecoder(obj).Decode(&snapshot)
	if err != nil {
		return fmt.Errorf("error decoding snapshot JSON: %w", err)
	}

	// rebuild each file
	for _, file := range snapshot.Files {
		outputPath := filepath.Join(destDir, file.Path)
		fmt.Println("Restoring:", outputPath)

		// create directory if needed
		if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
			return err
		}

		outputFile, err := os.Create(outputPath)
		if err != nil {
			return err
		}
		defer outputFile.Close()

		for _, chunkID := range file.Chunks {
			chunkObj, err := minioClient.GetObject(context.Background(), bucketChunks, chunkID, minio.GetObjectOptions{})
			if err != nil {
				return fmt.Errorf("error getting chunk %s: %w", chunkID, err)
			}
			defer chunkObj.Close()

			// read encrypted+compressed data
			var encData bytes.Buffer
			_, err = io.Copy(&encData, chunkObj)
			if err != nil {
				return err
			}

			chunkData, err := DecryptAndDecompress(encData.Bytes())
			if err != nil {
				return err
			}

			_, err = outputFile.Write(chunkData)
			if err != nil {
				return err
			}
		}
	}

	fmt.Println("Restored")
	return nil
}
