package fcm

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"time"
)

const (
	// FCM server address
	serverURL = "https://fcm.googleapis.com/fcm/send"

	// Network timeout for connecting to the server. Not setting it may create a large
	// pool of waiting connectins in case of network problems.
	connectionTimeout = 5 * time.Second

	PriorityHigh   = "high"
	PriorityNormal = "normal"
)

// Error messages
const (
	ErrorMissingRegistration       = "MissingRegistration"
	ErrorInvalidRegistration       = "InvalidRegistration"
	ErrorNotRegistered             = "NotRegistered"
	ErrorInvalidPackageName        = "InvalidPackageName"
	ErrorMismatchSenderId          = "MismatchSenderId"
	ErrorMessageTooBig             = "MessageTooBig"
	ErrorInvalidDataKey            = "InvalidDataKey"
	ErrorInvalidTtl                = "InvalidTtl"
	ErrorUnavailable               = "Unavailable"
	ErrorInternalServerError       = "InternalServerError"
	ErrorDeviceMessageRateExceeded = "DeviceMessageRateExceeded"
	ErrorTopicsMessageRateExceeded = "TopicsMessageRateExceeded"
)

// HttpMessage is an FCM HTTP request message
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

// HttpResponse is an FCM response message
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

// NewClient returns an FCM client. The client is expected to be
// long-lived. It maintains an internal pool of HTTP connections.
// Multiple sumultaneous Send requests can be issued on the same client.
func NewClient(apikey string) *Client {
	return &Client{
		apiKey: "key=" + apikey,
		connection: &http.Transport{
			Dial: (&net.Dialer{
				Timeout: connectionTimeout,
			}).Dial,
			TLSHandshakeTimeout: connectionTimeout,
		},
	}
}

// SendHttp is a blocking call to send an HTTP message to FCM server.
// Multiple Send requests can be issued simultaneously on the same
// Client.
func (c *Client) SendHttp(msg *HttpMessage) (*HttpResponse, error) {

	// Encode message to JSON
	var rw bytes.Buffer
	encoder := json.NewEncoder(&rw)
	err := encoder.Encode(msg)
	if err != nil {
		return nil, err
	}

	// Format request
	req, err := http.NewRequest(http.MethodPost, serverURL, &rw)
	if err != nil {
		return nil, err
	}
	req.Header.Add(http.CanonicalHeaderKey("Content-Type"), "application/json")
	req.Header.Add(http.CanonicalHeaderKey("Authorization"), c.apiKey)

	//debug, err := httputil.DumpRequest(req, true)
	//log.Printf("request: '%s'", string(debug))

	// Call the server, issue HTTP POST, wait for response
	httpResp, err := c.connection.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	// debug, err := httputil.DumpResponse(httpResp, true)
	// log.Printf("response: '%s'", string(debug))

	// Read response completely and close the body to make
	// the underlying connection reusable.
	body, err := ioutil.ReadAll(httpResp.Body)
	httpResp.Body.Close()
	if err != nil {
		return nil, err
	}

	if httpResp.StatusCode != http.StatusOK {
		// Assuming non-JSON response
		return nil, errors.New(httpResp.Status + ": " + string(body))
	}

	// Decode JSON response
	var response HttpResponse
	err = json.Unmarshal(body, &response)

	// Get value of retry-after if present
	if err == nil {
		c.retryAfter = httpResp.Header.Get(http.CanonicalHeaderKey("Retry-After"))
	}

	return &response, err
}

// GetRetryAfter returns the number fo seconds to wait before retrying Send in case the previous
// Send has failed.
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

// PostHttp is a non-blocking version of Send. Not implemented yet.
func (c *Client) PostHttp(msg *HttpMessage) (<-chan *HttpResponse, error) {
	return nil, errors.New("Not implmented")
}
