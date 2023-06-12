package main

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var targetDirectory = flag.String("directory", ".", "A directory to upload")

const (
	maxUploads = 4 // Maximum number of concurrent uploads
)

func haveToGetEnvironmentVariable(name string) string {
	val, ok := os.LookupEnv(name)
	if !ok {
		log.Panicf("Missing required environment variable %s\n", name)
	}
	if val == "" {
		log.Panicf("Required environment variable %s is empty\n", name)
	}
	return val
}

func cloudflareR2Client() *s3.Client {
	var accountId = haveToGetEnvironmentVariable("R2_ACCOUNT_ID")
	var accessKeyId = haveToGetEnvironmentVariable("R2_ACCESS_KEY_ID")
	var accessKeySecret = haveToGetEnvironmentVariable("R2_ACCESS_KEY_SECRET")

	r2Resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...any) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountId),
		}, nil
	})

	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithEndpointResolverWithOptions(r2Resolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyId, accessKeySecret, "")),
	)
	if err != nil {
		log.Fatal(err)
	}

	return s3.NewFromConfig(cfg)
}

func main() {
	var exitCode atomic.Int32

	flag.Parse()
	err := os.Chdir(*targetDirectory)
	if err != nil {
		log.Fatalf("Could not chdir to %v: %v", *targetDirectory, err)
	}

	var bucketName = haveToGetEnvironmentVariable("R2_BUCKET_NAME")

	client := cloudflareR2Client()

	directoryPath := "."

	// Get the total number of files in the directory
	totalFiles := countFiles(directoryPath)

	// Create a channel to track progress
	progress := make(chan int, totalFiles)

	// Display progress as a percentage
	go displayProgress(progress, totalFiles)

	// Create a wait group to ensure all uploads are completed
	var wg sync.WaitGroup

	// Create a semaphore to control concurrent uploads
	semaphore := make(chan struct{}, maxUploads)

	// Upload files
	err = filepath.WalkDir(directoryPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			log.Println("Error:", err)
			exitCode.Add(1)
			return nil
		}

		if !d.IsDir() {
			wg.Add(1)
			semaphore <- struct{}{} // Acquire semaphore

			go func() {
				defer func() {
					<-semaphore // Release semaphore
					wg.Done()
				}()

				md5Sum, err := calculateMD5(path)
				if err != nil {
					log.Printf("Couldn't calculate an md5 checksum for file %s: %v\n", path, err)
					exitCode.Add(1)
					return
				}

				file, err := os.Open(path)
				if err != nil {
					log.Println("Error:", err)
					exitCode.Add(1)
					return
				}
				defer file.Close()

				// Calculate progress
				progress <- 1

				// Upload file to S3
				_, err = client.PutObject(context.Background(), &s3.PutObjectInput{
					Bucket:             aws.String(bucketName),
					Key:                aws.String(path),
					Body:               file,
					ContentType:        aws.String("application/json"),
					ContentEncoding:    aws.String("gzip"),
					ContentDisposition: aws.String("inline"),
					ContentMD5:         aws.String(md5Sum),
				})
				if err != nil {
					log.Println("R2 PutObject error:", err)
					exitCode.Add(1)
					return
				}
			}()
		}

		return nil
	})

	if err != nil {
		exitCode.Add(1)
		log.Println("Error:", err)
	}

	// Wait for all uploads to complete
	wg.Wait()

	close(progress)

	os.Exit(int(exitCode.Load()))
}

func calculateMD5(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("opening a file for md5 calculation failed: %w", err)
	}

	defer file.Close()

	hash := md5.New()
	_, err = io.Copy(hash, file)

	if err != nil {
		return "", fmt.Errorf("reading file for md5 calculation failed: %w", err)
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// countFiles counts the total number of files in the given directory (recursively)
func countFiles(directoryPath string) int {
	count := 0

	err := filepath.WalkDir(directoryPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			log.Println("Error:", err)
			return nil
		}

		if !d.IsDir() {
			count++
		}

		return nil
	})

	if err != nil {
		log.Println("Error:", err)
	}

	return count
}

// displayProgress displays the progress as a percentage
func displayProgress(progress <-chan int, totalFiles int) {
	uploadedFiles := 0

	for p := range progress {
		uploadedFiles += p
		percentage := (float64(uploadedFiles) / float64(totalFiles)) * 100
		fmt.Printf("Upload progress: %.2f%%\n", percentage)
	}
}
