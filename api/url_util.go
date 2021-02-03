package api

func AttachParams(url string, params map[string]string) string {
	if len(params) == 0 {
		return url
	}

	url += "?"

	for key, val := range params {
		url += key + "=" + val + "&"
	}

	return url[:len(url)-1]
}
