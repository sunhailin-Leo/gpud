package tail

import (
	"errors"
	"time"

	query_log_common "github.com/leptonai/gpud/pkg/query/log/common"
)

type Op struct {
	labels map[string]string

	file     string
	commands [][]string

	linesToTail int
	dedup       bool

	perLineFunc func([]byte)

	matchFuncs    []query_log_common.MatchFunc
	selectFilters []*query_log_common.Filter
	rejectFilters []*query_log_common.Filter

	extractTime   query_log_common.ExtractTimeFunc
	skipEmptyLine bool

	ProcessMatched query_log_common.ProcessMatchedFunc
}

type OpOption func(*Op)

func (op *Op) ApplyOpts(opts []OpOption) error {
	for _, opt := range opts {
		opt(op)
	}

	if op.labels == nil {
		op.labels = make(map[string]string)
	}

	if op.file == "" && len(op.commands) == 0 {
		return errors.New("file or commands must be set")
	}

	if op.linesToTail == 0 {
		op.linesToTail = 100
	}

	if len(op.selectFilters) > 0 && len(op.rejectFilters) > 0 {
		return errors.New("cannot set both select and reject filters")
	}
	for i := range op.selectFilters {
		if err := op.selectFilters[i].Compile(); err != nil {
			return err
		}
	}
	for i := range op.rejectFilters {
		if err := op.rejectFilters[i].Compile(); err != nil {
			return err
		}
	}

	if op.extractTime == nil {
		op.extractTime = func(line []byte) (time.Time, []byte, error) {
			return time.Time{}, line, nil
		}
	}
	if op.ProcessMatched == nil {
		op.ProcessMatched = func(time.Time, []byte, *query_log_common.Filter) {}
	}

	return nil
}

func WithLabel(key, value string) OpOption {
	return func(op *Op) {
		if op.labels == nil {
			op.labels = make(map[string]string)
		}
		op.labels[key] = value
	}
}

func WithFile(file string) OpOption {
	return func(op *Op) {
		op.file = file
	}
}

func WithCommands(commands [][]string) OpOption {
	return func(op *Op) {
		op.commands = commands
	}
}

// Sets the number of lines to tail.
// If not set, defaults to 100.
func WithLinesToTail(n int) OpOption {
	return func(op *Op) {
		op.linesToTail = n
	}
}

// If true, dedup lines by the log line string.
// This is useful for logs that have the same message
// repeated multiple times with the same timestamp.
func WithDedup(dedup bool) OpOption {
	return func(op *Op) {
		op.dedup = dedup
	}
}

// Called for each line.
func WithPerLineFunc(f func([]byte)) OpOption {
	return func(op *Op) {
		op.perLineFunc = f
	}
}

// "OR" conditions to select logs.
// The line is sent when any of the match function returns non-empty strings.
func WithMatchFunc(matchFuncs ...query_log_common.MatchFunc) OpOption {
	return func(op *Op) {
		if len(matchFuncs) > 0 {
			op.matchFuncs = append(op.matchFuncs, matchFuncs...)
		}
	}
}

// "OR" conditions to select logs.
//
// The line is sent when any of the filters match.
// Useful for explicit blacklisting "error" logs
// (e.g., GPU error messages in dmesg).
func WithSelectFilter(filters ...*query_log_common.Filter) OpOption {
	return func(op *Op) {
		if len(filters) > 0 {
			op.selectFilters = append(op.selectFilters, filters...)
		}
	}
}

// "AND" conditions to exclude logs.
//
// The line is sent if and only if all of the filters do not match.
// Useful for explicit whitelisting logs and catch all other
// (e.g., good healthy log messages).
func WithRejectFilter(filters ...*query_log_common.Filter) OpOption {
	return func(op *Op) {
		if len(filters) > 0 {
			op.rejectFilters = append(op.rejectFilters, filters...)
		}
	}
}

func (op *Op) applyFilter(line any) (shouldInclude bool, matchedFilter *query_log_common.Filter, err error) {
	if len(op.matchFuncs) == 0 && len(op.selectFilters) == 0 && len(op.rejectFilters) == 0 {
		// no filters
		return true, nil, nil
	}

	for _, matchFunc := range op.matchFuncs {
		var eventName string
		switch line := line.(type) {
		case string:
			eventName, _ = matchFunc(line)
		case []byte:
			eventName, _ = matchFunc(string(line))
		}
		if eventName != "" {
			filter := &query_log_common.Filter{
				Name: eventName,
			}
			return true, filter, nil
		}
	}
	if len(op.selectFilters) == 0 && len(op.rejectFilters) == 0 {
		return false, nil, nil
	}

	// blacklist (e.g., error logs)
	for _, filter := range op.selectFilters {
		// assume regex is already compiled
		var matched bool
		switch line := line.(type) {
		case string:
			matched, err = filter.MatchString(line)
		case []byte:
			matched, err = filter.MatchBytes(line)
		}
		if err != nil { // regex has not been compiled
			return false, nil, err
		}
		if matched {
			matchedFilter = filter
			break
		}
	}
	if len(op.selectFilters) > 0 && matchedFilter == nil {
		// select filter non-empty, and the line didn't pass any
		// thus should not be included
		return false, nil, nil
	}

	// whitelist (e.g., good logs)
	rejected := false
	for _, filter := range op.rejectFilters {
		// assume regex is already compiled
		var matched bool
		switch line := line.(type) {
		case string:
			matched, err = filter.MatchString(line)
		case []byte:
			matched, err = filter.MatchBytes(line)
		}
		if err != nil { // regex has not been compiled
			return false, nil, err
		}
		if matched {
			rejected = true
			break
		}
	}

	if rejected {
		// means, the line matches a good log line regex
		// thus should not be marked as an event
		return false, nil, nil
	}

	return true, matchedFilter, nil
}

func WithExtractTime(f query_log_common.ExtractTimeFunc) OpOption {
	return func(op *Op) {
		if f != nil {
			op.extractTime = f
		}
	}
}

func WithSkipEmptyLine(skipEmptyLine bool) OpOption {
	return func(op *Op) {
		op.skipEmptyLine = skipEmptyLine
	}
}

// Called if the line is matched.
// If not set, the matched line is no-op.
// Useful to append to a slice or not to return a string slice
// to avoid extra heap allocation.
func WithProcessMatched(f query_log_common.ProcessMatchedFunc) OpOption {
	return func(op *Op) {
		if f != nil {
			op.ProcessMatched = f
		}
	}
}
