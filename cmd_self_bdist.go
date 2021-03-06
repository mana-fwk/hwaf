package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/gonuts/commander"
	"github.com/gonuts/flag"
	"github.com/hwaf/gas"
)

func hwaf_make_cmd_self_bdist() *commander.Command {
	cmd := &commander.Command{
		Run:       hwaf_run_cmd_self_bdist,
		UsageLine: "bdist [options]",
		Short:     "create a binary distribution of hwaf itself",
		Long: `
self bdist creates a binary distribution of hwaf itself.

ex:
 $ hwaf self bdist
 $ hwaf self bdist -version=20130101
`,
		Flag: *flag.NewFlagSet("hwaf-self-bdist", flag.ExitOnError),
	}
	cmd.Flag.Bool("v", false, "enable verbose output")
	cmd.Flag.String("version", "", "version of the binary distribution (default: 'time now')")

	return cmd
}

func hwaf_run_cmd_self_bdist(cmd *commander.Command, args []string) error {
	var err error

	n := "hwaf-self-" + cmd.Name()

	switch len(args) {
	case 0:
		// ok
	default:
		return fmt.Errorf("%s: does NOT take any argument", n)
	}

	verbose := cmd.Flag.Lookup("v").Value.Get().(bool)

	bdist_name := "hwaf"
	bdist_vers := cmd.Flag.Lookup("version").Value.Get().(string)
	bdist_variant := fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)

	if bdist_vers == "" {
		bdist_vers = time.Now().Format("20060102")
	}

	dirname := fmt.Sprintf("%s-%s-%s", bdist_name, bdist_vers, bdist_variant)
	fname := dirname + ".tar.gz"

	if verbose {
		fmt.Printf("%s [%s]...\n", n, fname)
	}

	tmpdir, err := ioutil.TempDir("", "hwaf-self-bdist-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpdir)
	//fmt.Printf(">>> [%s]\n", tmpdir)

	//
	top := filepath.Join(tmpdir, dirname)
	// create hierarchy of dirs for bdist
	for _, dir := range []string{
		"bin",
		"share",
		filepath.Join("share", "hwaf"),
	} {
		err = os.MkdirAll(filepath.Join(top, dir), 0755)
		if err != nil {
			return err
		}
	}

	// add hep-waftools cache
	hwaf_dir, err := gas.Abs("github.com/hwaf/hwaf")
	if err != nil {
		return err
	}

	src_hwaf_tools := filepath.Join(hwaf_dir, "py-hwaftools")
	hwaf_tools := filepath.Join(top, "share", "hwaf", "tools")

	err = copytree(hwaf_tools, src_hwaf_tools)
	if err != nil {
		return err
	}

	// remove git stuff
	err = os.RemoveAll(filepath.Join(hwaf_tools, ".git"))
	if err != nil {
		return err
	}

	// add share/hwaf/hwaf.conf
	err = ioutil.WriteFile(
		filepath.Join(top, "share", "hwaf", "hwaf.conf"),
		[]byte(`# hwaf config file
[hwaf]

## EOF ##
`),
		0644,
	)
	if err != nil {
		return err
	}

	// temporary GOPATH - install go-deps
	gopath := filepath.Join(tmpdir, "gocode")
	err = os.MkdirAll(gopath, 0755)
	if err != nil {
		return err
	}

	orig_gopath := os.Getenv("GOPATH")
	err = os.Setenv("GOPATH", gopath)
	if err != nil {
		return err
	}
	defer os.Setenv("GOPATH", orig_gopath)

	for _, gopkg := range []string{
		"github.com/hwaf/hwaf",
		"github.com/hwaf/hwaf-cmt2yml",
		"github.com/hwaf/hwaf-gen-extpackdist",
		"github.com/hwaf/hwaf-gen-lcgcmt",
		"github.com/hwaf/hwaf-rcore2yml",
	} {
		goget := exec.Command("go", "get", "-v", gopkg)
		goget.Dir = gopath
		if verbose {
			goget.Stdout = os.Stdout
			goget.Stderr = os.Stderr
		}
		err = goget.Run()
		if err != nil {
			return err
		}

		// install under /bin
		dst_fname := filepath.Join(top, "bin", filepath.Base(gopkg))
		dst, err := os.OpenFile(dst_fname, os.O_WRONLY|os.O_CREATE, 0755)
		if err != nil {
			return err
		}
		defer func(dst *os.File) error {
			err := dst.Sync()
			if err != nil {
				return err
			}
			err = dst.Close()
			return err
		}(dst)

		src_fname := filepath.Join(gopath, "bin", filepath.Base(gopkg))
		if !path_exists(src_fname) {
			// maybe a cross-compilation ?
			src_fname = filepath.Join(gopath, "bin", runtime.GOOS+"_"+runtime.GOARCH, filepath.Base(gopkg))
		}
		src, err := os.Open(src_fname)
		if err != nil {
			return err
		}
		defer func(src *os.File) error {
			return src.Close()
		}(src)

		_, err = io.Copy(dst, src)
		if err != nil {
			return err
		}
	}

	// add waf-bin
	waf_fname := filepath.Join(top, "bin", "waf")
	if path_exists(waf_fname) {
		err = os.Remove(waf_fname)
		if err != nil {
			return err
		}
	}
	waf_dst, err := os.OpenFile(waf_fname, os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		return err
	}
	defer func() error {
		err = waf_dst.Sync()
		if err != nil {
			return err
		}
		err = waf_dst.Close()
		if err != nil {
			return err
		}
		return err
	}()

	waf_src, err := os.Open(filepath.Join(
		gopath, "src", "github.com", "hwaf", "hwaf", "waf"),
	)
	if err != nil {
		return err
	}
	defer waf_src.Close()
	_, err = io.Copy(waf_dst, waf_src)
	if err != nil {
		return err
	}

	const bq = "`"
	// add setup.sh
	setup_fname, err := os.Create(filepath.Join(top, "setup-hwaf.sh"))
	if err != nil {
		return err
	}
	defer setup_fname.Close()
	_, err = fmt.Fprintf(setup_fname, `#!/bin/sh 

if [ "x${BASH_ARGV[0]}" = "x" ]; then
    ## assume zsh
    SOURCE="$0"
else
    SOURCE="${BASH_SOURCE[0]}"
fi

DIR="$( cd -P "$( dirname "$SOURCE" )" && pwd )"
echo ":: adding [$DIR/bin] to PATH"
export PATH=$DIR/bin:$PATH
## EOF
`)
	if err != nil {
		return err
	}

	err = setup_fname.Sync()
	if err != nil {
		return err
	}

	// add setup.csh
	csetup_fname, err := os.Create(filepath.Join(top, "setup-hwaf.csh"))
	if err != nil {
		return err
	}
	defer csetup_fname.Close()
	_, err = fmt.Fprintf(csetup_fname, `#!/bin/csh
# Absolute path to this script
set SCRIPT=%sreadlink -f "$0"%s
# Absolute path this script is in
set SCRIPTPATH=%sdirname "$SCRIPT"%s
echo ":: adding [$SCRIPTPATH/bin] to PATH"
setenv PATH $SCRIPTPATH/bin:$PATH

## EOF
`, bq, bq, bq, bq)
	if err != nil {
		return err
	}
	err = csetup_fname.Sync()
	if err != nil {
		return err
	}

	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	// package everything up
	err = _tar_gz(filepath.Join(pwd, fname), top)
	if err != nil {
		return err
	}

	if verbose {
		fmt.Printf("%s [%s]... [ok]\n", n, fname)
	}

	return err
}

// EOF
