// nolint
package traefik_plugin_insert_custom_header

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHeaderMutator(t *testing.T) {
	tests := []struct {
		desc          string
		mutations     []Mutation
		reqHeader     http.Header
		expRespHeader http.Header
	}{
		{
			desc: "should replace foo by bar in location header",
			mutations: []Mutation{
				{
					Header:      "Location",
					Regex:       "foo",
					Replacement: "bar",
				},
			},
			reqHeader: map[string][]string{
				"Location": {"foo", "anotherfoo"},
			},
			expRespHeader: map[string][]string{
				"Location": {"bar", "anotherbar"},
			},
		},
		{
			desc: "should replace http by https in location header",
			mutations: []Mutation{
				{
					Header:      "Location",
					Regex:       "^http://(.+)$",
					Replacement: "https://$1",
				},
			},
			reqHeader: map[string][]string{
				"Location": {"http://test:1000"},
			},
			expRespHeader: map[string][]string{
				"Location": {"https://test:1000"},
			},
		},
		{
			desc: "should clone the header with a new name",
			mutations: []Mutation{
				{
					Header:      "Host",
					NewName:     "X-Host",
					Regex:       "^(.+)$",
					Replacement: "$1",
				},
			},
			reqHeader: map[string][]string{
				"host": {"example.com"},
			},
			expRespHeader: map[string][]string{
				"Host":   {"example.com"},
				"X-Host": {"example.com"},
			},
		},
		{
			desc: "should create a new header with a new name and modified value",
			mutations: []Mutation{
				{
					Header:      "Host",
					NewName:     "X-Host",
					Regex:       "^(.+)\\.test\\.com$",
					Replacement: "$1",
				},
			},
			reqHeader: map[string][]string{
				"host": {"example.com.test.com"},
			},
			expRespHeader: map[string][]string{
				"Host":   {"example.com.test.com"},
				"X-Host": {"example.com"},
			},
		},
		{
			desc: "should rename the header with modified value",
			mutations: []Mutation{
				{
					Header:      "Host",
					NewName:     "X-Host",
					Regex:       "^(.+)\\.test\\.com$",
					Replacement: "$1",
				},
			},
			reqHeader: map[string][]string{
				"host": {"example.com.test.com"},
			},
			expRespHeader: map[string][]string{
				"X-Host": {"example.com"},
			},
		},
		{
			desc: "should rename the header",
			mutations: []Mutation{
				{
					Header:  "Host",
					NewName: "X-Host",
				},
			},
			reqHeader: map[string][]string{
				"host": {"example.com"},
			},
			expRespHeader: map[string][]string{
				"X-Host": {"example.com"},
			},
		},
		{
			desc: "should clone the header",
			mutations: []Mutation{
				{
					Header:  "Host",
					NewName: "X-Host",
				},
			},
			reqHeader: map[string][]string{
				"host": {"example.com"},
			},
			expRespHeader: map[string][]string{
				"host":   {"example.com"},
				"X-Host": {"example.com"},
			},
		},
		{
			desc: "preserve the value if regex does not match",
			mutations: []Mutation{
				{
					Header:      "Host",
					NewName:     "X-Host",
					Regex:       "^(.+)\\.test\\.com$",
					Replacement: "$1",
				},
			},
			reqHeader: map[string][]string{
				"host": {"example.com"},
			},
			expRespHeader: map[string][]string{
				"host":   {"example.com"},
				"X-Host": {"example.com"},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			config := &Config{
				Mutations: test.mutations,
			}

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
			mutator, err := New(context.Background(), next, config, "mutateHeaders")
			if err != nil {
				t.Fatal(err)
			}

			r := httptest.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()

			for k, v := range test.reqHeader {
				for _, h := range v {
					r.Header.Add(k, h)
				}
			}

			mutator.ServeHTTP(w, r)
			for k, expected := range test.expRespHeader {
				values := r.Header.Values(k)

				if !testEq(values, expected) {
					t.Errorf("Slice arent equals: expect: %+v, result: %+v", expected, values)
				}
			}
		})
	}
}

func testEq(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
