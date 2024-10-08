package wmn

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	DefaultURL       = "https://raw.githubusercontent.com/WebBreacher/WhatsMyName/refs/heads/main/wmn-data.json"
	DefaultQueryTime = 45 * time.Second
)

type WmnData struct {
	Sites []Site `json:"sites"`
}

type Site struct {
	Name     string   `json:"name"`
	URICheck string   `json:"uri_check"`
	ECode    int      `json:"e_code"`
	EString  string   `json:"e_string"`
	MString  string   `json:"m_string"`
	MCode    int      `json:"m_code"`
	Known    []string `json:"known"`
	Cat      string   `json:"cat"`
}

func Sites() ([]Site, error) {
	return SitesFromURL(DefaultURL)
}

func SitesFromURL(url string) ([]Site, error) {

	response, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("unable to get Sites from URL '%s': %v", url, err)
	}
	defer response.Body.Close()

	byteValue, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read body from URL '%s': %v", url, err)
	}

	var w WmnData
	err = json.Unmarshal(byteValue, &w)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal byteData: %v", err)
	}

	if len(w.Sites) == 0 {
		return nil, fmt.Errorf("No Sites found")
	}

	return w.Sites, nil
}

func (s *Site) UrlForUser(user string) string {
	return strings.TrimSpace(strings.Replace(s.URICheck, "{account}", user, 1))
}

func (s *Site) CheckUser(ctx context.Context, user string) (bool, error) {
	ctx2, cancel := context.WithTimeout(ctx, DefaultQueryTime)
	defer cancel()
	return s.checkUser(ctx2, user)
}

func (s *Site) CheckUserWithDuration(ctx context.Context, user string, timeout time.Duration) (bool, error) {
	ctx2, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return s.checkUser(ctx2, user)
}

func (s *Site) checkUser(ctx context.Context, user string) (bool, error) {
	// Ignore invalid certificates
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.UrlForUser(user), nil)
	if err != nil {
		return false, fmt.Errorf("unable to create request: %v", err)
	}

	response, err := httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("unable to perform request: %v", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return false, fmt.Errorf("error reading response body: %v", err)
	}

	return response.StatusCode == s.ECode && strings.Contains(string(body), s.EString), nil
}
