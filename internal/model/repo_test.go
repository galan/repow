package model

import (
	"repo/internal/say"
	"testing"
)

const dummyHost string = "blabla.com"

type matchCase struct {
	input    string
	expected string
}

var matchCases = []matchCase{
	{
		input: "origin	https://oauth2:ccc@" + dummyHost + "/group/services/some-service.git (fetch)",
		expected: "group/services/some-service",
	},
	{
		input: "origin	git@" + dummyHost + ":group/libraries/some-service.core.git (fetch)",
		expected: "group/libraries/some-service.core",
	},
	{
		input: "origin	git@" + dummyHost + ":group/infrastructure/project.git (fetch)",
		expected: "group/infrastructure/project",
	},
	{
		input: "origin	https://" + dummyHost + "/galan/maven-parent.git (fetch)",
		expected: "galan/maven-parent",
	},
	{
		input: "origin	git@" + dummyHost + ":group/infrastructure/project.git (fetch)",
		expected: "group/infrastructure/project",
	},
}

func TestMatches(t *testing.T) {
	say.VerboseEnabled = true
	for _, test := range matchCases {
		got := ParseRemotePath(test.input, dummyHost)
		if got != test.expected {
			t.Errorf("got %s, wanted %s", got, test.expected)
		}
	}
}
