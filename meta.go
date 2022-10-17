package mailerlite

import (
	"net/url"
	"strconv"
)

type Meta struct {
	CurrentPage int         `json:"current_page"`
	From        int         `json:"from"`
	LastPage    int         `json:"last_page"`
	Links       []MetaLinks `json:"links"`
	Path        string      `json:"path"`
	PerPage     int         `json:"per_page"`
	To          int         `json:"to"`
	Total       int         `json:"total"`
}

// Links manages links that are returned along with a List
type Links struct {
	First string `json:"first"`
	Last  string `json:"last"`
	Prev  string `json:"prev"`
	Next  string `json:"next"`
}

type MetaLinks struct {
	URL    interface{} `json:"url"`
	Label  string      `json:"label"`
	Active bool        `json:"active"`
}

// NextPageToken is the page token to request the next page of the list
func (l *Links) NextPageToken() (string, error) {
	return l.nextPageToken()
}

// PrevPageToken is the page token to request the previous page of the list
func (l *Links) PrevPageToken() (string, error) {
	return l.prevPageToken()
}

func (l *Links) nextPageToken() (string, error) {
	if l == nil || l.Next == "" {
		return "", nil
	}
	token, err := pageTokenFromURL(l.Next)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (l *Links) prevPageToken() (string, error) {
	if l == nil || l.Prev == "" {
		return "", nil
	}
	token, err := pageTokenFromURL(l.Prev)
	if err != nil {
		return "", err
	}
	return token, nil
}

// IsLastPage returns true if the current page is the last
func (l *Links) IsLastPage() bool {
	return l.isLast()
}

func (l *Links) isLast() bool {
	return l.Next == ""
}

func pageForURL(urlText string) (int, error) {
	u, err := url.ParseRequestURI(urlText)
	if err != nil {
		return 0, err
	}

	pageStr := u.Query().Get("page")
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		return 0, err
	}

	return page, nil
}

func pageTokenFromURL(urlText string) (string, error) {
	u, err := url.ParseRequestURI(urlText)
	if err != nil {
		return "", err
	}
	return u.Query().Get("page_token"), nil
}