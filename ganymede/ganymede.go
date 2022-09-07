package ganymede

import (
	"bytes"
	"encoding/json"
	"fmt"
	"ganymede-migrate/ceres"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type CreateVod struct {
	ChannelID        string `json:"channel_id"`
	ID               string `json:"id"`
	ExtID            string `json:"ext_id"`
	Title            string `json:"title"`
	Platform         string `json:"platform"`
	Type             string `json:"type"`
	Duration         int64  `json:"duration"`
	Views            int64  `json:"views"`
	Resolution       string `json:"resolution"`
	ThumbnailPath    string `json:"thumbnail_path"`
	WebThumbnailPath string `json:"web_thumbnail_path"`
	VideoPath        string `json:"video_path"`
	ChatPath         string `json:"chat_path"`
	ChatVideoPath    string `json:"chat_video_path"`
	InfoPath         string `json:"info_path"`
	StreamedAt       string `json:"streamed_at"`
}

type Channel struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	ImagePath   string `json:"image_path"`
	UpdatedAt   string `json:"updated_at"`
	CreatedAt   string `json:"created_at"`
	Edges       Edges  `json:"edges"`
}

type Edges struct {
}

type Service struct {
	Host              string
	AccessTokenCookie *http.Cookie
}

func NewService() *Service {
	gUsername := os.Getenv("GANYMEDE_USERNAME")
	gPassword := os.Getenv("GANYMEDE_PASSWORD")
	gHost := os.Getenv("GANYMEDE_HOST")
	if gUsername == "" || gPassword == "" || gHost == "" {
		panic("GANYMEDE_USERNAME, GANYMEDE_PASSWORD, and GANYMEDE_HOST must be set")
	}

	client := &http.Client{}

	jsonBody := []byte(`{"username": "` + gUsername + `", "password": "` + gPassword + `"}`)
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/auth/login", gHost), bytes.NewBuffer(jsonBody))
	if err != nil {
		log.Panicf("failed to create request: %v", err)
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Panicf("failed to authenticate: %v", err)
	}

	defer resp.Body.Close()

	// check status code
	if resp.StatusCode != 200 {
		log.Panicf("failed to authenticate: %v", resp.Status)
	}

	// get cookies
	cookies := resp.Cookies()

	for _, cookie := range cookies {
		if cookie.Name == "access-token" {
			return &Service{Host: gHost, AccessTokenCookie: cookie}
		}
	}

	return nil
}

func (s *Service) GetChannel(cName string) (Channel, error) {

	// get channel
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/channel/name/%s", s.Host, cName), nil)
	if err != nil {
		return Channel{}, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.AddCookie(s.AccessTokenCookie)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return Channel{}, fmt.Errorf("failed to get channel: %s with error: %v", cName, err)
	}

	defer resp.Body.Close()

	// check status code
	if resp.StatusCode != 200 {
		return Channel{}, fmt.Errorf("failed to get channel: %s with error: %v", cName, resp.Status)
	}

	// get channel
	var channel Channel
	err = json.NewDecoder(resp.Body).Decode(&channel)
	if err != nil {
		log.Panicf("failed to decode channel: %v", err)
	}

	return channel, nil

}

func (s *Service) CreateVod(vod ceres.VOD, vID string, channel Channel) error {
	var thumbnailPath string
	if vod.ThumbnailPath != "" {
		thumbnailPath = fmt.Sprintf("/vods/%s/%s_%s/%s-thumbnail.jpg", channel.Name, vod.ID, vID, vod.ID)
	} else {
		thumbnailPath = ""
	}
	var webThumbnailPath string
	if vod.ThumbnailPath != "" {
		webThumbnailPath = fmt.Sprintf("/vods/%s/%s_%s/%s-web_thumbnail.jpg", channel.Name, vod.ID, vID, vod.ID)
	} else {
		webThumbnailPath = ""
	}
	var videoPath string
	if vod.VideoPath != "" {
		videoPath = fmt.Sprintf("/vods/%s/%s_%s/%s-video.mp4", channel.Name, vod.ID, vID, vod.ID)
	} else {
		videoPath = ""
	}
	var chatPath string
	if vod.ChatPath != "" {
		chatPath = fmt.Sprintf("/vods/%s/%s_%s/%s-chat.json", channel.Name, vod.ID, vID, vod.ID)
	} else {
		chatPath = ""
	}
	var chatVideoPath string
	if vod.ChatVideoPath != "" {
		chatVideoPath = fmt.Sprintf("/vods/%s/%s_%s/%s-chat.mp4", channel.Name, vod.ID, vID, vod.ID)
	} else {
		chatVideoPath = ""
	}
	var infoPath string
	if vod.VODInfoPath != "" {
		infoPath = fmt.Sprintf("/vods/%s/%s_%s/%s-info.json", channel.Name, vod.ID, vID, vod.ID)
	} else {
		infoPath = ""
	}
	var views int64
	if vod.ViewCount != 0 {
		views = vod.ViewCount
	} else {
		views = 1
	}

	// create vod
	ganymedeVodRequest := CreateVod{
		ChannelID:        channel.ID,
		ID:               vID,
		ExtID:            vod.ID,
		Title:            vod.Title,
		Platform:         "twitch",
		Type:             vod.BroadcastType,
		Duration:         vod.Duration,
		Views:            views,
		Resolution:       vod.Resolution,
		ThumbnailPath:    thumbnailPath,
		WebThumbnailPath: webThumbnailPath,
		VideoPath:        videoPath,
		ChatPath:         chatPath,
		ChatVideoPath:    chatVideoPath,
		InfoPath:         infoPath,
		StreamedAt:       vod.CreatedAt,
	}

	jsonBody, err := json.Marshal(ganymedeVodRequest)
	if err != nil {
		return fmt.Errorf("failed to marshal ganymede vod request: %v", err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/vod", s.Host), bytes.NewBuffer(jsonBody))

	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Add("Content-Type", "application/json")

	req.AddCookie(s.AccessTokenCookie)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to create vod: %v", err)
	}

	defer resp.Body.Close()

	// check status code
	if resp.StatusCode == 409 {
		return fmt.Errorf("vod already exists")
	} else if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)

		return fmt.Errorf("failed to create vod with error %v: %s", resp.Status, body)
	}

	return nil
}

func (s *Service) RenameVodFiles(vod ceres.VOD, vID string, channel Channel) error {
	// Create new vod directory
	// https://github.com/Zibbp/ganymede-migrate/issues/1
	newVodDir := fmt.Sprintf("/vods/%s/%s_%s", channel.Name, vod.ID, vID)
	err := os.MkdirAll(newVodDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create new vod directory: %v", err)
	}

	var newthumbnailPath string
	if vod.ThumbnailPath != "" {
		newthumbnailPath = fmt.Sprintf("/vods/%s/%s_%s/%s-thumbnail.jpg", channel.Name, vod.ID, vID, vod.ID)
		oldThumbnailPath := fmt.Sprintf("/vods/%s", vod.ThumbnailPath)
		err := os.Rename(oldThumbnailPath, newthumbnailPath)
		if err != nil {
			log.Printf("failed to rename thumbnail in vod: %s with error: %v", vod.ID, err)
		}
	} else {
		newthumbnailPath = ""
	}
	var newwebThumbnailPath string
	if vod.ThumbnailPath != "" {
		newwebThumbnailPath = fmt.Sprintf("/vods/%s/%s_%s/%s-web_thumbnail.jpg", channel.Name, vod.ID, vID, vod.ID)
		oldWebThumbnailPath := fmt.Sprintf("/vods/%s", vod.WebThumbnailPath)
		err := os.Rename(oldWebThumbnailPath, newwebThumbnailPath)
		if err != nil {
			log.Printf("failed to rename web thumbnail in vod: %s with error: %v", vod.ID, err)
		}
	} else {
		newwebThumbnailPath = ""
	}
	var newvideoPath string
	if vod.VideoPath != "" {
		newvideoPath = fmt.Sprintf("/vods/%s/%s_%s/%s-video.mp4", channel.Name, vod.ID, vID, vod.ID)
		oldVideoPath := fmt.Sprintf("/vods/%s", vod.VideoPath)
		err := os.Rename(oldVideoPath, newvideoPath)
		if err != nil {
			log.Printf("failed to rename video in vod: %s with error: %v", vod.ID, err)
		}
	} else {
		newvideoPath = ""
	}
	var newchatPath string
	if vod.ChatPath != "" {
		newchatPath = fmt.Sprintf("/vods/%s/%s_%s/%s-chat.json", channel.Name, vod.ID, vID, vod.ID)
		oldChatPath := fmt.Sprintf("/vods/%s", vod.ChatPath)
		err := os.Rename(oldChatPath, newchatPath)
		if err != nil {
			log.Printf("failed to rename chat in vod: %s with error: %v", vod.ID, err)
		}
	} else {
		newchatPath = ""
	}
	var newchatVideoPath string
	if vod.ChatVideoPath != "" {
		newchatVideoPath = fmt.Sprintf("/vods/%s/%s_%s/%s-chat.mp4", channel.Name, vod.ID, vID, vod.ID)
		oldChatVideoPath := fmt.Sprintf("/vods/%s", vod.ChatVideoPath)
		err := os.Rename(oldChatVideoPath, newchatVideoPath)
		if err != nil {
			log.Printf("failed to rename chat video in vod: %s with error: %v", vod.ID, err)
		}
	} else {
		newchatVideoPath = ""
	}
	var newinfoPath string
	if vod.VODInfoPath != "" {
		newinfoPath = fmt.Sprintf("/vods/%s/%s_%s/%s-info.json", channel.Name, vod.ID, vID, vod.ID)
		oldInfoPath := fmt.Sprintf("/vods/%s", vod.VODInfoPath)
		err := os.Rename(oldInfoPath, newinfoPath)
		if err != nil {
			log.Printf("failed to rename info in vod: %s with error: %v", vod.ID, err)
		}
	} else {
		newinfoPath = ""
	}
	return nil
}

func (s *Service) RemoveOldFolders(vod ceres.VOD, vID string, channel Channel) error {
	err := os.RemoveAll(fmt.Sprintf("/vods/%s/%s", channel.Name, vod.ID))
	if err != nil {
		log.Printf("failed to remove vod folder: %v", err)
	}
	return nil
}
