package s3

import (
	"context"
	"io"
	"path"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type Instance interface {
	UploadFile(ctx context.Context, opts *s3manager.UploadInput) error
	DownloadFile(ctx context.Context, output io.WriterAt, opts *s3.GetObjectInput) error
	ListBuckets(ctx context.Context) (*s3.ListBucketsOutput, error)
	ComposeKey(s ...string) string
}

type s3Inst struct {
	session    *session.Session
	downloader *s3manager.Downloader
	uploader   *s3manager.Uploader
	s3         *s3.S3
	ns         string
}

func New(ctx context.Context, o Options) (Instance, error) {
	s, err := session.NewSession(&aws.Config{
		Credentials:      credentials.NewStaticCredentials(o.AccessToken, o.SecretKey, ""),
		Region:           aws.String(o.Region),
		S3ForcePathStyle: aws.Bool(true),
		Endpoint:         aws.String(o.Endpoint),
	})
	if err != nil {
		return nil, err
	}

	return &s3Inst{
		session:    s,
		downloader: s3manager.NewDownloader(s),
		uploader:   s3manager.NewUploader(s),
		s3:         s3.New(s),
		ns:         o.Namespace,
	}, nil
}

func (a *s3Inst) ListBuckets(ctx context.Context) (*s3.ListBucketsOutput, error) {
	return a.s3.ListBucketsWithContext(ctx, &s3.ListBucketsInput{})
}

func (a *s3Inst) UploadFile(ctx context.Context, opts *s3manager.UploadInput) error {
	_, err := a.uploader.UploadWithContext(ctx, opts)

	return err
}

func (a *s3Inst) DownloadFile(ctx context.Context, output io.WriterAt, opts *s3.GetObjectInput) error {
	_, err := a.downloader.DownloadWithContext(ctx, output, opts)

	return err
}

func (a *s3Inst) ComposeKey(s ...string) string {
	return path.Join(a.ns, path.Join(s...))
}
