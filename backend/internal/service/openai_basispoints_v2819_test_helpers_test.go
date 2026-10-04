package service

import "net/http"

type bps2819TestUpstream struct {
	httpUpstreamRecorder
	send func(*http.Request, string) (*http.Response, error)
}

func (u *bps2819TestUpstream) Do(r *http.Request, proxy string, _ int64, _ int) (*http.Response, error) {
	return u.send(r, proxy)
}
