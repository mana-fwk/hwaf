package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/gonuts/commander"
	"github.com/gonuts/flag"
)

func hwaf_make_cmd_waf_show_pkg_uses() *commander.Command {
	cmd := &commander.Command{
		Run:       hwaf_run_cmd_waf_show_pkg_uses,
		UsageLine: "pkg-uses",
		Short:     "show local project's dependencies",
		Long: `
show pkg-uses displays the uses of a given package.

ex:
 $ hwaf show pkg-uses Control/AthenaCommon
`,
		Flag:        *flag.NewFlagSet("hwaf-waf-show-pkg-uses", flag.ExitOnError),
		CustomFlags: true,
	}
	return cmd
}

func hwaf_run_cmd_waf_show_pkg_uses(cmd *commander.Command, args []string) {
	var err error
	//n := "hwaf-" + cmd.Name()

	top := hwaf_root()
	waf := filepath.Join(top, "bin", "waf")
	if !path_exists(waf) {
		err = fmt.Errorf(
			"no such file [%s]\nplease re-run 'hwaf self init'\n",
			waf,
		)
		handle_err(err)
	}

	subargs := append([]string{"show-pkg-uses"}, args...)
	sub := exec.Command(waf, subargs...)
	sub.Stdout = os.Stdout
	sub.Stderr = os.Stderr
	err = sub.Run()
	handle_err(err)
}

// EOF
