package pages

import "net/url"

type q map[string]string

func href(path string, params q) string {
	query := url.Values{}
	for k, v := range params {
		query.Add(k, v)
	}
	dest := &url.URL{Path: path, RawQuery: query.Encode()}
	return dest.String()
}
