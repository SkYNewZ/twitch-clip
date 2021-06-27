package twitch

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	streamsURI         = "/streams"
	followedStreamsURI = "/streams/followed"
)

var _ StreamsI = (*streamsClient)(nil)

var (
	ErrTooManyUserLoginNames = errors.New("too many user login sets. Cannot be more than 100")
)

type streamsResponse struct {
	Data       []*Stream `json:"data"`
	Pagination struct {
		Cursor string `json:"cursor"`
	} `json:"pagination"`
}

// Stream describes a Twitch stream
type Stream struct {
	GameID       string    `json:"game_id,omitempty"`
	GameName     string    `json:"game_name,omitempty"`
	ID           string    `json:"id,omitempty"`
	IsMature     bool      `json:"is_mature,omitempty"`
	Language     string    `json:"language,omitempty"`
	StartedAt    time.Time `json:"started_at,omitempty"`
	TagIds       []string  `json:"tag_ids,omitempty"`
	ThumbnailURL string    `json:"thumbnail_url,omitempty"`
	Title        string    `json:"title,omitempty"`
	Type         string    `json:"type,omitempty"`
	UserID       string    `json:"user_id,omitempty"`
	UserLogin    string    `json:"user_login,omitempty"`
	UserName     string    `json:"user_name,omitempty"`
	ViewerCount  int       `json:"viewer_count,omitempty"`
}

type StreamsI interface {
	// GetStream returns information about active streams.
	// Streams are returned sorted by number of current viewers, in descending order.
	// If any, returns streams broadcast by one or more specified user login names. You can specify up to 100 names.
	// https://dev.twitch.tv/docs/api/reference#get-streams
	GetStream(userLogin ...string) ([]*Stream, error)

	// GetFollowed returns information about active streams belonging to channels that the authenticated user follows.
	// Streams are returned sorted by number of current viewers, in descending order.
	// Across multiple pages of results, there may be duplicate or missing streams, as viewers join and leave streams.
	// https://dev.twitch.tv/docs/api/reference#get-followed-streams
	GetFollowed() ([]*Stream, error)
}

type streamsClient struct {
	c *Client
}

func (s *streamsClient) GetStream(userLogin ...string) ([]*Stream, error) {
	if len(userLogin) > 100 {
		return nil, ErrTooManyUserLoginNames
	}

	u := apiURL + streamsURI
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to make requestw: %w", err)
	}

	// Specify wanted users
	q := req.URL.Query()
	for _, u := range userLogin {
		q.Add("user_login", u)
	}
	req.URL.RawQuery = q.Encode()

	resp, err := s.c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("twitch error: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		// try to read Twitch error
		e := new(Error)
		if err := json.Unmarshal(body, e); err == nil {
			return nil, e
		}

		return nil, fmt.Errorf("twitch error: %s", body)
	}

	data := new(streamsResponse)
	if err := json.Unmarshal(body, data); err != nil {
		return nil, fmt.Errorf("unable to read response body: %v", err)
	}

	return data.Data, nil
}

func (s *streamsClient) GetFollowed() ([]*Stream, error) {
	u := apiURL + followedStreamsURI
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to make requestw: %w", err)
	}

	// Specify wanted users
	q := req.URL.Query()
	q.Set("user_id", s.c.Users.Me().ID)
	req.URL.RawQuery = q.Encode()

	resp, err := s.c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("twitch error: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		// try to read Twitch error
		e := new(Error)
		if err := json.Unmarshal(body, e); err == nil {
			return nil, e
		}

		return nil, fmt.Errorf("twitch error: %s", body)
	}

	data := new(streamsResponse)
	if err := json.Unmarshal(body, data); err != nil {
		return nil, fmt.Errorf("unable to read response body: %v", err)
	}

	return data.Data, nil
}
