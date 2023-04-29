package main

import (
	"embed"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	bolt "go.etcd.io/bbolt"
)

var (
	port   int
	dbPath string
	db     *bolt.DB
)

func init() {
	flag.IntVar(&port, "port", 9092, "HTTP server port number")
	flag.StringVar(&dbPath, "db", "my.db", "BoltDB file path")
}

func main() {
	flag.Parse()

	// Open Bolt database
	var err error
	o := bolt.DefaultOptions
	o.Timeout = 5 * time.Second

	db, err = bolt.Open(dbPath, 0600, o)
	if err != nil {
		log.Fatal(fmt.Errorf("fail to open bolt db: %w", err))
	}
	defer db.Close()

	// Initialize router and routes
	mux := http.NewServeMux()
	mux.HandleFunc("/buckets/", listBucketKeys)

	// Start server
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	log.Printf("Server started on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

type bucketDetail struct {
	Key      string
	Value    string
	IsNested bool
}

func listTopBuckets() ([]bucketDetail, error) {
	var bucketDetails []bucketDetail
	err := db.View(func(tx *bolt.Tx) error {
		return tx.ForEach(func(name []byte, _ *bolt.Bucket) error {
			bucketDetails = append(bucketDetails, bucketDetail{
				Key:      toString(name),
				Value:    "",
				IsNested: true,
			})
			return nil
		})
	})
	return bucketDetails, err
}

func listNestedBuckets(bucketNames []string) ([]bucketDetail, error) {
	var bucketDetails []bucketDetail
	err := db.View(func(tx *bolt.Tx) error {
		var bkt = tx.Bucket([]byte(bucketNames[0]))
		for _, bucketName := range bucketNames[1:] {
			bkt = bkt.Bucket([]byte(bucketName))
			if bkt == nil {
				return fmt.Errorf("bucket '%s' not found", bucketName)
			}
		}

		err := bkt.ForEach(func(k, v []byte) error {
			// nested buckets
			isNested := false
			if v == nil {
				isNested = true
			}

			bucketDetails = append(bucketDetails, bucketDetail{
				Key:      toString(k),
				Value:    toString(v),
				IsNested: isNested,
			})
			return nil
		})

		return err
	})
	return bucketDetails, err
}

//go:embed templates/*
var templates embed.FS

// listBucketKeys handles the /buckets/... endpoint
func listBucketKeys(w http.ResponseWriter, r *http.Request) {
	// Extract bucket name from URL path
	path := strings.TrimPrefix(r.URL.Path, "/buckets/")
	bucketNames := strings.Split(path, "/")

	var bucketDetails []bucketDetail
	var err error

	// No bucket specified, list all root-level buckets
	if path == "" {
		bucketDetails, err = listTopBuckets()
	} else {
		bucketDetails, err = listNestedBuckets(bucketNames)
	}

	// Render HTML template
	type templateData struct {
		Error         error
		BucketPath    string
		BucketDetails []bucketDetail
	}
	data := templateData{
		Error:         err,
		BucketPath:    path,
		BucketDetails: bucketDetails,
	}

	tmpl, err := template.ParseFS(templates, "templates/buckets.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/html")
	err = tmpl.Execute(w, data)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// toString defines how to convert bucket key/value to string
func toString(buf []byte) string {
	if buf == nil {
		return ""
	}
	return string(buf)
}
