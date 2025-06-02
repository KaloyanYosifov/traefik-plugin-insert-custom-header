// nolint
package traefik_plugin_insert_custom_header

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
)

// Mutation holds one mutation body configuration.
type Mutation struct {
	Header      string `json:"header,omitempty"`
	NewName     string `json:"newName,omitempty"`
	Regex       string `json:"regex,omitempty"`
	Replacement string `json:"replacement,omitempty"`
}

type FromUrlMutation struct {
	NewName     string `json:"header,omitempty"`
	Regex       string `json:"regex,omitempty"`
	Replacement string `json:"replacement,omitempty"`
}

// Config holds the plugin configuration.
type Config struct {
	Mutations        []Mutation        `json:"mutations,omitempty"`
	FromUrlMutations []FromUrlMutation `json:"fromUrlMutations,omitempty"`
}

// CreateConfig creates and initializes the plugin configuration.
func CreateConfig() *Config {
	return &Config{}
}

type mutation struct {
	oldName      string
	newName      string
	deleteSource bool
	mutate       bool
	regex        *regexp.Regexp
	replacement  string
}

type fromUrlMutation struct {
	newName     string
	regex       *regexp.Regexp
	replacement string
}

type HeaderMutator struct {
	name             string
	next             http.Handler
	mutations        []mutation
	fromUrlMutations []fromUrlMutation
}

// New creates and returns a new HeaderMutator plugin.
func New(_ context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	mutations := make([]mutation, len(config.Mutations))

	for i, m := range config.Mutations {
		mt := mutation{oldName: m.Header, newName: m.NewName}
		if m.Regex != "" {
			regex, err := regexp.Compile(m.Regex)
			if err != nil {
				return nil, fmt.Errorf("error compiling regex %q: %w", m.Regex, err)
			}
			if m.Replacement == "" {
				return nil, fmt.Errorf("replacement is required when regex is set")
			}
			mt.mutate = true
			mt.regex = regex
			mt.replacement = m.Replacement
		} else {
			mt.mutate = false
		}

		mutations[i] = mt
	}

	fromUrlMutations := make([]fromUrlMutation, len(config.FromUrlMutations))
	for i, m := range config.FromUrlMutations {
		mt := fromUrlMutation{newName: m.NewName}
		if m.Regex == "" {
			return nil, fmt.Errorf("regex for url mutation cannot be empty")
		}

		regex, err := regexp.Compile(m.Regex)
		if err != nil {
			return nil, fmt.Errorf("error compiling regex %q: %w", m.Regex, err)
		}

		if m.Replacement == "" {
			return nil, fmt.Errorf("replacement is required")
		}

		mt.regex = regex
		mt.replacement = m.Replacement

		fromUrlMutations[i] = mt
	}

	return &HeaderMutator{
		name:             name,
		next:             next,
		mutations:        mutations,
		fromUrlMutations: fromUrlMutations,
	}, nil
}

func (h *HeaderMutator) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	for _, m := range h.mutations {
		// skip if it's not a mutation or cloning
		if m.newName == "" && !m.mutate {
			continue
		}

		// if the header is not present, skip
		headerValues := req.Header.Values(m.oldName)
		if len(headerValues) == 0 {
			continue
		}

		// delete/rename case
		if m.deleteSource {
			req.Header.Del(m.oldName)
		}

		// if new name is not set consider it as a in-place mutation
		if m.newName == "" {
			m.newName = m.oldName
		}
		// clean the old header values
		req.Header.Del(m.newName)

		for _, v := range headerValues {
			if m.mutate {
				mv := m.regex.ReplaceAllString(v, m.replacement)
				if mv != "" {
					req.Header.Add(m.newName, mv)
				} else {
					req.Header.Add(m.newName, v)
				}
			} else {
				req.Header.Add(m.newName, v)
			}
		}
	}

	for _, m := range h.fromUrlMutations {
		v := req.URL.String()
		mv := m.regex.ReplaceAllString(v, m.replacement)

		if mv != "" {
			req.Header.Add(m.newName, mv)
		} else {
			req.Header.Add(m.newName, v)
		}
	}

	h.next.ServeHTTP(rw, req)
}
