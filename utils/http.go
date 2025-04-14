package utils

import (
	"encoding/csv"
	"net/http"
	"net/url"

	"github.com/labstack/echo/v4"
)

// GetQueryParamFromUrl extracts the value of a specified query parameter from a given URL.
func GetQueryParamFromUrl(rawUrl string, param string) string {
	// Parse the URL
	parsedUrl, err := url.Parse(rawUrl)
	if err != nil {
		return ""
	}

	// Get the query parameters
	queryParams := parsedUrl.Query()

	// Get the value of the specified query parameter
	value := queryParams.Get(param)

	return value
}

// SetBackUrlAndGetQueryParamFromUrl sets the backUrl and extracts the value of a specified query parameter from a given URL.
// return backUrl, searchParam, error
func SetBackUrlAndGetQueryParamFromUrl(c echo.Context, paramKey, defaultBackUrlValue string) (string, string, error) {

	backUrl := c.FormValue("backUrl")
	searchParam := ""
	if backUrl == "" {
		backUrl = defaultBackUrlValue
	}
	// try to get search param from the backUrl
	searchParam = GetQueryParamFromUrl(backUrl, paramKey)

	c.Response().Header().Set("HX-Push-Url", backUrl)

	return backUrl, searchParam, nil
}

// WriteCSV writes the given data to the response writer in CSV format.
func WriteCSV(w http.ResponseWriter, data [][]string) error {
	writer := csv.NewWriter(w)
	defer writer.Flush()

	for _, record := range data {
		if err := writer.Write(record); err != nil {
			return err
		}
	}
	return nil
}
