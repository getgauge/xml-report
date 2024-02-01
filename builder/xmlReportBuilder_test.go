/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package builder

import (
	"testing"

	"path/filepath"

	"github.com/getgauge/gauge-proto/go/gauge_messages"
	. "gopkg.in/check.v1"
)

func Test(t *testing.T) {
	TestingT(t)
}

type MySuite struct{}

var _ = Suite(&MySuite{})

func (s *MySuite) TestGetSpecNameWhenHeadingIsPresent(c *C) {
	want := "heading"

	got := getSpecName(&gauge_messages.ProtoSpec{SpecHeading: "heading"})

	c.Assert(want, Equals, got)
}

func (s *MySuite) TestGetSpecNameWhenHeadingIsNotPresent(c *C) {
	want := "example.spec"

	got := getSpecName(&gauge_messages.ProtoSpec{FileName: filepath.Join("specs", "specs1", "example.spec")})

	c.Assert(want, Equals, got)
}

func (s *MySuite) TestHasParseErrors(c *C) {
	errors := []*gauge_messages.Error{
		{Type: gauge_messages.Error_PARSE_ERROR},
		{Type: gauge_messages.Error_VALIDATION_ERROR},
	}

	got := hasParseErrors(errors)

	c.Assert(true, Equals, got)
}

func (s *MySuite) TestHasParseErrorsWithNoErrors(c *C) {
	errors := []*gauge_messages.Error{}

	got := hasParseErrors(errors)

	c.Assert(false, Equals, got)
}

func (s *MySuite) TestHasParseErrorsWithOnyValidationErrors(c *C) {
	errors := []*gauge_messages.Error{
		{Type: gauge_messages.Error_VALIDATION_ERROR},
		{Type: gauge_messages.Error_VALIDATION_ERROR},
	}

	got := hasParseErrors(errors)

	c.Assert(false, Equals, got)
}

func (s *MySuite) TestGetErrorTestCase(c *C) {
	res := &gauge_messages.ProtoSpecResult{
		ProtoSpec: &gauge_messages.ProtoSpec{
			SpecHeading: "heading",
		},
		Errors: []*gauge_messages.Error{
			{
				Type:    gauge_messages.Error_PARSE_ERROR,
				Message: "parse error",
			},
			{
				Type:    gauge_messages.Error_VALIDATION_ERROR,
				Message: "validation error",
			},
		},
		ExecutionTime: int64(1),
	}

	want := JUnitTestCase{
		Classname: "heading",
		Name:      "heading",
		Time:      formatTime(int(1)),
		Failure: &JUnitFailure{
			Message:  "Parse/Validation Errors",
			Type:     "Parse/Validation Errors",
			Contents: "[Parse Error] parse error\n[Validation Error] validation error",
		},
	}

	got := getErrorTestCase(res)

	c.Assert(want, DeepEquals, got)
}
