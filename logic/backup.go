package logic

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
)

const (
	chunkSize = 4 * 1024 * 1024 // 4MB
	bucketChunks = "chunks"
	bucketSnapshots = "snapshots"
)

var encryptionKey = []byte("0123456789abcdef0123456789abcdef") // 32 bytes (for the first try)

type FileEntry struct {
	Path   string   `json:"path"`
	Chunks []string `json:"chunks"`
}

type Snapshot struct {
	Timestamp time.Time   `json:"timestamp"`
	Files     []FileEntry `json:"files"`
}

var fileNames map[string]string

func Backup(minioClient *minio.Client, sourceDir string) error {
	var snapshot Snapshot
	snapshot.Timestamp = time.Now()
	var relPath string

	err := filepath.WalkDir(sourceDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}

		relPath, err = filepath.Rel(sourceDir, path)
		if err != nil{
			fmt.Println("Error reading directory")
			return err
		}
		fmt.Println("Backing up:", relPath)

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		var chunks []string
		buf := make([]byte, chunkSize)

		for {
			n, err := file.Read(buf)
			if err == io.EOF {
				break
			}
			if err != nil {
				return err
			}

			chunk := buf[:n]
			hash := sha256.Sum256(chunk)
			chunkID := hex.EncodeToString(hash[:])

			// eheck if chunk already exists
			found, _ := minioClient.StatObject(context.Background(), bucketChunks, chunkID, minio.StatObjectOptions{})
			if found.Size == 0 {
				// encrypt and compress
				encChunk, err := EncryptAndCompress(chunk)
				if err != nil {
					return err
				}

				// upload to S3
				_, err = minioClient.PutObject(
					context.Background(),
					bucketChunks,
					chunkID,
					encChunk,
					-1,
					minio.PutObjectOptions{},
				)
				if err != nil {
					return err
				}
			}

			chunks = append(chunks, chunkID)
		}

		snapshot.Files = append(snapshot.Files, FileEntry{
			Path:   relPath,
			Chunks: chunks,
		})

		// fileNames[relPath] = fmt.Sprintf("snap-%d.json", time.Now().Unix())

		return nil
	})

	if err != nil {
		return err
	}

	// save snapshots in json
	snapData, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return err
	}
	snapName := fmt.Sprintf("snap-%d.json", time.Now().Unix())
	_, err = minioClient.PutObject(
		context.Background(),
		bucketSnapshots,
		snapName,
		strings.NewReader(string(snapData)),
		int64(len(snapData)),
		minio.PutObjectOptions{ContentType: "application/json"},
	)
	if err != nil {
		return err
	}

	fmt.Println("Snapshot saved as", snapName)
	return nil
}
