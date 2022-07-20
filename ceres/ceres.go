package ceres

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type Auth struct {
	User        User   `json:"user"`
	AccessToken string `json:"accessToken"`
}

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Roles    string `json:"roles"`
}

type Vods []VOD

type VOD struct {
	ID               string  `json:"id"`
	Title            string  `json:"title"`
	BroadcastType    string  `json:"broadcastType"`
	Duration         int64   `json:"duration"`
	ViewCount        int64   `json:"viewCount"`
	Resolution       string  `json:"resolution"`
	Downloading      bool    `json:"downloading"`
	ThumbnailPath    string  `json:"thumbnailPath"`
	WebThumbnailPath string  `json:"webThumbnailPath"`
	VideoPath        string  `json:"videoPath"`
	ChatPath         string  `json:"chatPath"`
	ChatVideoPath    string  `json:"chatVideoPath"`
	VODInfoPath      string  `json:"vodInfoPath"`
	CreatedAt        string  `json:"createdAt"`
	UpdatedAt        string  `json:"updatedAt"`
	Channel          Channel `json:"channel"`
}

type Channel struct {
	ID          string `json:"id"`
	Login       string `json:"login"`
	DisplayName string `json:"displayName"`
}

type Service struct {
	Host        string
	AccessToken string
}

func NewService() *Service {
	cUsername := os.Getenv("CERES_USERNAME")
	cPassword := os.Getenv("CERES_PASSWORD")
	cHost := os.Getenv("CERES_HOST")
	if cUsername == "" || cPassword == "" || cHost == "" {
		panic("CERES_USERNAME, CERES_PASSWORD, and CERES_HOST must be set")
	}

	client := &http.Client{}

	// set body
	data := url.Values{}
	data.Set("username", cUsername)
	data.Set("password", cPassword)

	encodedData := data.Encode()

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/v1/auth/login", cHost), strings.NewReader(encodedData))
	if err != nil {
		panic(err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		log.Panicf("failed to authenticate: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Panicf("failed to read response body: %v", err)

	}

	var auth Auth
	err = json.Unmarshal(body, &auth)
	if err != nil {
		log.Panicf("failed to unmarshal response: %v", err)
	}

	return &Service{Host: cHost, AccessToken: auth.AccessToken}
}

func (s *Service) GetAllVods() (Vods, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/v1/vods/all", s.Host), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", s.AccessToken))

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var vods Vods
	err = json.Unmarshal(body, &vods)
	if err != nil {
		return nil, err
	}

	return vods, nil
}
