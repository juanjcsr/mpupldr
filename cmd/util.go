package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func checkFile(filename string) (os.FileInfo, error) {
	// b, err := ioutil.ReadFile(filename)
	s, err := os.Stat(filename)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func getNewFileName(fp, ext string) string {
	b := filepath.Base(fp)
	e := filepath.Ext(fp)
	fname := strings.TrimSpace(b[0 : len(b)-len(e)])
	return fname + "." + ext
}

func getS3Session(s3t *S3MapboxTokens) (*session.Session, error) {
	s, err := session.NewSession(&aws.Config{
		Region:      aws.String("us-east-1"),
		Credentials: credentials.NewStaticCredentials(s3t.AccessKeyID, s3t.SecretAccessKey, s3t.SessionToken),
	})
	if err != nil {
		return nil, fmt.Errorf("Could not reach AWS S3 service, try later: %s", err)
	}
	return s, nil
}

func execTippeCanoe(file, filename string, args []string) error {
	cmd := exec.Command("tippecanoe", "-o", file, filename)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("could not call tippecanoe: %s", err)
	}
	return nil
}

func checkMapboxUpload(id, user string) (bool, error) {
	mURL := "https://api.mapbox.com/uploads/v1/%s/%s?access_token=%s"

	url := fmt.Sprintf(mURL, user, id, mapboxToken)
	mUR := new(MapboxUploadResult)
	r, err := http.Get(url)
	if err != nil {
		return false, fmt.Errorf("could not make request: %v", err)
	}
	defer r.Body.Close()

	if err = json.NewDecoder(r.Body).Decode(mUR); err != nil {
		return false, fmt.Errorf("could not decode response: %v", err)
	}
	return mUR.Complete, nil
}

func saveToMapbox(filename, bucket, key, username, tilename string) (*MapboxUploadResult, error) {
	mURL := "https://api.mapbox.com/uploads/v1/%s?access_token=%s"

	url := fmt.Sprintf(mURL, username, mapboxToken)
	s3URL := "http://%s.s3.amazonaws.com/%s"
	tileset := "%s.%s"

	mUp := MapboxUploadData{
		URL:     fmt.Sprintf(s3URL, bucket, key),
		Tileset: fmt.Sprintf(tileset, username, tilename),
	}

	b, err := json.Marshal(mUp)
	if err != nil {
		return nil, fmt.Errorf("could not create upload json: %v", err)
	}
	fmt.Printf("/nUploading to mapbox: %s /n", url)
	r, err := http.Post(url, "application/json", bytes.NewBuffer(b))
	bd := r.Body
	defer bd.Close()

	mUR := new(MapboxUploadResult)
	if err = json.NewDecoder(bd).Decode(mUR); err != nil {
		return nil, fmt.Errorf("could not decode data: %v", err)
	}
	fmt.Printf("%+v /n", mUR)
	fmt.Println(mUR)
	return mUR, nil
}

func uploadToS3(filename string, s *session.Session, bucket string, key string) error {

	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	fileinfo, _ := file.Stat()
	size := fileinfo.Size()
	buff := make([]byte, size)
	file.Read(buff)
	fmt.Println("uploading to S3...")
	_, err = s3.New(s).PutObject(&s3.PutObjectInput{
		Bucket:        aws.String(bucket),
		Key:           aws.String(key),
		Body:          bytes.NewReader(buff),
		ContentLength: aws.Int64(size),
	})
	if err != nil {
		return fmt.Errorf("could not upload file: %v", err)
	}
	return nil
}

func getS3Tokens(at, user string) *S3MapboxTokens {
	s3URL := "https://api.mapbox.com/uploads/v1/%s/credentials?access_token=%s"
	url := fmt.Sprintf(s3URL, user, at)
	// fmt.Printf("uploading file to mapbox to: %s", url)
	r, err := http.Post(url, "", nil)
	if err != nil {
		log.Fatalf("could not get S3 credentials: %s", err)
	}
	b := r.Body
	defer b.Close()
	mt := new(S3MapboxTokens)
	if err := json.NewDecoder(b).Decode(mt); err != nil {
		log.Fatalf("could not read s3 tokens: %s", err)
	}
	fmt.Printf("%+v", mt)
	return mt
}

type S3MapboxTokens struct {
	AccessKeyID     string `json:"accessKeyId"`
	Bucket          string `json:"bucket"`
	Key             string `json:"key"`
	SecretAccessKey string `json:"secretAccessKey"`
	SessionToken    string `json:"sessionToken"`
	URL             string `json:"url"`
}

type MapboxUploadData struct {
	URL     string `json:"url"`
	Tileset string `json:"tileset"`
}

type MapboxUploadResult struct {
	Complete bool   `json:"complete"`
	Tileset  string `json:"tileset"`
	Error    string `json:"error"`
	ID       string `json:"id"`
	Name     string `json:"name"`
	Modified string `json:"modified"`
	Created  string `json:"created"`
	Owner    string `json:"owner"`
	Progress int    `json:"progress"`
}

func (mu *MapboxUploadResult) String() string {
	r, _ := json.Marshal(mu)
	return string(r)
}
