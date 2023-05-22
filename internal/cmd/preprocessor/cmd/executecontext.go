// Copyright 2023 The CUE Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"cuelang.org/go/cue/load"
)

type executeContext struct {
	// executor is the underlying executor behind this context
	executor *executor

	// filter is a set of absolute path directories that should be considered as
	// part of an execute step. If filter is nil, then all directories are
	// considered.
	filter map[string]bool

	// order captures the order in which we discovered index pages. That is,
	// order contains the directories in the order in which we walked the
	// directory to discover the page roots.  It provides a consistent order for
	// the subsequent processing of pages.
	order []string

	// pages is a set of all the pages discovered by the executor.  The string
	// key is the full directory path of the page source.
	pages map[string]*page

	errorContext
	*executionContext
}

func (ec *executeContext) Format(state fmt.State, verb rune) {
	fmt.Fprintf(state, "%v", ec.executor)
}

func (e *executor) newExecuteContext(filter map[string]bool) *executeContext {
	return &executeContext{
		executor:         e,
		pages:            make(map[string]*page),
		filter:           filter,
		executionContext: e.executionContext,
		errorContext:     e.errorContext,
	}
}

func (ec *executeContext) execute() error {
	// Recursively walk wd to find $lang.md and _$lang.md files for all
	// supported languages.
	if err := ec.findPages(); err != nil {
		return err
	}

	// Load all the CUE in one go
	cfg := &load.Config{
		Dir: ec.executor.root,
	}
	var insts []string
	for _, d := range ec.order {
		// Find all CUE files
		//
		// TODO: filter for only those that are part of the site package? Maybe
		// later
		entries, err := os.ReadDir(d)
		if err != nil {
			return ec.errorf("%v: failed to read %s: %v", ec, d, err)
		}
		for _, e := range entries {
			n := e.Name()
			if filepath.Ext(n) == ".cue" {
				insts = append(insts, filepath.Join(d, n))
			}
		}
	}
	bps := load.Instances(insts, cfg)
	if l := len(bps); l != 1 {
		return ec.errorf("%v: expected 1 build instance; saw %d", ec, l)
	}
	v := ec.executor.ctx.BuildInstance(bps[0])
	if err := v.Err(); err != nil {
		return ec.errorf("%v: error in site configuration: %v", ec, err)
	}
	ec.config = v

	var pageWaits []<-chan struct{}

	// Process the pages we found.
	for _, d := range ec.order {
		p := ec.pages[d]
		// v := vs[i]
		// if err := v.Validate(); err != nil {
		// 	return fmt.Errorf("failed to validate CUE package %s: %w", p.relPath, err)
		// }
		done := make(chan struct{})
		go func() {
			p.process()
			close(done)
		}()
		pageWaits = append(pageWaits, done)
	}

	for i, d := range ec.order {
		p := ec.pages[d]
		w := pageWaits[i]
		<-w
		ec.updateInError(p.isInError())
		ec.logf("%s", p.bytes())
	}

	return errorIfInError(ec)
}

func (ec *executeContext) findPages() error {
	dirsToWalk := []string{ec.executor.wd}

	var dir string
	for len(dirsToWalk) > 0 {
		dir, dirsToWalk = dirsToWalk[0], dirsToWalk[1:]

		entries, err := os.ReadDir(dir)
		if err != nil {
			return ec.errorf("%v: failed to read dir %s: %v", ec, dir, err)
		}

		searchForRootFiles := ec.filter == nil || ec.filter[dir]

		var p *page

		for _, entry := range entries {
			name := entry.Name()

			// Add child directories to our list of dirs to walk if they don't
			// begin with '.' or '_'.
			if entry.IsDir() {
				if name[0] != '_' && name[0] != '.' {
					dirsToWalk = append(dirsToWalk, filepath.Join(dir, name))
				}
				continue
			}

			// name is now know to be a file. Only proceed if the filter
			// was nil or dir matched the filter
			if !searchForRootFiles {
				continue
			}

			// Is the file path a page root?
			match := pageRootFileRegexp.FindStringSubmatch(name)
			if match == nil {
				continue
			}

			// We found a page root! Extract key parts of the match
			prefix := match[1]
			lang := lang(match[2])
			ext := match[3]

			if p == nil {
				// Keep the order so we process the resulting pages in the same order.
				// For now this doesn't really serve any specific purpose, other than to
				// make the result of execute deterministic order-wise.
				ec.order = append(ec.order, dir)

				relDir, err := filepath.Rel(ec.executor.wd, dir)
				if err != nil {
					return ec.errorf("%v: failed to determine %s relative to %s: %v", ec, dir, ec.executor.wd, err)
				}

				p, err = ec.newPage(dir, relDir)
				if err != nil {
					return ec.errorf("%v: failed to create page in dir %s: %v", ec, dir, err)
				}
				ec.debugf("%v: found page at %v", ec, dir)
				ec.pages[dir] = p
			}

			// Register the root file with the page
			rf := p.newRootFile(name, lang, prefix, ext)
			p.debugf("%v: added root file named %s", p, name)
			p.rootFiles[name] = rf
			rootPages := p.langTargets[lang]
			rootPages = append(rootPages, rf)
			p.langTargets[lang] = rootPages
		}
	}

	return nil
}
