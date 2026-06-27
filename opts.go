package gii

import "net/http"

type GiiClientOption func(*GiiClient)

func WithHTTPClient(httpClient *http.Client) GiiClientOption {
	return func(client *GiiClient) {
		if httpClient != nil {
			client.httpclient = httpClient
		}
	}
}

func WithBaseAddress(baseAddress string) GiiClientOption {
	return func(client *GiiClient) {
		if baseAddress != "" {
			client.baseAddress = baseAddress
		}
	}
}

func WithWorkDir(workDir string) GiiClientOption {
	return func(client *GiiClient) {
		if workDir != "" {
			client.workDir = workDir
		}
	}
}
