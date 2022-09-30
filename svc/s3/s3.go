package s3

import (
	"context"
	"io"
	"path"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type Instance interface {
	UploadFile(ctx context.Context, opts *s3.PutObjectInput) error
	DownloadFile(ctx context.Context, output io.Writer, opts *s3.GetObjectInput) error
	DeleteFile(ctx context.Context, opts *s3.DeleteObjectInput) error
	ListBuckets(ctx context.Context) (*s3.ListBucketsOutput, error)
	CopyFile(ctx context.Context, opts *s3.CopyObjectInput) error
	SetACL(ctx context.Context, opts *s3.PutObjectAclInput) error
	ComposeKey(s ...string) string
}

type s3Inst struct {
	session *session.Session
	s3      *s3.S3
	ns      string
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
		session: s,
		s3:      s3.New(s),
		ns:      o.Namespace,
	}, nil
}

func (a *s3Inst) ListBuckets(ctx context.Context) (*s3.ListBucketsOutput, error) {
	return a.s3.ListBucketsWithContext(ctx, &s3.ListBucketsInput{})
}

func (a *s3Inst) UploadFile(ctx context.Context, opts *s3.PutObjectInput) error {
	_, err := a.s3.PutObjectWithContext(ctx, opts)

	return err
}

func (a *s3Inst) DeleteFile(ctx context.Context, opts *s3.DeleteObjectInput) error {
	_, err := a.s3.DeleteObjectWithContext(ctx, opts)

	return err
}

func (a *s3Inst) DownloadFile(ctx context.Context, output io.Writer, opts *s3.GetObjectInput) error {
	resp, err := a.s3.GetObject(opts)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	_, err = io.Copy(output, resp.Body)

	return err
}

func (a *s3Inst) SetACL(ctx context.Context, opts *s3.PutObjectAclInput) error {
	_, err := a.s3.PutObjectAclWithContext(ctx, opts)

	return err
}

func (a *s3Inst) CopyFile(ctx context.Context, opts *s3.CopyObjectInput) error {
	_, err := a.s3.CopyObject(opts)

	return err
}

func (a *s3Inst) ComposeKey(s ...string) string {
	return path.Join(a.ns, path.Join(s...))
}
