package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"restic-backup-system/logic"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func main() {
	mode := flag.String("mode", "", "backup or restore")
	backup_folder := flag.String("backup_folder", "./my-folder", "write a folder to be stored in backup (for backup mode)")
	restore_folder := flag.String("restore_folder", "./restores", "write a folder where all the files will be restored (for restore mode)")
	flag.Parse()
	
	endpoint := "localhost:9000"
	accessKey := "miniobaga"
	secretKey := "miniobaga"

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false,
	})
	if err != nil {
		fmt.Println("Error starting MINIO:", err)
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

	switch *mode{
	case "backup":
		//storing 
		err = logic.Backup(minioClient, *backup_folder)
		if err != nil {
			panic(err)
		}
	case "restore":
		// restoring a snapshot
		err = logic.Restore(minioClient, "snap-1752325410.json", *restore_folder)
		if err != nil {
			panic(err)
		}
		fmt.Println("Restored")
	default:
		fmt.Println("Usage:")
		fmt.Println("  ./restic-backup-system -mode=backup backup_folder=path_to_folder")
		fmt.Println("  ./restic-backup-system -mode=restore restore_folder=path_to_folder")
		os.Exit(1)
	}
}
