package main_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func test_with_hscript(t *testing.T, t_name, content string, t_err error) {

	workdir, err := ioutil.TempDir("", "hwaf-test-")
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer os.RemoveAll(workdir)

	err = os.Chdir(workdir)
	if err != nil {
		t.Fatalf(err.Error())
	}

	hwaf, err := newlogger("hwaf.log")
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer hwaf.Close()

	for _, cmd := range [][]string{
		{"hwaf", "init", "-v=1", "."},
		{"hwaf", "setup", "-v=1"},
		{"hwaf", "pkg", "create", "-script=hscript", "-v=1", "mypkg"},
		{"hwaf", "pkg", "ls"},
	} {
		err := hwaf.Run(cmd[0], cmd[1:]...)
		if err != nil {
			hwaf.Display()
			t.Fatalf("cmd %v failed: %v", cmd, err)
		}
	}

	mypkgdir := filepath.Join("src", "mypkg")

	// hscript.yml file
	ff, err := os.Create(filepath.Join(mypkgdir, "hscript.yml"))
	if err != nil {
		t.Fatalf(err.Error())
	}
	_, err = ff.WriteString(content)
	if err != nil {
		t.Fatalf(err.Error())
	}

	ff.Sync()
	ff.Close()

	err = os.MkdirAll(filepath.Join(mypkgdir, "waftools"), 0777)
	if err != nil {
		t.Fatalf(err.Error())
	}

	// waftools/script-1.py
	ff, err = os.Create(filepath.Join(mypkgdir, "waftools", "script-1.py"))
	if err != nil {
		t.Fatalf(err.Error())
	}
	_, err = ff.WriteString(`
## -*- python -*-
def no_configure(ctx):
    pass

def no_build(ctx):
    pass

`)
	if err != nil {
		t.Fatalf(err.Error())
	}

	ff.Sync()
	ff.Close()

	// waftools/script-2.py
	ff, err = os.Create(filepath.Join(mypkgdir, "waftools", "script-2.py"))
	if err != nil {
		t.Fatalf(err.Error())
	}
	_, err = ff.WriteString(`
## -*- python -*-
import waflib.Logs as msg

def configure(ctx):
    msg.info("tool script-2 loaded from configure")
    pass

def build(ctx):
    msg.info("tool script-1 loaded from configure")
    pass

`)
	if err != nil {
		t.Fatalf(err.Error())
	}

	ff.Sync()
	ff.Close()

	for _, cmd := range [][]string{
		{"hwaf", "configure"},
		//{"hwaf"},
	} {
		err := hwaf.Run(cmd[0], cmd[1:]...)
		if err != nil && t_err == nil {
			hwaf.Display()
			t.Fatalf("in test [%s] cmd %v failed: %v", t_name, cmd, err)
		}
	}
}

func TestHwafHscriptSections(t *testing.T) {

	//var no_error error = nil
	var w_error error = fmt.Errorf("error expected")

	for _, tt := range []struct {
		name     string
		content  string
		expected error
	}{
		{
			name: "empty hscript",
			content: `
# -*- yaml -*-

`,
			expected: w_error,
		},
		{
			name: "empty content hscript",
			content: `
# -*- yaml -*-
package: {}
options: {}
configure: {}
build: {}
`,
			expected: w_error,
		},
		{
			name: "valid empty content hscript",
			content: `
# -*- yaml -*-
package: {name: "mypkg"}
options: {}
configure: {}
build: {}
`,
			expected: nil,
		},
		{
			name: "mispelled packages",
			content: `
# -*- yaml -*-
packaGes: {
}
`,
			expected: w_error,
		},
		{
			name: "missing packages content",
			content: `
# -*- yaml -*-
package: {
}
`,
			expected: w_error,
		},
		{
			name: "misspelled packages content",
			content: `
# -*- yaml -*-
package: {
 naMe: [],
}
`,
			expected: w_error,
		},
		{
			name: "miss-type packages content",
			content: `
# -*- yaml -*-
package: {
 name: ["mypkg"],
}
`,
			expected: w_error,
		},
		{
			name: "miss-spelled configure content (environ)",
			content: `
# -*- yaml -*-
package: {
 name: "mypkg",
 authors: "me",
 deps: {
 },
}

options: {}

configure: {
 tools: [],
 environ: {},
}
`,
			expected: w_error,
		},
		{
			name: "miss-type configure content (declare-tags)",
			content: `
# -*- yaml -*-
package: {
 name: "mypkg",
 authors: "me",
 deps: {
 },
}

options: {}

configure: {
 tools: [],
 env: {},
 declare-tags: {},
}
`,
			expected: w_error,
		},
		{
			name: "miss-type configure content (apply-tags)",
			content: `
# -*- yaml -*-
package: {
 name: "mypkg",
 authors: "me",
 deps: {
 },
}

options: {}

configure: {
 tools: [],
 env: {},
 declare-tags: [],
 apply-tags: {},
}
`,
			expected: w_error,
		},
		{
			name: "valid hscript",
			content: `
# -*- yaml -*-

## with comments
package: {
 ## this is the name of the package
 name: "mypkg",
 authors: ["me", "you", "everybody"],
 managers: ["somebody", "higher being"],

 deps: {
 },
}

## options declared to the command line
options: {
 tools: ["compiler_cxx", "find_python"],
}

configure: {
 tools: ["compiler_cxx", "find_python"],
 env: {
  MYPATH: "/some/path",
  PREPENDPATH: "/mypath/python:${PREPENDPATH}",
  APPENDPATH:  "${APPENDPATH}:/mypath/python",
 },
 declare-tags: [
  {x86_64-slc5-gcc48-opt: [x86_64, linux, slc5, 64b, gcc48, gcc, opt]},
  {x86_64-slc6-gcc48-opt: [x86_64, linux, slc5, 64b, gcc48, gcc, opt]},
  {my-graphics-tag: []},
 ],
 apply-tags: [
   x86_64-slc6-gcc48-opt,
   my-graphics-tag,
 ],

 hwaf-call: [
  "waftools/script-1.py",
  "waftools/script-2.py",
 ],
}

build: {
 hwaf-call: [
  "waftools/script-1.py",
  "waftools/script-2.py",
 ],

 cxx-hello-world: {
   features: "cxx cxxshlib",
   source:   "src/mypkgtool.cxx",
   target:   "hello-world",
 },

 cxx-hello-app: {
   features: [cxx, cxxprogram],
   source:   src/myapp.cxx,
   target:   hello-app,
   use:      [cxx-hello-world],
   cxxflags: [-O3],
   defines:  [MYDEFINE=1, NDEBUG],
   cflags:   -g,
   install_path: "${INSTALL_AREA}/share/bin",
 },
}
`,
			expected: nil,
		},
	} {
		test_with_hscript(t, tt.name, tt.content, tt.expected)
	}
}

// EOF
