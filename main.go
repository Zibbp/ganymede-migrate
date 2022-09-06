package main

import (
	"encoding/json"
	"ganymede-migrate/ceres"
	"ganymede-migrate/ganymede"
	"github.com/google/uuid"
	"io"
	"io/ioutil"
	"log"
	"os"
)

func main() {

	f, err := os.OpenFile("/data/log.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, os.ModePerm)
	if err != nil {
		return
	}
	defer func() {
		f.Close()
	}()

	//Just combine it, OS Stdout stands for standard output stream
	multiWriter := io.MultiWriter(os.Stdout, f)
	log.SetOutput(multiWriter)

	ceresService := ceres.NewService()
	log.Println("Logged into Ceres.")

	ganymedeService := ganymede.NewService()
	if ganymedeService == nil {
		log.Fatal("Failed to login to Ganymede.")
	}
	log.Println("Logged into Ganymede.")

	vods, err := ceresService.GetAllVods()
	if err != nil {
		log.Println("GetAllVods failed: ", err)
		return
	}

	var failedVods []ceres.VOD

	shouldRename := os.Getenv("SHOULD_RENAME")
	shouldDelete := os.Getenv("SHOULD_DELETE")

	for _, vod := range vods {
		if shouldDelete != "true" {
			// Get channel from Ganymede
			channel, err := ganymedeService.GetChannel(vod.Channel.Login)
			if err != nil {
				log.Printf("Skipping VOD: %s because of: GetChannel failed: %v", vod.ID, err)
				failedVods = append(failedVods, vod)
				continue
			}
			// Generate UUID for VOD creation
			vID, err := uuid.NewUUID()
			if err != nil {
				log.Panicf("Failed to generate UUID: %v", err)
			}
			// Create VOD in Ganymede
			err = ganymedeService.CreateVod(vod, vID.String(), channel)
			if err != nil {
				if err.Error() == "vod already exists" {
					continue
				} else {
					log.Printf("Skipping VOD: %s because of: CreateVod failed: %v", vod.ID, err)
					failedVods = append(failedVods, vod)
					continue
				}
			}
		}
		
		// Rename VOD files for Ganymede
		// Only run if ENV SHOULD_RENAME is set to true and ENV SHOULD_DELETE is not set
		if shouldRename == "true" && shouldDelete != "true" {
			err = ganymedeService.RenameVodFiles(vod, vID.String(), channel)
			if err != nil {
				log.Printf("Skipping VOD: %s because of: RenameVodFiles failed: %v", vod.ID, err)
				failedVods = append(failedVods, vod)
				continue
			}
		}
		// Remove old VOD folders from ceres
		// Only run if ENV SHOULD_DELETE is set to true
		if shouldDelete == "true" {
			err = ganymedeService.RemoveOldFolders(vod, vID.String(), channel)
			if err != nil {
				log.Printf("Skipping VOD: %s because of: RemoveOldFolders failed: %v", vod.ID, err)
				failedVods = append(failedVods, vod)
				continue
			}
		}
	}

	// Write failed VODs
	if len(failedVods) > 0 {
		file, err := json.Marshal(failedVods)
		if err != nil {
			log.Printf("Failed to marshal failed VODs: %v", err)
		}
		err = ioutil.WriteFile("/data/failed_vods.json", file, 0644)
	}

}
