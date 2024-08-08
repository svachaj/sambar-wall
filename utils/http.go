package utils

import (
	"errors"
	"net/url"
)

// GetQueryParamFromUrl extracts the value of a specified query parameter from a given URL.
func GetQueryParamFromUrl(rawUrl string, param string) (string, error) {
	// Parse the URL
	parsedUrl, err := url.Parse(rawUrl)
	if err != nil {
		return "", err
	}

	// Get the query parameters
	queryParams := parsedUrl.Query()

	// Get the value of the specified query parameter
	value := queryParams.Get(param)
	if value == "" {
		return "", errors.New("parameter not found")
	}

	return value, nil
}
