package main

import (
	"github.com/chromedp/cdproto/fetch"
)

func getYsResponseHeaders() []*fetch.HeaderEntry {
		return []*fetch.HeaderEntry {
			&fetch.HeaderEntry{
			  Name: "status",
			  Value: "200",
			},
			&fetch.HeaderEntry{
			  Name: "content-type",
			  Value: "text/html; charset=utf-8",
			},
			&fetch.HeaderEntry{
			  Name: "server-timing",
			  Value: "intid;desc=f0cea0fc94aed428",
			},
			&fetch.HeaderEntry{
			  Name: "x-frame-options",
			  Value: "SAMEORIGIN",
			},
			&fetch.HeaderEntry{
			  Name: "x-content-type-options",
			  Value: "nosniff",
			},
			&fetch.HeaderEntry{
			  Name: "x-request-id",
			  Value: "2424b7869338f400b3cdfe2aebd34e44",
			},
			&fetch.HeaderEntry{
			  Name: "content-encoding",
			  Value: "gzip",
			},
			&fetch.HeaderEntry{
			  Name: "x-akamai-transformed",
			  Value: "9 - 0 pmb=mTOE,1",
			},
			&fetch.HeaderEntry{
			  Name: "expires",
			  Value: "Fri, 13 Mar 2020 05:54:38 GMT",
			},
			&fetch.HeaderEntry{
			  Name: "cache-control",
			  Value: "max-age=0, no-cache, no-store",
			},
			&fetch.HeaderEntry{
			  Name: "pragma",
			  Value: "no-cache",
			},
			&fetch.HeaderEntry{
			  Name: "date",
			  Value: "Fri, 13 Mar 2020 05:54:38 GMT",
			},
			&fetch.HeaderEntry{
			  Name: "vary",
			  Value: "Accept-Encoding",
			},
			&fetch.HeaderEntry{
			  Name: "strict-transport-security",
			  Value: "max-age=15768000",
			},
		}
}
