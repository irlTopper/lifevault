package modules

import (
	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/s3"
	"github.com/revel/revel"
)

type AmazonWebServices struct {
	S3 S3Buckets
}

type S3Buckets struct {
	Desk    *s3.Bucket
	DeskDev *s3.Bucket
}

var AWS AmazonWebServices

func InitAWS() {
	// Get the aws auth
	auth, err := aws.GetAuth(revel.Config.StringDefault("aws.key", ""), revel.Config.StringDefault("aws.secret", ""))
	if err != nil {
		revel.WARN.Println("[AWS]: Auth setup failed:", err)
	}

	// Connect to the tw-desk bucket
	client := s3.New(auth, aws.USEast)
	AWS.S3.Desk = client.Bucket("tw-desk")
	_, err = AWS.S3.Desk.GetBucketContents()
	if err != nil {
		revel.WARN.Println("[AWS]: Issue connecting to tw-desk bucket:", err)
	}

	// Connect to the tw-desk-dev bucket
	AWS.S3.DeskDev = client.Bucket("tw-desk-dev")
	_, err = AWS.S3.DeskDev.GetBucketContents()
	if err != nil {
		revel.WARN.Println("[AWS]: Issue connecting to tw-desk-dev bucket:", err)
	}
}
