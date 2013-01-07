package main

import (
	"fmt"
	"os"
	// "os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/sbinet/go-commander"
	"github.com/sbinet/go-flag"
)

func hwaf_make_cmd_waf_bdist() *commander.Command {
	cmd := &commander.Command{
		Run:       hwaf_run_cmd_waf_bdist,
		UsageLine: "bdist [output-filename]",
		Short:     "create a binary distribution from the project or packages",
		Long: `
bdist creates a binary distribution from the project or packages.

ex:
 $ hwaf bdist
 $ hwaf bdist -name=mana
 $ hwaf bdist -name=mana -version=20121218
 $ hwaf bdist -name=mana -version -cmtcfg=x86_64-linux-gcc-opt
 $ hwaf bdist mana-20121218-x86_64-linux-gcc-opt
`,
		Flag: *flag.NewFlagSet("hwaf-bdist", flag.ExitOnError),
		//CustomFlags: true,
	}
	cmd.Flag.Bool("q", false, "only print error and warning messages, all other output will be suppressed")
	cmd.Flag.String("name", "", "name of the binary distribution (default: project name)")
	cmd.Flag.String("version", "", "version of the binary distribution (default: project version)")
	cmd.Flag.String("cmtcfg", "", "CMTCFG quadruplet for the binary distribution (default: project CMTCFG)")
	return cmd
}

func hwaf_run_cmd_waf_bdist(cmd *commander.Command, args []string) {
	var err error
	n := "hwaf-" + cmd.Name()

	fname := ""
	switch len(args) {
	case 0:
		fname = ""
	case 1:
		fname = args[0]
	default:
		err = fmt.Errorf("%s: too many arguments (%d)", n, len(args))
		handle_err(err)
	}

	bdist_name := cmd.Flag.Lookup("name").Value.Get().(string)
	bdist_vers := cmd.Flag.Lookup("version").Value.Get().(string)
	bdist_cmtcfg := cmd.Flag.Lookup("cmtcfg").Value.Get().(string)

	workdir, err := get_workarea_root()
	handle_err(err)

	// FIXME: get actual value from waf, somehow
	pinfo_name := filepath.Join(workdir, "__build__", "project.info")
	if !path_exists(pinfo_name) {
		err = fmt.Errorf(
			"no such file [%s]. did you run \"hwaf configure\" ?", 
			pinfo_name,
			)
		handle_err(err)
	}
	pinfo,err := NewProjectInfo(pinfo_name)
	handle_err(err)
	if fname == "" {
		if bdist_name == "" {
			bdist_name = workdir
			bdist_name = filepath.Base(bdist_name)
		}
		if bdist_vers == "" {
			bdist_vers = time.Now().Format("20060102")
		}
		if bdist_cmtcfg == "" {
			bdist_cmtcfg, err = pinfo.Get("CMTCFG")
			handle_err(err)
		}
		fname = bdist_name + "-" + bdist_vers + "-" + bdist_cmtcfg
	}
	if !strings.HasSuffix(fname, ".tar.gz") {
		fname = fname + ".tar.gz"
	}

	//fmt.Printf(">> fname=[%s]\n", fname)
	fname = filepath.Join(workdir, fname)

	// FIXME: get actual value from waf, somehow
	install_area := filepath.Join(workdir, "install-area")
	if !path_exists(install_area) {
		err = fmt.Errorf(
			"no such directory [%s]. did you run \"hwaf install\" ?",
			install_area,
			)
		handle_err(err)
	}
	err = os.Chdir(install_area)
	handle_err(err)

	err = _tar_gz(fname, ".")
	handle_err(err)
}

// EOF