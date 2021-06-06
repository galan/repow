package gitlab

import (
	"repo/internal/hoster"
	"testing"
)

type matchCase struct {
	path     string
	options  hoster.RequestOptions
	tags     []string
	expected bool
}

var matchCases = []matchCase{
	{
		path:     "my-group/foo/bar",
		options:  hoster.RequestOptions{},
		expected: true,
	},
	{
		path:     "my-group/foo/bar",
		options:  hoster.RequestOptions{IncludePatterns: []string{"^my-group/foo"}},
		expected: true,
	},
	{
		path:     "my-group/foo/bar",
		options:  hoster.RequestOptions{IncludePatterns: []string{"^my-group/bar"}},
		expected: false,
	},
	{
		path:     "my-group/foo/bar",
		options:  hoster.RequestOptions{IncludePatterns: []string{"^my-group", "foo"}},
		expected: true,
	},
	{
		path:     "my-group/foo/bar",
		options:  hoster.RequestOptions{IncludePatterns: []string{}, ExcludePatterns: []string{"bar"}},
		expected: false,
	},
	{
		path:     "my-group/foo/bar",
		options:  hoster.RequestOptions{IncludePatterns: []string{"^my-group"}, ExcludePatterns: []string{"bar"}},
		expected: false,
	},
	{
		path:     "my-group/foo/bar",
		options:  hoster.RequestOptions{IncludePatterns: []string{"^my-group"}, ExcludePatterns: []string{"baz", "boz"}},
		expected: true,
	},
	{
		path:     "my-group/foo/bar",
		options:  hoster.RequestOptions{IncludePatterns: []string{"^my-group", "bar$"}, ExcludePatterns: []string{"foo", "bar"}},
		expected: false,
	},
}

func TestMatches(t *testing.T) {
	for _, test := range matchCases {
		got := matches(test.options, test.path, test.tags)
		if got != test.expected {
			t.Errorf("got %t, wanted %t for %v", got, test.expected, test)
		}
	}
}
