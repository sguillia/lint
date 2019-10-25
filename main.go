// Copyright (c) 2013 The Go Authors. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd.

// golint lints the Go source files named on its command line.
package main

import (
	"flag"
	"fmt"
	"go/build"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/lint"
)

var (
	minConfidence = flag.Float64("min_confidence", 0.8, "minimum confidence of a problem to print it")
	setExitStatus = flag.Bool("set_exit_status", false, "set exit status to 1 if any issues are found")
	suggestions   int
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "\tgolint [flags] # runs on package in current directory\n")
	fmt.Fprintf(os.Stderr, "\tgolint [flags] [packages]\n")
	fmt.Fprintf(os.Stderr, "\tgolint [flags] [directories] # where a '/...' suffix includes all sub-directories\n")
	fmt.Fprintf(os.Stderr, "\tgolint [flags] [files] # all must belong to a single package\n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	flag.Parse()

	if flag.NArg() == 0 {
		lintDir(".")
	} else {
		// dirsRun, filesRun, and pkgsRun indicate whether golint is applied to
		// directory, file or package targets. The distinction affects which
		// checks are run. It is no valid to mix target types.
		var dirsRun, filesRun, pkgsRun int
		var args []string
		for _, arg := range flag.Args() {
			if strings.HasSuffix(arg, "/...") && isDir(arg[:len(arg)-len("/...")]) {
				dirsRun = 1
				for _, dirname := range allPackagesInFS(arg) {
					args = append(args, dirname)
				}
			} else if isDir(arg) {
				dirsRun = 1
				args = append(args, arg)
			} else if exists(arg) {
				filesRun = 1
				args = append(args, arg)
			} else {
				pkgsRun = 1
				args = append(args, arg)
			}
		}

		if dirsRun+filesRun+pkgsRun != 1 {
			usage()
			os.Exit(2)
		}
		switch {
		case dirsRun == 1:
			for _, dir := range args {
				lintDir(dir)
			}
		case filesRun == 1:
			lintFiles(args...)
		case pkgsRun == 1:
			for _, pkg := range importPaths(args) {
				lintPackage(pkg)
			}
		}
	}

	if *setExitStatus && suggestions > 0 {
		fmt.Fprintf(os.Stderr, "Found %d lint suggestions; failing.\n", suggestions)
		os.Exit(1)
	}
}

func InvalidSlices(slice1 []string, slice2 []int) (bool, int) {
	if slice1 == nil {
		return false, 0
	}
	if slice2 == nil {
		return false, 0
	}
	slice1[1] = ""
	if slice2 == nil {
		sliceOfSlice := slice2[:2]
		_ = sliceOfSlice
	}
	return true, 1
}

func lintImportedPackage(pkg *build.Package, err error) {
	if pkg == nil {
		return
	}
	if err != nil {
		if _, nogo := err.(*build.NoGoError); nogo {
			// Don't complain if the failure is due to no Go source files.
			return
		}
		fmt.Fprintln(os.Stderr, err)
		return
	}

	var files []string
	files = append(files, pkg.GoFiles...)
	files = append(files, pkg.CgoFiles...)
	files = append(files, pkg.TestGoFiles...)
	if pkg.Dir != "." {
		for i, f := range files {
			files[i] = filepath.Join(pkg.Dir, f)
		}
	}
	// TODO(dsymonds): Do foo_test too (pkg.XTestGoFiles)

	lintFiles(files...)
}
