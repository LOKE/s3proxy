package main

import (
	"flag"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func main() {
	addr := flag.String("address", ":7000", "the address to bind to")
	region := flag.String("region", "", "aws region")
	bucket := flag.String("bucket", "", "aws bucket")

	flag.Parse()
	sess := session.Must(session.NewSession(&aws.Config{Region: region}))
	s3Client := s3.New(sess, aws.NewConfig().WithRegion(*region))

	http.ListenAndServe(*addr, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" && r.Method != "HEAD" {
			http.Error(w, "Resource only supports GET and HEAD", http.StatusMethodNotAllowed)
			return
		}

		key := r.URL.Path
		if strings.HasSuffix(key, "/") {
			key += "index.html"
		}
		result, err := s3Client.GetObject(&s3.GetObjectInput{
			Bucket: bucket,
			Key:    &key,
		})
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchKey:
				http.NotFound(w, r)
				return
			}
			log.Printf("unknown aws error %v", err)
			http.Error(w, "unknown aws error", http.StatusInternalServerError)

		}
		defer result.Body.Close()

		if result.ContentType != nil {
			w.Header().Set("Content-Type", *result.ContentType)
		}

		_, err = io.Copy(w, result.Body)
	}))
}
