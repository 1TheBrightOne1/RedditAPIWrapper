package api

func AttachParams(url string, params map[string]string) string {
	if len(params) == 0 {
		return url
	}

	url += "?"

	count := 0
	for key, val := range params {
		if val == "" {
			continue
		}
		count++
		url += key + "=" + val + "&"
	}

	//If no params were added because they were all blank
	if count == 0 {
		return url[:len(url)-1]
	}

	return url[:len(url)-1]
}
