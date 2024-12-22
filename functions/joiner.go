package functions

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	http "github.com/bogdanfinn/fhttp"
	"io"
	"log"
	"main/console"
	"net/url"
	"strings"
)

type PropResponse struct {
	Guild   string `json:"guild_id"`
	Channel struct {
		ID   string `json:"id"`
		Type int    `json:"type"`
	}
}

func BuildClient(method, urlStr string, body io.Reader, token, cookie, properties *string) *http.Response {
	client := &http.Client{}
	
	proxyURL, err := url.Parse("http://user:password@proxy-address:proxy-port")
	if err != nil {
		log.Fatalf("failed to parse proxy URL: %v\n", err)
	}
	client.Transport = &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
	}

	req, err := http.NewRequest(method, urlStr, body)
	if err != nil {
		log.Fatalf("failed to create request: %v\n", err)
	}

	if token != nil {
		req.Header.Set("Authorization", *token)
	}
	if cookie != nil {
		req.Header.Set("Cookie", *cookie)
	}
	if properties != nil {
		req.Header.Set("X-Properties", *properties)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("failed to perform request: %v\n", err)
	}
	return resp
}

func JoinServer(token, invite, cookie, properties string) {
	data := strings.NewReader(`{}`)
	resp := BuildClient(http.MethodPost, fmt.Sprintf("https://canary.discord.com/api/v9/invites/%s", invite), data, &token, &cookie, &properties)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Fatalf("failed to close client: %v\n", err)
		}
	}(resp.Body)
	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	switch resp.StatusCode {
	case http.StatusOK:
		console.DisplayText("JOINED", console.Colors["green"], token[:20], invite)
	case http.StatusBadRequest:
		console.DisplayText("FLAG", console.Colors["yellow"], token[:20], invite)
	case http.StatusTooManyRequests:
		console.DisplayText("RATE LIMITED", console.Colors["magenta"], token[:20], invite)
	default:
		var prop BodyResponse
		err = json.Unmarshal(bodyText, &prop)
		if err != nil {
			log.Fatalf("failed to unmarshal: %v\n", err)
		}
		console.DisplayText("FATAL", console.Colors["red"], token[:20], prop.Message.(string))
	}
}

func GetProperties(invite string) string {
	resp := BuildClient(http.MethodGet, fmt.Sprintf("https://canary.discord.com/api/v9/invites/%s?inputValue=%s&with_counts=true&with_expiration=true", invite, invite), nil, nil, nil, nil)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Fatalf("failed to close client: %v\n", err)
		}
	}(resp.Body)
	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	var prop PropResponse
	err = json.Unmarshal(bodyText, &prop)
	if err != nil {
		log.Fatalf("failed to unmarshal: %v\n", err)
	}
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("properties status code error: %d\n", resp.StatusCode)
	}
	guildId := prop.Guild
	channelId := prop.Channel.ID
	chanType := prop.Channel.Type
	str := fmt.Sprintf(`{"location": "Join Guild","location_guild_id":"%s","location_channel_id":"%s","location_channel_type":%d}`, guildId, channelId, chanType)
	return base64.StdEncoding.EncodeToString([]byte(str))
}
