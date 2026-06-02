package agent

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

type GeoInfo struct {
	IP      string
	Country string
}

func DetectGeo() GeoInfo {
	client := &http.Client{Timeout: 4 * time.Second}
	for _, url := range []string{
		"https://ipapi.co/json/",
		"https://ipwho.is/",
	} {
		info, ok := detectGeoFrom(client, url)
		if ok {
			return info
		}
	}
	return GeoInfo{Country: "unknown"}
}

func detectGeoFrom(client *http.Client, url string) (GeoInfo, bool) {
	resp, err := client.Get(url)
	if err != nil {
		return GeoInfo{}, false
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return GeoInfo{}, false
	}
	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return GeoInfo{}, false
	}
	ip := firstString(body, "ip")
	country := firstString(body, "country_name", "country")
	if strings.EqualFold(country, "unknown") {
		country = ""
	}
	if ip == "" && country == "" {
		return GeoInfo{}, false
	}
	if country == "" {
		country = "unknown"
	}
	return GeoInfo{IP: ip, Country: country}, true
}

func firstString(body map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := body[key].(string); ok && value != "" {
			return value
		}
	}
	return ""
}
