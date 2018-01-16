package cmd

import (
	"fmt"
	"log"
	"time"

	"github.com/spf13/cobra"
)

var mapboxToken string
var mapboxUser string

func init() {
	uploadCmd.PersistentFlags().StringVarP(&mapboxToken, "mapbox-token", "m", "", "Mapbox secret access token")
	uploadCmd.MarkPersistentFlagRequired("mapbox-token")
	uploadCmd.PersistentFlags().StringVarP(&mapboxUser, "mapbox-user", "u", "", "Mapbox username")
	uploadCmd.MarkPersistentFlagRequired("mapbox-user")
	RootCmd.AddCommand(uploadCmd)
}

var uploadCmd = &cobra.Command{
	Use:   "upload [geojson to upload] [name]",
	Short: "uploads geojson to mapbox",
	Long:  "stages the current geojson to Mapbox S3 and then to Mapbox",
	Run: func(cmd *cobra.Command, args []string) {
		Upload(args[0], args[1])
	},
	Args: cobra.ExactArgs(2),
}

func Upload(filename, name string) {
	_, err := checkFile(filename)
	if err != nil {
		log.Fatalf("could not find geojson file %s: %v", filename, err)
	}
	fn := getNewFileName(filename, "mbtiles")

	err = execTippeCanoe(fn, filename, []string{"-zg", "--drop-densest-as-needed"})
	if err != nil {
		log.Fatalln(err)
	}

	s3t := getS3Tokens(mapboxToken, mapboxUser)
	session, err := getS3Session(s3t)
	if err != nil {
		log.Fatalln(err)
	}
	err = uploadToS3(fn, session, s3t.Bucket, s3t.Key)
	if err != nil {
		log.Fatalln(err)
	}

	res, err := saveToMapbox(fn, s3t.Bucket, s3t.Key, mapboxUser, name)
	d, err := checkMapboxUpload(res.ID, mapboxUser)
	if err != nil {
		log.Fatalf("could not check mapbox upload data %v", err)
	}
	for d == false {
		fmt.Printf(".")
		time.Sleep(time.Second * 4)
		d, _ = checkMapboxUpload(res.ID, mapboxUser)
	}
	fmt.Printf("COMPLETED %+v", res)
}
