package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const (
	maxUploads         = 16              // Maximum number of concurrent uploads
	retryMaxAttempts   = 4               // Maximum number of retry attempts
	retrySleepDuration = 1 * time.Second // Duration to wait between retries
)

func cloudflareR2Client() *s3.Client {
	r2Resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: fmt.Sprintf("https://%s.r2.cloudflarestorage.com", ACCOUNT_ID),
		}, nil
	})

	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithEndpointResolverWithOptions(r2Resolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(ACCESS_KEY_ID, ACCESS_KEY_SECRET, "")),
	)
	if err != nil {
		log.Fatal(err)
	}

	return s3.NewFromConfig(cfg)
}

func main() {
	var exitCode atomic.Int32

	err := os.Chdir(*targetDirectory)
	if err != nil {
		log.Fatalf("Could not chdir to %v: %v", *targetDirectory, err)
	}

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

				file, err := os.Open(path)
				if err != nil {
					log.Println("Error:", err)
					exitCode.Add(1)
					return
				}
				defer file.Close()

				// Upload file to S3 with retry
				err = retryUpload(client, path, file)
				if err != nil {
					log.Println("Upload error:", err)
					exitCode.Add(1)
				}

				// Calculate progress
				progress <- 1
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

	// Quoting os.Exit's documentation:
	// "For portability, the status code should be in the range [0, 125]."
	// let's cap the exit code to 125 for that
	exitWith := exitCode.Load()
	if exitWith > 125 {
		exitWith = 125
	}

	os.Exit(int(exitWith))
}

// retryUpload retries the upload operation with a maximum number of attempts
func retryUpload(client *s3.Client, path string, file *os.File) error {
	attempt := 1
	for {
		_, err := client.PutObject(context.Background(), &s3.PutObjectInput{
			Bucket:             aws.String(*bucketName),
			Key:                aws.String(path),
			Body:               file,
			ContentType:        aws.String(*contentType),
			ContentEncoding:    aws.String(*contentEncoding),
			ContentDisposition: aws.String("inline"),
		})
		if err == nil {
			return nil
		}

		if attempt >= retryMaxAttempts {
			return fmt.Errorf("maximum retry attempts exceeded: %w", err)
		}

		log.Printf("Upload error (attempt %d): %v. Retrying...", attempt, err)

		attempt++
		time.Sleep(retrySleepDuration)
	}
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
