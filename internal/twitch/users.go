package twitch

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"time"

	"github.com/nfnt/resize"
	log "github.com/sirupsen/logrus"
)

const usersURI = "/users"

var _ UsersI = (*usersClient)(nil)

var (
	ErrTooManyLoginNames = errors.New("too many login sets. Cannot be more than 100")
)

type usersResponse struct {
	Data []*User `json:"data"`
}

// User describes a Twitch user
type User struct {
	BroadcasterType string    `json:"broadcaster_type"`
	CreatedAt       time.Time `json:"created_at"`
	Description     string    `json:"description"`
	DisplayName     string    `json:"display_name"`
	ID              string    `json:"id"`
	Login           string    `json:"login"`
	OfflineImageURL string    `json:"offline_image_url"`
	ProfileImageURL string    `json:"profile_image_url"`
	Type            string    `json:"type"`
	ViewCount       int       `json:"view_count"`
}

type UsersI interface {
	// Get returns information about one or more specified Twitch users.
	// Users are identified by optional user IDs and/or login name.
	// If neither a user ID nor a login name is specified, the user is looked up by Bearer token.
	// https://dev.twitch.tv/docs/api/reference#get-users
	Get(login ...string) ([]*User, error)

	// ProfileImageBytes load the given profile from URL
	// Reduce its size and send it as bytes
	ProfileImageBytes(user *User) ([]byte, error)

	// Me returns current connected user
	Me() *User
}

type usersClient struct {
	c *Client
}

func (u *usersClient) Get(login ...string) ([]*User, error) {
	if len(login) > 100 {
		return nil, ErrTooManyLoginNames
	}

	ul := apiURL + usersURI
	req, err := http.NewRequest(http.MethodGet, ul, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to make request: %w", err)
	}

	// Specify wanted users
	q := req.URL.Query()
	for _, u := range login {
		q.Add("login", u)
	}
	req.URL.RawQuery = q.Encode()

	resp, err := u.c.httpClient.Do(req)
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

	data := new(usersResponse)
	if err := json.Unmarshal(body, data); err != nil {
		return nil, fmt.Errorf("unable to read response body: %w", err)
	}

	return data.Data, nil
}

func (u *usersClient) ProfileImageBytes(user *User) ([]byte, error) {
	// check if exist in cache
	if data, found := u.retrieveImageFromCache(user.Login); found {
		log.Debugf("image [%s] found in cache", user.Login)
		return data, nil
	}

	// download it
	resp, err := u.c.httpClient.Get(user.ProfileImageURL)
	if err != nil {
		return nil, fmt.Errorf("unable to read profile image URL: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non 200 response code")
	}

	// duplicate response body
	var contentTypeBuffer bytes.Buffer
	tee := io.TeeReader(resp.Body, &contentTypeBuffer)
	defer resp.Body.Close() // we are done with body

	// read it, it contains the image
	imageBytes, _ := io.ReadAll(tee)
	var imageBuffer = bytes.NewBuffer(imageBytes)

	// detect image type
	// https://stackoverflow.com/a/38175140
	buff := make([]byte, 512)
	if _, err := contentTypeBuffer.Read(buff); err != nil {
		return nil, fmt.Errorf("unable determine image type: %w", err)
	}
	imageContentType := http.DetectContentType(buff)

	// load image
	var img image.Image
	switch imageContentType {
	case "image/png":
		img, err = png.Decode(imageBuffer)
	case "image/jpeg":
		img, err = jpeg.Decode(imageBuffer)
	default:
		return nil, fmt.Errorf("unexpected image content-type: %s", imageContentType)
	}

	if err != nil {
		return nil, fmt.Errorf("unable to decode image: %w", err)
	}

	// resize it
	m := resize.Resize(32, 32, img, resize.NearestNeighbor)

	imageBuffer.Reset()
	switch imageContentType {
	case "image/png":
		err = png.Encode(imageBuffer, m)
	case "image/jpeg":
		err = jpeg.Encode(imageBuffer, m, nil)
	}

	if err != nil {
		return nil, fmt.Errorf("unable to encode image: %w", err)
	}

	// store image in cache
	data := imageBuffer.Bytes()
	log.Debugf("storing image [%s] in cache", user.Login)
	if err := u.storeImageInCache(data, user.Login); err != nil {
		log.Errorf("unable to store image [%s] in cache: %v", user.Login, err)
	}

	return data, nil
}

// storeImageInCache write given bytes to file
func (u *usersClient) storeImageInCache(image []byte, name string) error {
	return u.c.cache.Write(name, image)
}

// retrieveImageFromCache return given file if exist
func (u *usersClient) retrieveImageFromCache(name string) ([]byte, bool) {
	if !u.c.cache.Has(name) {
		return nil, false
	}

	data, err := u.c.cache.Read(name)
	if err != nil {
		log.Warningf("error while loading file: %v", err)
		return nil, false
	}

	return data, true
}

func (u *usersClient) Me() *User {
	return u.c.me
}
