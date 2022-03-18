package mailerlite_test

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/mailerlite/mailerlite-go"
	"github.com/stretchr/testify/assert"
)

const (
	testKey = "valid-api-key"
)

type RoundTripFunc func(req *http.Request) *http.Response

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

//NewTestClient returns *http.Client with Transport replaced to avoid making real calls
func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: fn,
	}
}

func TestNewClient(t *testing.T) {
	ml := mailerlite.NewClient(testKey)

	assert.Equal(t, ml.APIKey(), testKey)
	assert.Equal(t, ml.Client(), http.DefaultClient)

	client := new(http.Client)
	ml.SetHttpClient(client)
	assert.Equal(t, client, ml.Client())

}

func TestCanMakeMockApiCall(t *testing.T) {
	client := mailerlite.NewClient(testKey)

	testClient := NewTestClient(func(req *http.Request) *http.Response {
		// Test request parameters
		assert.Equal(t, req.URL.String(), "https://connect.mailerlite.com/api/subscribers")
		return &http.Response{
			StatusCode: http.StatusAccepted,
			Body:       ioutil.NopCloser(bytes.NewBufferString(`OK`)),
		}
	})

	ctx := context.TODO()

	client.SetHttpClient(testClient)

	listOptions := &mailerlite.ListSubscriberOptions{}

	_, res, err := client.Subscriber.List(ctx, listOptions)
	if err != nil {
		return
	}

	assert.Equal(t, res.StatusCode, http.StatusAccepted)

}

func TestWillHandleError(t *testing.T) {
	client := mailerlite.NewClient(testKey)

	testClient := NewTestClient(func(req *http.Request) *http.Response {
		// return nil to force error from mock server
		return nil
	})

	ctx := context.TODO()

	client.SetHttpClient(testClient)

	listOptions := &mailerlite.ListSubscriberOptions{}

	_, _, err := client.Subscriber.List(ctx, listOptions)
	if err != nil {
		return
	}

	assert.Error(t, err)
}

func TestWillHandleAPIError(t *testing.T) {
	client := mailerlite.NewClient(testKey)

	testClient := NewTestClient(func(req *http.Request) *http.Response {
		return &http.Response{
			StatusCode: http.StatusBadRequest,
			Request:    req,
			Body: ioutil.NopCloser(strings.NewReader(`{"message":"The given data was invalid.",
			"errors": [{"filter": "The filter must be an array."}]}`)),
		}
	})

	ctx := context.TODO()

	client.SetHttpClient(testClient)

	listOptions := &mailerlite.ListSubscriberOptions{}

	_, _, err := client.Subscriber.List(ctx, listOptions)

	assert.Error(t, err)
	assert.IsType(t, err, &mailerlite.ErrorResponse{})
	assert.Equal(t, err.Error(), "GET https://connect.mailerlite.com/api/subscribers: 400 The given data was invalid.")
}

func TestWillHandleAPIFilters(t *testing.T) {
	client := mailerlite.NewClient(testKey)

	testClient := NewTestClient(func(req *http.Request) *http.Response {
		return &http.Response{
			StatusCode: http.StatusAccepted,
			Request:    req,
			Body: ioutil.NopCloser(strings.NewReader(`{
				"data": [
				  {
					"id": "123456789",
					"email": "client@example.com",
					"status": "active"
				  }
				],
				"links": {
				  "first": "https://connect.mailerlite.com/api/subscribers?page=1",
				  "last": "https://connect.mailerlite.com/api/subscribers?page=2",
				  "prev": null,
				  "next": "https://connect.mailerlite.com/api/subscribers?page=2"
				},
				"meta": {
				  "current_page": 1,
				  "from": 1,
				  "last_page": 2,
				  "links": [
					{
					  "url": null,
					  "label": "&laquo; Previous",
					  "active": false
					},
					{
					  "url": "https://connect.mailerlite.com/api/subscribers?page=1",
					  "label": "1",
					  "active": true
					},
					{
					  "url": "https://connect.mailerlite.com/api/subscribers?page=2",
					  "label": "2",
					  "active": false
					},
					{
					  "url": "https://connect.mailerlite.com/api/subscribers?page=2",
					  "label": "Next &raquo;",
					  "active": false
					}
				  ],
				  "path": "https://connect.mailerlite.com/api/subscribers",
				  "per_page": "1",
				  "to": 1,
				  "total": 2
				}
			}`)),
		}
	})

	ctx := context.TODO()

	client.SetHttpClient(testClient)

	listOptions := &mailerlite.ListSubscriberOptions{
		Filter: &mailerlite.Filter{Name: "status", Value: "active"},
	}

	subscribers, _, _ := client.Subscriber.List(ctx, listOptions)

	assert.Equal(t, len(subscribers.Data), 1)
	assert.Equal(t, subscribers.Data[0].Status, "active")
}

func TestWillHandleAPIAuthError(t *testing.T) {

	client := mailerlite.NewClient(testKey)

	testClient := NewTestClient(func(req *http.Request) *http.Response {
		return &http.Response{
			StatusCode: http.StatusUnauthorized,
			Request:    req,
			Body:       ioutil.NopCloser(strings.NewReader(`{"message": "Unauthenticated."}`)),
		}
	})

	ctx := context.TODO()

	client.SetHttpClient(testClient)

	client.SetAPIKey("invalid-api-key")

	listOptions := &mailerlite.ListSubscriberOptions{}

	_, _, err := client.Subscriber.List(ctx, listOptions)

	assert.Error(t, err)
	assert.IsType(t, err, &mailerlite.AuthError{})
	assert.Equal(t, err.Error(), "GET https://connect.mailerlite.com/api/subscribers: 401 Unauthenticated.")
}
