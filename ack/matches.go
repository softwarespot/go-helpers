package ack

import (
	"fmt"
	"io"
	"regexp"
	"slices"

	"github.com/fatih/color"
)

// Matches represents the collection of matches found
type Matches struct {
	pattern *regexp.Regexp
	lines   []*Match
}

// Match represents a single match found in the text
type Match struct {
	Line           int             `json:"line"`
	Text           string          `json:"text"`
	BeforeContext  []*MatchContext `json:"beforeContext"`
	AfterContext   []*MatchContext `json:"afterContext"`
	ChangedContext bool            `json:"changedContext"`
}

// MatchContext represents the context lines around a match
type MatchContext struct {
	Line int    `json:"line"`
	Text string `json:"text"`
}

func newMatches(pattern *regexp.Regexp) *Matches {
	return &Matches{
		pattern: pattern,
		lines:   nil,
	}
}

// All returns a defensive copy of all matches to prevent external modification.
func (ms *Matches) All() []*Match {
	return slices.Clone(ms.lines)
}

// Len returns the number of matches.
func (ms *Matches) Len() int {
	return len(ms.lines)
}

func (ms *Matches) add(match *Match) {
	ms.lines = append(ms.lines, match)
}

// PrintOptions defines options for printing matches
type PrintOptions struct {
	LocationsWithMatches    bool
	LocationsWithoutMatches bool
	CountsOnly              bool
	IsPiped                 bool
	NoColor                 bool
}

// Print outputs matches in either colored or piped format based on options.
// It supports various output modes including location-only, counts, and full matches with context.
func (ms *Matches) Print(w io.Writer, location string, opts PrintOptions) error {
	hasMatches := len(ms.lines) > 0
	locationOnly := opts.LocationsWithMatches || opts.LocationsWithoutMatches

	if locationOnly && !opts.CountsOnly {
		shouldPrint := (opts.LocationsWithMatches && hasMatches) || (opts.LocationsWithoutMatches && !hasMatches)
		if shouldPrint {
			if _, err := fmt.Fprintln(w, location); err != nil {
				return fmt.Errorf("printing location: %w", err)
			}
		}
		return nil
	}

	if opts.CountsOnly {
		return ms.printCounts(w, location, hasMatches, locationOnly, opts)
	}

	if !hasMatches {
		return nil
	}

	if opts.IsPiped {
		return ms.printMatchesPiped(w, location)
	}
	return ms.printMatchesColored(w, location, opts)
}

func (ms *Matches) printCounts(w io.Writer, location string, hasMatches, locationOnly bool, opts PrintOptions) error {
	shouldSkip := locationOnly && ((opts.LocationsWithMatches && !hasMatches) || (opts.LocationsWithoutMatches && hasMatches))
	if shouldSkip {
		return nil
	}

	if opts.IsPiped || opts.NoColor {
		if _, err := fmt.Fprintf(w, "%s:%d\n", location, len(ms.lines)); err != nil {
			return fmt.Errorf("printing counts: %w", err)
		}
		return nil
	}

	colorMatchLocation := color.New(color.FgGreen, color.Bold)
	if _, err := fmt.Fprintf(w, "%s:%d\n", colorMatchLocation.Sprint(location), len(ms.lines)); err != nil {
		return fmt.Errorf("printing counts: %w", err)
	}
	return nil
}

func (ms *Matches) printMatchesPiped(w io.Writer, location string) error {
	for _, m := range ms.lines {
		if m.ChangedContext {
			if _, err := fmt.Fprintln(w, "--"); err != nil {
				return fmt.Errorf("printing context separator: %w", err)
			}
		}
		for _, bc := range m.BeforeContext {
			if _, err := fmt.Fprintf(w, "%s:%d-%s\n", location, bc.Line, bc.Text); err != nil {
				return fmt.Errorf("printing before context: %w", err)
			}
		}
		if _, err := fmt.Fprintf(w, "%s:%d:%s\n", location, m.Line, m.Text); err != nil {
			return fmt.Errorf("printing match: %w", err)
		}
		for _, ac := range m.AfterContext {
			if _, err := fmt.Fprintf(w, "%s:%d-%s\n", location, ac.Line, ac.Text); err != nil {
				return fmt.Errorf("printing after context: %w", err)
			}
		}
	}
	return nil
}

func (ms *Matches) printMatchesColored(w io.Writer, location string, opts PrintOptions) error {
	colorMatchLocation := color.New(color.FgGreen, color.Bold)
	colorMatchLineNo := color.New(color.FgYellow, color.Bold)
	colorContent := color.New(color.FgWhite)
	colorMatchContent := color.New(color.BgYellow, color.FgBlack)

	if opts.NoColor {
		colorMatchLocation.DisableColor()
		colorMatchLineNo.DisableColor()
		colorContent.DisableColor()
		colorMatchContent.DisableColor()
	}

	highlightMatch := func(s string) string {
		return colorMatchContent.Sprint(s)
	}

	if _, err := colorMatchLocation.Fprintln(w, location); err != nil {
		return fmt.Errorf("printing location: %w", err)
	}

	for _, m := range ms.lines {
		if m.ChangedContext {
			if _, err := colorContent.Fprintln(w, "--"); err != nil {
				return fmt.Errorf("printing context separator: %w", err)
			}
		}

		for _, bc := range m.BeforeContext {
			if _, err := colorContent.Fprintf(w, "%s-%s\n", colorMatchLineNo.Sprint(bc.Line), bc.Text); err != nil {
				return fmt.Errorf("printing before context: %w", err)
			}
		}

		if _, err := colorContent.Fprintf(w, "%s:%s\n", colorMatchLineNo.Sprint(m.Line), ms.pattern.ReplaceAllStringFunc(m.Text, highlightMatch)); err != nil {
			return fmt.Errorf("printing match: %w", err)
		}

		for _, ac := range m.AfterContext {
			if _, err := colorContent.Fprintf(w, "%s-%s\n", colorMatchLineNo.Sprint(ac.Line), ac.Text); err != nil {
				return fmt.Errorf("printing after context: %w", err)
			}
		}
	}
	return nil
}
