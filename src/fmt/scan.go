// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fmt

import (
	"errors"
	"io"
	"math"
	"os"
	"reflect"
	"strconv"
	"sync"
	"unicode/utf8"
)

// doScanf does the real work when scanning with a format string.
// At the moment, it handles only pointers to basic types.
func (s *ss) doScanf(format string, a []interface{}) (numProcessed int, err error) {
	defer errorHandler(&err)
	end := len(format) - 1
	// We process one item per non-trivial format
	for i := 0; i <= end; {
		w := s.advance(format[i:])
		if w > 0 {
			i += w
			continue
		}
		// Either we failed to advance, we have a percent character, or we ran out of input.
		if format[i] != '%' {
			// Can't advance format. Why not?
			if w < 0 {
				s.errorString("input does not match format")
			}
			// Otherwise at EOF; "too many operands" error handled below
			break
		}
		i++ // % is one byte

		// do we have 20 (width)?
		var widPresent bool
		s.maxWid, widPresent, i = parsenum(format, i, end)
		if !widPresent {
			s.maxWid = hugeWid
		}

		c, w := utf8.DecodeRuneInString(format[i:])
		i += w

		if c != 'c' {
			s.SkipSpace()
		}
		if c == '%' {
			s.scanPercent()
			continue // Do not consume an argument.
		}
		s.argLimit = s.limit
		if f := s.count + s.maxWid; f < s.argLimit {
			s.argLimit = f
		}

		if numProcessed >= len(a) { // out of operands
			s.errorString("too few operands for format '%" + format[i-w:] + "'")
			break
		}
		arg := a[numProcessed]

		s.scanOne(c, arg)
		numProcessed++
		s.argLimit = s.limit
	}
	if numProcessed < len(a) {
		s.errorString("too many operands")
	}
	return
}
