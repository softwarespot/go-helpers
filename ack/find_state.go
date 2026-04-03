package ack

import (
	"fmt"
	"regexp"
)

type findState struct {
	pattern *regexp.Regexp
	opts    FindOptions

	matches *Matches

	currLineNo int
	currMatch  *Match
	lastLineNo int

	maxCountRemaining int

	beforeContextBuffer   *beforeContextBuffer
	afterContextRemaining int
}

func newFindState(term string, opts FindOptions) (*findState, error) {
	if !opts.UseRegExp {
		term = regexp.QuoteMeta(term)

		// It's important to do this after quoting the meta characters; otherwise "(?i)" will become quoted too
		if !opts.UseCaseSensitive {
			term = fmt.Sprintf("(?i)%s", term)
		}
	}

	pattern, err := regexp.Compile(term)
	if err != nil {
		return nil, fmt.Errorf("invalid regular expression pattern of %q provided: %w", term, err)
	}

	return &findState{
		pattern: pattern,
		opts:    opts,

		matches: newMatches(pattern),

		currLineNo: 0,
		currMatch:  nil,
		lastLineNo: 0,

		maxCountRemaining: opts.MaxCount,

		beforeContextBuffer:   newBeforeContextBuffer(opts.BeforeContext),
		afterContextRemaining: 0,
	}, nil
}

func (fs *findState) handle(text string) bool {
	fs.currLineNo++

	if fs.pattern.MatchString(text) {
		return fs.handleMatch(text)
	}

	if fs.opts.BeforeContext > 0 {
		fs.beforeContextBuffer.add(
			&MatchContext{
				Line: fs.currLineNo,
				Text: text,
			},
		)
	}
	if fs.afterContextRemaining > 0 {
		fs.afterContextRemaining--
		fs.currMatch.AfterContext = append(
			fs.currMatch.AfterContext,
			&MatchContext{
				Line: fs.currLineNo,
				Text: text,
			},
		)
	}
	return true
}

// handleMatch returns false if no further lines should be processed
func (fs *findState) handleMatch(text string) bool {
	if fs.opts.MaxCount > 0 {
		if fs.maxCountRemaining == 0 {
			return false
		}
		fs.maxCountRemaining--
	}

	gapSize := fs.currLineNo - fs.lastLineNo - 1

	// Indicates there's a gap in the output i.e. when lines exist between matches that won't be shown
	// as either "before context" or "after context"
	ctxWindow := fs.opts.BeforeContext + fs.opts.AfterContext
	changedContext := fs.lastLineNo > 0 && ctxWindow > 0 && gapSize >= ctxWindow

	beforeContextRemaining := min(gapSize, fs.opts.BeforeContext)
	fs.afterContextRemaining = fs.opts.AfterContext

	fs.currMatch = &Match{
		Line:           fs.currLineNo,
		Text:           text,
		BeforeContext:  fs.beforeContextBuffer.lastN(beforeContextRemaining),
		AfterContext:   nil,
		ChangedContext: changedContext,
	}
	fs.matches.add(fs.currMatch)

	fs.lastLineNo = fs.currLineNo + fs.opts.AfterContext

	return true
}
