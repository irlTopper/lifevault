package aws

import (
	"net/http"
	"net/url"
	"time"

	"github.com/AdRoll/goamz/aws"
	"github.com/AdRoll/goamz/s3"
	"github.com/irlTopper/lifevault/app/modules/logger/app"
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
	if !revel.DevMode || revel.Config.BoolDefault("aws.enabled", false) {
		// Get the aws auth
		auth, err := aws.GetAuth(revel.Config.StringDefault("aws.key", ""), revel.Config.StringDefault("aws.secret", ""), "", time.Now().AddDate(1, 0, 0))
		if err != nil {
			revel.WARN.Println("[AWS]: Auth setup failed:", err)
		}

		// Connect to the tw-desk bucket
		client := s3.New(auth, aws.USEast)
		AWS.S3.Desk = client.Bucket("tw-desk")
		_, err = AWS.S3.Desk.Get("testconnectionworks")
		if err != nil && err.Error() != "The specified key does not exist." {
			logger.Log.Panicln("[AWS]: Issue connecting to tw-desk bucket:", err)
		}

		// Connect to the tw-desk-dev bucket
		AWS.S3.DeskDev = client.Bucket("tw-desk-dev")
		_, err = AWS.S3.DeskDev.Get("testconnectionworks")
		if err != nil && err.Error() != "The specified key does not exist." {
			logger.Log.Panicln("[AWS]: Issue connecting to tw-desk-dev bucket:", err)
		}

		logger.Log.Println("[AWS]: Connected to AWS")
	}
}

func (s *S3Buckets) GetFileURL(s3Path string, fileName string, download bool) (downloadURL string) {
	vals := url.Values{}

	if download {
		vals["response-content-disposition"] = []string{"attachment; filename=\"" + fileName + "\""}
	}

	if revel.DevMode {
		downloadURL = s.DeskDev.SignedURLWithArgs(s3Path, time.Now().UTC().Add(time.Hour*1), vals, http.Header{})
	} else {
		downloadURL = s.Desk.SignedURLWithArgs(s3Path, time.Now().UTC().Add(time.Hour*1), vals, http.Header{})
	}
	return
}
