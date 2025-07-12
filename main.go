package main

import (
	"context"
	"fmt"
	"restic-backup-system/logic"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func main() {
	endpoint := "localhost:9000"
	accessKey := "miniobaga"
	secretKey := "miniobaga"

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false,
	})
	if err != nil {
		panic(err)
	}

	// chunks bucket
	err = minioClient.MakeBucket(context.Background(), "chunks", minio.MakeBucketOptions{})
	if err != nil {
		fmt.Println("chunks bucket exists or error:", err)
	}

	// snapshots bucket
	err = minioClient.MakeBucket(context.Background(), "snapshots", minio.MakeBucketOptions{})
	if err != nil {
		fmt.Println("snapshots bucket exists or error:", err)
	}

	fmt.Println("Buckets created or already exist")

	err = logic.Backup(minioClient, "./my-folder")
	if err != nil {
		panic(err)
	}

	// restoring a snapshot
	err = logic.Restore(minioClient, "snap-1752325410.json", "./restored")
	if err != nil {
		panic(err)
	}
	fmt.Println("Restored")
}