// Package ack provides grep/ack-like text search functionality with support for
// regular expressions, case-insensitive matching, and context lines around matches.
package ack

import (
	"bufio"
	"fmt"
	"io"
)

// FindOptions defines options for finding matches
type FindOptions struct {
	UseCaseSensitive bool
	UseRegExp        bool
	MaxCount         int
	BeforeContext    int
	AfterContext     int
	BufferSize       int
}

// Find searches for all occurrences of term in the input reader.
// It returns a Matches object containing all found matches with optional context lines.
// The search behavior is controlled by opts, supporting case-sensitive/insensitive search,
// literal text or regular expression matching, and context lines before/after matches.
func Find(r io.Reader, term string, opts FindOptions) (*Matches, error) {
	fs, err := newFindState(term, opts)
	if err != nil {
		return nil, err
	}

	s := bufio.NewScanner(r)

	if opts.BufferSize > 0 {
		buf := make([]byte, 0, opts.BufferSize)
		s.Buffer(buf, cap(buf))
	}

	for s.Scan() {
		if !fs.handle(s.Text()) {
			break
		}
	}
	if s.Err() != nil {
		return nil, fmt.Errorf("error scanning input: %w", s.Err())
	}
	return fs.matches, nil
}
