package fcm

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	// FCM server address
	server_url = "https://fcm.googleapis.com/fcm/send"

	// Network timeout for connecting to the server. Not setting it may create a large
	// pool of waiting connectins in acse of network problems.
	connection_timeout = 5 * time.Second
)

// Message is an FCM request message
type HttpMessage struct {
	To                    string        `json:"to,omitempty"`
	RegistrationIds       []string      `json:"registration_ids,omitempty"`
	Condition             string        `json:"condition,omitempty"`
	CollapseKey           string        `json:"collapse_key,omitempty"`
	Priority              string        `json:"priority,omitempty"`
	ContentAvailable      bool          `json:"content_available,omitempty"`
	TimeToLive            *uint         `json:"time_to_live,omitempty"`
	RestrictedPackageName string        `json:"restricted_package_name,omitempty"`
	DryRun                bool          `json:"dry_run,omitempty"`
	Data                  interface{}   `json:"data,omitempty"`
	Notification          *Notification `json:"notification,omitempty"`
}

// Response is an FCM response message
type HttpResponse struct {
	MulticastId  int      `json:"multicast_id"`
	Success      int      `json:"success"`
	Fail         int      `json:"failure"`
	CanonicalIds int      `json:"canonical_ids"`
	Results      []Result `json:"results,omitempty"`
}

type Result struct {
	MessageId      string `json:"message_id"`
	RegistrationId string `json:"registration_id"`
	Error          string `json:"error"`
}

// Notification notification message structure
type Notification struct {
	Title        string `json:"title,omitempty"`
	Body         string `json:"body,omitempty"`
	Sound        string `json:"sound,omitempty"`
	ClickAction  string `json:"click_action,omitempty"`
	BodyLocKey   string `json:"body_loc_key,omitempty"`
	BodyLocArgs  string `json:"body_loc_args,omitempty"`
	TitleLocKey  string `json:"title_loc_key,omitempty"`
	TitleLocArgs string `json:"title_loc_args,omitempty"`

	// Android only
	Icon  string `json:"icon,omitempty"`
	Tag   string `json:"tag,omitempty"`
	Color string `json:"color,omitempty"`

	// iOS only
	Badge string `json:"badge,omitempty"`
}

type Client struct {
	apiKey     string
	connection *http.Transport
	retryAfter string
}

func NewClient(apikey string) *Client {
	return &Client{
		apiKey: "key=" + apikey,
		connection: &http.Transport{
			Dial: (&net.Dialer{
				Timeout: connection_timeout,
			}).Dial,
			TLSHandshakeTimeout: connection_timeout,
		},
	}
}

// Send is a blocking call to send a message to FCM server.
// Multiple Send requests can be issued simultaneously on the same
// Client.
func (c *Client) Send(msg *HttpMessage) (*HttpResponse, error) {

	// Encode message to JSON
	var rw bytes.Buffer
	encoder := json.NewEncoder(&rw)
	err := encoder.Encode(msg)
	if err != nil {
		return nil, err
	}

	// Format request
	req, err := http.NewRequest(http.MethodPost, server_url, &rw)
	if err != nil {
		return nil, err
	}
	req.Header.Add(http.CanonicalHeaderKey("Content-Type"), "application/json")
	req.Header.Add(http.CanonicalHeaderKey("Authorization"), c.apiKey)

	// Call the server, issue HTTP POST, wait for response
	httpResp, err := c.connection.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	// Read response completely and close the body to make
	// the underlying connection reusable.
	data, err := ioutil.ReadAll(httpResp.Body)
	httpResp.Body.Close()
	if err != nil {
		return nil, err
	}

	// Decode JSON response
	var response HttpResponse
	err = json.Unmarshal(data, &response)

	// Get value of retry-after if present
	if err == nil {
		c.retryAfter = httpResp.Header.Get(http.CanonicalHeaderKey("Retry-After"))
	}

	return &response, err
}

func (c *Client) GetRetryAfter() uint {
	if c.retryAfter == "" {
		return 0
	}
	if ra, err := strconv.Atoi(c.retryAfter); err == nil {
		return uint(ra)
	}
	if ts, err := http.ParseTime(c.retryAfter); err == nil {
		sec := ts.Sub(time.Now()).Seconds()
		if sec < 0 {
			return 0
		}
		return uint(sec)
	}
	return 0
}

func (c *Client) Post(msg *HttpMessage) <-chan *HttpResponse {
	return nil
}
