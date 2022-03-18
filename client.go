package mailerlite

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"

	"github.com/google/go-querystring/query"
)

const APIBase string = "https://connect.mailerlite.com/api"

// Client - base api client
type Client struct {
	apiBase string
	apiKey  string
	client  *http.Client

	common service // Reuse a single struct.

	// Services
	Subscriber *SubscriberService
}

type service struct {
	client *Client
}

// Response is a MailerLite API response. This wraps the standard http.Response
type Response struct {
	*http.Response
}

// ErrorResponse is a MailerLite API error response. This wraps the standard http.Response
type ErrorResponse struct {
	Response *http.Response // HTTP response that caused this error
	Message  string         `json:"message"` // error message
	Errors   interface{}    `json:"errors"`
}

func (r *ErrorResponse) Error() string {
	return fmt.Sprintf("%v %v: %d %v",
		r.Response.Request.Method, r.Response.Request.URL,
		r.Response.StatusCode, r.Message)
}

// AuthError occurs when using HTTP Authentication fails
type AuthError ErrorResponse

func (r *AuthError) Error() string { return (*ErrorResponse)(r).Error() }

// NewClient - creates a new client instance.
func NewClient(apiKey string) *Client {
	client := &Client{
		apiBase: APIBase,
		apiKey:  apiKey,
		client:  http.DefaultClient,
	}

	client.common.client = client
	client.Subscriber = (*SubscriberService)(&client.common)

	return client
}

// APIKey - Get api key after it has been created
func (c *Client) APIKey() string {
	return c.apiKey
}

// Client - Get the current client
func (c *Client) Client() *http.Client {
	return c.client
}

// SetHttpClient - Set the client if you want more control over the client implementation
func (c *Client) SetHttpClient(client *http.Client) {
	c.client = client
}

// SetAPIKey - Set the client api key
func (c *Client) SetAPIKey(apikey string) {
	c.apiKey = apikey
}

func (c *Client) newRequest(method, path string, body interface{}) (*http.Request, error) {
	reqURL := fmt.Sprintf("%s%s", c.apiBase, path)
	reqBodyBytes := new(bytes.Buffer)

	if method == http.MethodPost ||
		method == http.MethodPut ||
		method == http.MethodDelete {
		err := json.NewEncoder(reqBodyBytes).Encode(body)
		if err != nil {
			return nil, err
		}
	} else if method == http.MethodGet {
		reqURL, _ = addOptions(reqURL, body)
	}

	req, err := http.NewRequest(method, reqURL, reqBodyBytes)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Mailerlite-Client-Golang-v1")

	return req, nil
}

func (c *Client) do(ctx context.Context, req *http.Request, v interface{}) (*Response, error) {
	req = req.WithContext(ctx)
	resp, err := c.client.Do(req)
	if err != nil {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		return nil, err
	}

	response := newResponse(resp)

	err = checkResponse(resp)
	if err != nil {
		defer resp.Body.Close()
		return response, err
	}

	if v != nil {
		err = json.NewDecoder(resp.Body).Decode(v)
		if err != nil {
			return nil, err
		}
	}

	return response, err
}

// newResponse creates a new Response for the provided http.Response.
// r must not be nil.
func newResponse(r *http.Response) *Response {
	response := &Response{Response: r}
	return response
}

func checkResponse(r *http.Response) error {
	if r.StatusCode == http.StatusAccepted {
		return nil
	}

	if c := r.StatusCode; c >= 200 && c <= 299 {
		return nil
	}

	errorResponse := &ErrorResponse{Response: r}
	data, err := ioutil.ReadAll(r.Body)

	if err == nil && len(data) > 0 {
		err := json.Unmarshal(data, errorResponse)
		if err != nil {
			errorResponse.Message = string(data)
		}
	}

	switch {
	case r.StatusCode == http.StatusUnauthorized:
		return (*AuthError)(errorResponse)
	default:
		return errorResponse
	}
}

func addOptions(s string, opt interface{}) (string, error) {
	v := reflect.ValueOf(opt)

	if v.Kind() == reflect.Ptr && v.IsNil() {
		return s, nil
	}

	origURL, err := url.Parse(s)
	if err != nil {
		return s, err
	}

	origValues := origURL.Query()

	newValues, err := query.Values(opt)
	if err != nil {
		return s, err
	}

	filterKey := ""
	filterValue := ""

	for k, v := range newValues {
		if k == "Filter" {
			continue
		}
		if k == "Filter[Name]" {
			filterKey = fmt.Sprintf("filter[%s]", v[0])
			continue
		}
		if k == "Filter[Value]" {
			filterValue = v[0]
			continue
		}
		origValues[k] = v
	}

	if filterKey != "" {
		origValues.Add(filterKey, filterValue)
	}

	origURL.RawQuery = origValues.Encode()

	return origURL.String(), nil
}

// Bool is a helper routine that allocates a new bool value
// to store v and returns a pointer to it.
func Bool(v bool) *bool { return &v }

// Int is a helper routine that allocates a new int value
// to store v and returns a pointer to it.
func Int(v int) *int { return &v }

// Int64 is a helper routine that allocates a new int64 value
// to store v and returns a pointer to it.
func Int64(v int64) *int64 { return &v }

// String is a helper routine that allocates a new string value
// to store v and returns a pointer to it.
func String(v string) *string { return &v }
