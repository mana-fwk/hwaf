package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gonuts/commander"
	"github.com/gonuts/flag"
	//gocfg "github.com/gonuts/config"
	_ "github.com/hwaf/git-tools/utils"
)

func hwaf_make_cmd_pkg_rm() *commander.Command {
	cmd := &commander.Command{
		Run:       hwaf_run_cmd_pkg_rm,
		UsageLine: "rm [options] <local-pkg-name> [<pkg2> [...]]",
		Short:     "remove a package from the current workarea",
		Long: `
rm removes a package from the current workarea.

ex:
 $ hwaf pkg rm ./src/foo/pkg
 $ hwaf pkg rm Control/AthenaKernel
 $ hwaf pkg rm Control/AthenaKernel Control/AthenaServices
 $ hwaf pkg rm -f Control/AthenaKernel
`,
		Flag: *flag.NewFlagSet("hwaf-pkg-rm", flag.ExitOnError),
	}
	cmd.Flag.Bool("q", true, "only print error and warning messages, all other output will be suppressed")
	cmd.Flag.Bool("f", false, "force removing the package (from disk and from internal db)")

	return cmd
}

func hwaf_run_cmd_pkg_rm(cmd *commander.Command, args []string) {
	var err error
	n := "hwaf-pkg-" + cmd.Name()
	switch len(args) {
	case 0:
		err = fmt.Errorf("%s: you need to give (at least) one package name to remove", n)
		handle_err(err)
	}

	quiet := cmd.Flag.Lookup("q").Value.Get().(bool)
	force := cmd.Flag.Lookup("f").Value.Get().(bool)

	cfg, err := g_ctx.LocalCfg()
	handle_err(err)

	srcdir := "src"
	if cfg.HasOption("hwaf-cfg", "cmtpkgs") {
		srcdir, err = cfg.String("hwaf-cfg", "cmtpkgs")
		handle_err(err)
	}

	do_rm := func(pkgname string) error {
		var err error

		pkgname = os.ExpandEnv(pkgname)
		pkgname = filepath.Clean(pkgname)
		if !quiet {
			fmt.Printf("%s: remove package [%s]...\n", n, pkgname)
		}

		pkg := pkgname
		if !g_ctx.PkgDb.HasPkg(pkg) {
			pkg = filepath.Join(srcdir, pkgname)
		}
		if !g_ctx.PkgDb.HasPkg(pkg) {
			return fmt.Errorf("%s: no such package [%s] in db", n, pkg)
		}

		// people may have already removed it via a simple 'rm -rf foo'...
		if !path_exists(pkg) && !force {
			err = fmt.Errorf(
				`%s: no such package [%s]
%s: did you remove it by hand ? (re-try with 'hwaf pkg rm -f %s')`,
				n, pkgname,
				n, pkgname,
			)
			return err
		}

		vcspkg, err := g_ctx.PkgDb.GetPkg(pkg)
		if err != nil {
			return err
		}

		switch vcspkg.Type {

		case "svn", "git":
			if path_exists(vcspkg.Path) {
				err = os.RemoveAll(vcspkg.Path)
				if err != nil {
					return err
				}
				path := vcspkg.Path
				// clean-up dir if empty...
				for {
					path = filepath.Dir(path)
					if path == srcdir {
						break
					}
					content, err := filepath.Glob(filepath.Join(path, "*"))
					if err != nil {
						return err
					}
					if len(content) > 0 {
						break
					}
					err = os.RemoveAll(path)
					if err != nil {
						return err
					}
				}
			}

		case "local":
			if path_exists(vcspkg.Path) {
				err = os.RemoveAll(vcspkg.Path)
				if err != nil {
					return err
				}
				path := vcspkg.Path
				// clean-up dir if empty...
				for {
					path = filepath.Dir(path)
					if path == srcdir {
						break
					}
					content, err := filepath.Glob(filepath.Join(path, "*"))
					if err != nil {
						return err
					}
					if len(content) > 0 {
						break
					}
					err = os.RemoveAll(path)
					if err != nil {
						return err
					}
				}
			}

		default:
			return fmt.Errorf("%s: VCS of type [%s] is not handled", n, vcspkg.Type)
		}

		err = g_ctx.PkgDb.Remove(pkg)
		if err != nil {
			if !force {
				return err
			}
		}

		if !quiet {
			fmt.Printf("%s: remove package [%s]... [ok]\n", n, pkgname)
		}
		return nil
	}

	errs := []error{}
	for _, pkgname := range args {
		err := do_rm(pkgname)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		for _, err := range errs {
			fmt.Printf("%s\n", err.Error())
		}
		handle_err(errs[0])
	}
}

// EOF
