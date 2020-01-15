package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"os"

	"github.com/alien4cloud/alien4cloud-go-client/v2/alien4cloud"
)

// Command arguments
var url, user, password, csar, workspace string

func init() {
	// Initialize command arguments
	flag.StringVar(&url, "url", "http://localhost:8088", "Alien4Cloud URL")
	flag.StringVar(&user, "user", "admin", "User")
	flag.StringVar(&password, "password", "changeme", "Password")
	flag.StringVar(&csar, "csar", "", "Path to the CSAR to upload")
	flag.StringVar(&workspace, "workspace", "", "Upload CSAR into the given workspace (premium feature leave empty on OSS version)")
}

func main() {

	// Parsing command arguments
	flag.Parse()

	if csar == "" {
		log.Panic("Mandatory argument 'csar' missing")
	}

	f, err := os.Open(csar)
	if err != nil {
		log.Panicf("Failed to read CSAR file: %v", err)
	}

	client, err := alien4cloud.NewClient(url, user, password, "", true)
	if err != nil {
		log.Panic(err)
	}

	// Can use context for cancelation
	ctx := context.Background()

	err = client.Login(ctx)
	if err != nil {
		log.Panic(err)
	}

	csarDef, err := client.CatalogService().UploadCSAR(ctx, f, workspace)
	if err != nil {
		var pErr alien4cloud.ParsingErr
		if !errors.As(err, &pErr) {
			log.Panicf("failed to upload CSAR: %v", err)
		}
		if pErr.HasCriticalErrors() {
			log.Println("Upload failed!")
			log.Fatal(err)
		}
		log.Printf("Non-critical errors: %v", err)
	}

	log.Println()
	log.Println()
	log.Printf("CSAR uploaded!")
	log.Printf("CSAR definition:\n%#v", csarDef)
}
