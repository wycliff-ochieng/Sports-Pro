package filestore

import (
	"log"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type FileStore struct {
	Client *minio.Client
	Bucket string
}

func NewFileStore(endpoint, accessKey, secretKey, bucket string, useSSL bool) (*FileStore, error) {

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds: credentials.NewStaticV4(accessKey, secretKey, ""),
	})

	if err != nil {
		log.Printf("creating the minIO client error: %s", err)
	}
	return &FileStore{
		Client: minioClient,
		Bucket: bucket,
	}, err
}
