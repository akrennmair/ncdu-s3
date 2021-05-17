package main

import (
	"encoding/json"
	"flag"
	"io"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/fraugster/cli"
)

type record struct {
	children     dir
	Name         string `json:"name"`
	ASize        int64  `json:"asize,omitempty"`
	DSize        int64  `json:"dsize,omitempty"`
	Inode        int64  `json:"ino,omitempty"`
	Device       int    `json:"dev"`
	ModifiedTime int64  `json:"mtime,omitempty"`
}

type dir map[string]*record

func main() {
	ctx := cli.Context()

	flag.Parse()

	if len(flag.Args()) < 1 {
		log.Fatalf("missing S3 URL")
	}

	s3url, err := url.Parse(flag.Arg(0))
	if err != nil {
		log.Fatalf("Parsing S3 URL %s failed: %v", flag.Arg(0), err)
	}
	if s3url.Scheme != "s3" {
		log.Fatalf("Expected s3 URL, got %s URL instead", s3url.Scheme)
	}

	var out io.Writer = os.Stdout

	if len(flag.Args()) == 2 {
		f, err := os.OpenFile(flag.Arg(1), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			log.Fatalf("Couldn't open %s for writing: %v", flag.Arg(1), err)
		}
		defer f.Close()
		out = f
	}

	sesh, err := session.NewSession()
	if err != nil {
		log.Fatalf("Failed to create AWS session: %v", err)
	}

	s3Client := s3.New(sesh)

	bucket := s3url.Host
	path := s3url.Path[1:]

	var prefix *string
	if path != "" {
		prefix = &path
	}

	allFiles := make(dir)
	inode := int64(9001)

	if err := s3Client.ListObjectsPagesWithContext(ctx, &s3.ListObjectsInput{
		Bucket: aws.String(bucket),
		Prefix: prefix,
	}, func(output *s3.ListObjectsOutput, lastPage bool) bool {
		for _, obj := range output.Contents {
			key := *obj.Key
			if path != "" {
				key = key[len(path):]
			}
			if len(key) == 0 {
				continue
			}

			if key[0] == '/' {
				key = key[1:]
			}

			parts := strings.Split(key, "/")

			curDir := allFiles
			for idx, part := range parts {
				if part == "" {
					break
				}
				rec, ok := curDir[part]
				if !ok {
					rec = &record{
						children: make(dir),
						Name:     part,
					}
					curDir[part] = rec
				}
				if idx == len(parts)-1 {
					rec.Device = 0
					rec.Inode = inode
					inode++
					rec.ASize = *obj.Size
					rec.DSize = *obj.Size
					rec.ModifiedTime = obj.LastModified.Unix()
				} else {
					curDir = rec.children
				}
			}
		}
		return !lastPage
	}); err != nil {
		log.Fatalf("Listing files in %s failed: %v", s3url.String(), err)
	}

	allData := []interface{}{
		1,
		1,
		map[string]interface{}{
			"progname":  "ncdu-s3",
			"progver":   "0.0.0",
			"timestamp": time.Now().Unix(),
		},
	}

	fileList := []interface{}{
		record{
			Name: s3url.String(),
		},
	}

	fileList = append(fileList, listFiles(allFiles)...)

	allData = append(allData, fileList)

	rawData, err := json.MarshalIndent(allData, "", "  ")
	if err != nil {
		log.Fatalf("Couldn't generate JSON: %v", err)
	}

	n, err := out.Write(rawData)
	if err != nil {
		log.Fatalf("Couldn't write generated JSON to output: %v", err)
	}
	if n < len(rawData) {
		log.Fatalf("Short write when writing generated JSON to output: %v", err)
	}
}

func listFiles(d dir) (result []interface{}) {
	for _, rec := range d {
		if rec.Name == "" {
			continue
		}
		if len(rec.children) > 0 {
			result = append(result, append([]interface{}{rec}, listFiles(rec.children)...))
		} else {
			result = append(result, rec)
		}
	}
	return result
}
