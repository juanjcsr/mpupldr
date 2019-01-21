package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

var mapboxToken string
var mapboxUser string
var mapboxargs []string

func init() {
	uploadCmd.PersistentFlags().StringVarP(&mapboxToken, "mapbox-token", "m", "", "Mapbox secret access token")
	uploadCmd.MarkPersistentFlagRequired("mapbox-token")
	uploadCmd.PersistentFlags().StringVarP(&mapboxUser, "mapbox-user", "u", "", "Mapbox username")
	uploadCmd.MarkPersistentFlagRequired("mapbox-user")
	uploadCmd.Flags().StringArrayVarP(&mapboxargs, "tippecanoe-args", "p", []string{""}, "Tippecanoe args")
	RootCmd.AddCommand(uploadCmd)

	uploadDir.PersistentFlags().StringVarP(&mapboxToken, "mapbox-token", "m", "", "Mapbox secret access token")
	uploadDir.MarkPersistentFlagRequired("mapbox-token")
	uploadDir.PersistentFlags().StringVarP(&mapboxUser, "mapbox-user", "u", "", "Mapbox username")
	uploadDir.MarkPersistentFlagRequired("mapbox-user")
	uploadDir.Flags().StringArrayVarP(&mapboxargs, "tippecanoe-args", "p", []string{""}, "Tippecanoe args")
	RootCmd.AddCommand(uploadDir)
}

var uploadDir = &cobra.Command{
	Use:   "uploadDir [path/to/directory]",
	Short: "upload geojsons in directory to Mapbox",
	Long:  "transfors and uploads all geojsons files inside given directory to Mapbox",
	Run: func(cmd *cobra.Command, args []string) {
		UploadDir(args[0])
	},
	Args: cobra.ExactArgs(1),
}

var uploadCmd = &cobra.Command{
	Use:   "upload [geojson to upload] [name]",
	Short: "uploads geojson to mapbox",
	Long:  "stages the current geojson to Mapbox S3 and then to Mapbox",
	Run: func(cmd *cobra.Command, args []string) {
		_, err := Upload(args[0], args[1])
		if err != nil {
			log.Fatalf("could not process file: %v", err)
		}
	},
	Args: cobra.ExactArgs(2),
}

func UploadDir(dirname string) {
	err := os.Chdir(dirname)
	if err != nil {
		log.Fatalf("could not open directory: %v", err)
	}

	files, err := ioutil.ReadDir(".")
	if err != nil {
		log.Fatalf("could not read directory: %v", err)
	}

	for _, f := range files {
		filen := f.Name()
		t := time.Now().Format("2006-01-02")

		name := filen[0 : len(filen)-len(filepath.Ext(filen))]
		fullName := name + "-" + t

		Upload(filen, fullName)
	}
}

func Upload(filename, name string) (*MapboxUploadResult, error) {
	_, err := checkFile(filename)
	if err != nil {
		return nil, fmt.Errorf("could not find geojson file %s: %v", filename, err)
	}
	fn := getNewFileName(name, "mbtiles")

	err = execTippeCanoe(fn, filename, mapboxargs)
	if err != nil {
		return nil, fmt.Errorf("tippecanoe with errors: %v", err)
	}

	s3t := getS3Tokens(mapboxToken, mapboxUser)
	session, err := getS3Session(s3t)
	if err != nil {
		return nil, err
	}
	err = uploadToS3(fn, session, s3t.Bucket, s3t.Key)
	if err != nil {
		return nil, err
	}

	res, err := saveToMapbox(fn, s3t.Bucket, s3t.Key, mapboxUser, name)
	d, err := checkMapboxUpload(res.ID, mapboxUser)
	if err != nil {
		return nil, fmt.Errorf("could not check mapbox upload data %v", err)
	}
	for d == false {
		fmt.Printf(".")
		time.Sleep(time.Second * 4)
		d, _ = checkMapboxUpload(res.ID, mapboxUser)
	}
	fmt.Printf("COMPLETED:\n%s -- %s\n\n", filename, res.Tileset)
	return res, nil
}
