package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

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
	flag.Parse()
	err := os.Chdir(*targetDirectory)
	if err != nil {
		log.Fatalf("Could not chdir to %v: %v", *targetDirectory, err)
	}

	var bucketName = "top-of-github"

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
					return
				}
				defer file.Close()

				// Calculate progress
				progress <- 1

				// Upload file to S3
				_, err = client.PutObject(context.Background(), &s3.PutObjectInput{
					Bucket: aws.String(bucketName),
					Key:    aws.String(path),
					Body:   file,
				})
				if err != nil {
					log.Println("Error:", err)
					return
				}
			}()
		}

		return nil
	})

	if err != nil {
		log.Println("Error:", err)
	}

	// Wait for all uploads to complete
	wg.Wait()

	close(progress)

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
