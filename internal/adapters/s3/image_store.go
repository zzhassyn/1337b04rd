package s3

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type Config struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	Region          string
	PostsBucket     string
	CommentsBucket  string
	PublicEndpoint  string
}

type ImageStore struct {
	cfg    Config
	client *http.Client
	log    *slog.Logger
}

func New(cfg Config, log *slog.Logger) (*ImageStore, error) {
	s := &ImageStore{
		cfg:    cfg,
		client: &http.Client{Timeout: 30 * time.Second},
		log:    log,
	}

	if err := s.ensureBucket(context.Background(), cfg.PostsBucket); err != nil {
		return nil, fmt.Errorf("s3: ensure posts bucker: %w", err)
	}

	if err := s.ensureBucket(context.Background(), cfg.CommentsBucket); err != nil {
		return nil, fmt.Errorf("s3: ensure comments bucket: %w", err)
	}

	return s, nil
}

func (s *ImageStore) UploadPostImage(ctx context.Context, filename string, data io.Reader) (string, error) {
	return s.upload(ctx, s.cfg.PostsBucket, filename, data)
}

func (s *ImageStore) UploadCommentImage(ctx context.Context, filename string, data io.Reader) (string, error) {
	return s.upload(ctx, s.cfg.CommentsBucket, filename, data)
}

func (s *ImageStore) ensureBucket(ctx context.Context, bucket string) error {
	url := fmt.Sprintf("%s/%s", s.cfg.Endpoint, bucket)

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, http.NoBody)
	if err != nil {
		return err
	}

	s.signRequest(req, bucket, "", []byte{})

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusConflict {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ensureBucket %s: status %d: %s", bucket, resp.StatusCode, string(b))
	}

	s.log.Info("s3: bucket ready", "bucket", bucket)

	return nil
}

func (s *ImageStore) upload(ctx context.Context, bucket, filename string, data io.Reader) (string, error) {
	body, err := io.ReadAll(data)
	if err != nil {
		return "", nil
	}

	objectKey := filename
	url := fmt.Sprintf("%s/%s/%s", s.cfg.Endpoint, bucket, objectKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, strings.NewReader(string(body)))
	if err != nil {
		return "", fmt.Errorf("s3.upload new request: %w", err)
	}

	contentType := detectContentType(filename)
	req.Header.Set("Content-Type", contentType)
	req.ContentLength = int64(len(body))

	s.signRequest(req, bucket, objectKey, body)

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("s3.upload do: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusCreated {
		respbody, _ := io.ReadAll(resp.Body)

		return "", fmt.Errorf("s3.upload: status %d: %s", resp.StatusCode, string(respbody))
	}

	publicURL := fmt.Sprintf("%s/%s/%s", s.cfg.PublicEndpoint, bucket, objectKey)
	s.log.Info("s3: uploaded", "bucket", bucket, "key", objectKey, "url", publicURL)

	return publicURL, nil
}

func (s *ImageStore) signRequest(req *http.Request, bucket, key string, body []byte) {
	now := time.Now().UTC()
	dateShort := now.Format("20060102")
	dateLong := now.Format("20060102T150405Z")

	req.Header.Set("x-amz-date", dateLong)
	req.Header.Set("Host", req.URL.Host)

	bodyHash := hashSHA256(body)
	req.Header.Set("x-amz-content-sha256", bodyHash)

	canonicalHeaders := fmt.Sprintf("host:%s\nx-amz-content-sha256:%sx-amz-date:%s\n", req.URL.Host, bodyHash, dateLong)
	signedHeaders := "host;x-amz-content-sha256;x-amz-date"

	canonicalURI := req.URL.Path
	if canonicalURI == "" {
		canonicalURI = "/"
	}

	canonicalRequest := strings.Join([]string{
		req.Method, canonicalURI, "", canonicalHeaders, signedHeaders, bodyHash},
		"\n")

	credentialScope := fmt.Sprintf("%s/%s/s3/aws_request", dateShort, s.cfg.Region)
	stringToSign := strings.Join([]string{"AWS4-HMAC-SHA256", dateLong, credentialScope, hashSHA256([]byte(canonicalRequest))},
		"\n")

	signingKey := deriveSigningKey(s.cfg.SecretAccessKey, dateShort, s.cfg.Region, "s3")
	signature := hex.EncodeToString(hmacSHA256(signingKey, []byte(stringToSign)))

	authHeader := fmt.Sprintf(
		"AWS4-HMAC-SHA256 Credential=%s/%s,SignedHeaders=%s,Signature=%s",
		s.cfg.AccessKeyID, credentialScope, signedHeaders, signature,
	)
	req.Header.Set("Authorization", authHeader)
}

func hashSHA256(data []byte) string {
	h := sha256.Sum256(data)

	return hex.EncodeToString(h[:])
}

func hmacSHA256(key, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)

	return h.Sum(nil)
}

func deriveSigningKey(secret, date, region, service string) []byte {
	kDate := hmacSHA256([]byte("AWS4"+secret), []byte(date))
	kRegion := hmacSHA256(kDate, []byte(region))
	kService := hmacSHA256(kRegion, []byte(service))
	kSigning := hmacSHA256(kService, []byte("aws4_request"))

	return kSigning
}

func detectContentType(filename string) string {
	lower := strings.ToLower(filename)

	switch {
	case strings.HasSuffix(lower, ".jpg") || strings.HasSuffix(lower, ".jpeg"):
		return "image/jpeg"
	case strings.HasSuffix(lower, ".png"):
		return "image/png"
	case strings.HasSuffix(lower, ".gif"):
		return "image/gif"
	case strings.HasPrefix(lower, ".webp"):
		return "image/webp"
	default:
		return "application/octet-stream"
	}
}
