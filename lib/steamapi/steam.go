package steamapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type steamRespGetPublishedFileDetails struct {
	Response struct {
		PublishedFileDetails []struct {
			Title string `json:"title"`
		} `json:"publishedfiledetails"`
	} `json:"response"`
}

func FetchSteamTitle(id string) (string, error) {
	body := bytes.NewBufferString(
		fmt.Sprintf("itemcount=1&publishedfileids[0]=%s", id),
	)

	req, err := http.NewRequest("POST",
		"https://api.steampowered.com/ISteamRemoteStorage/GetPublishedFileDetails/v1/",
		body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var sr steamRespGetPublishedFileDetails
	if err := json.NewDecoder(resp.Body).Decode(&sr); err != nil {
		return "", err
	}

	if len(sr.Response.PublishedFileDetails) == 0 {
		return "", fmt.Errorf("not found")
	}

	return sr.Response.PublishedFileDetails[0].Title, nil
}
