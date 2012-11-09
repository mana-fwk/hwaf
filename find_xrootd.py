# -*- python -*-

# stdlib imports ---
import os
import os.path as osp

# waf imports ---
import waflib.Utils
import waflib.Logs as msg
from waflib.Configure import conf

#
_heptooldir = osp.dirname(osp.abspath(__file__))

def options(opt):

    opt.load('findbase', tooldir=_heptooldir)

    opt.add_option(
        '--with-xrootd',
        default=None,
        help="Look for xrootd at the given path")
    return

def configure(conf):
    conf.load('findbase platforms', tooldir=_heptooldir)
    return

@conf
def find_xrootd(ctx, **kwargs):
    
    if not ctx.env.CXX:
        msg.fatal('load a C++ compiler first')
        pass


    kwargs = ctx._findbase_setup(kwargs)
    
    kwargs['mandatory'] = kwargs.get('mandatory', False)
    ctx.check_with(
        ctx.check,
        "xrootd",
        features='cxx cxxprogram',
        header_name="xrootd/XrdVersion.hh",
        uselib_store='xrootd',
        **kwargs
        )

    bindir = osp.join(ctx.env.XROOTD_HOME, 'bin')
    libdir = osp.join(ctx.env.XROOTD_HOME, 'lib')
    incdir = osp.join(ctx.env.XROOTD_HOME, 'include')

    ctx.define_uselib(
        name="xrootd-posix", 
        libpath=libdir,
        libname="XrdPosix",
        incpath=incdir, 
        incname="xrootd/XrdPosix/XrdPosix.hh",
        )

    ctx.define_uselib(
        name="xrootd-client",
        libpath=libdir,
        libname="XrdClient",
        incpath=incdir, 
        incname="xrootd/XrdClient/XrdClient.hh",
        )

    ctx.define_uselib(
        name="xrootd-utils",
        libpath=libdir,
        libname="XrdUtils",
        incpath=incdir, 
        incname="xrootd/Xrd/XrdConfig.hh",
        )

    ctx.find_program(
        "xrdcp", 
        var="XRDCP-BIN", 
        path_list=bindir,
        **kwargs)

    ctx.find_program(
        "xrootd", 
        var="XROOTD-BIN", 
        path_list=bindir,
        **kwargs)

    # -- check everything is kosher...
    ctx.check_cxx(
        msg="Checking compilation xrootd",
        fragment='''\
        #include "xrootd/XrdVersion.hh"

        int main(int argc, char* argv[]) {
          return 0;
        }
        ''',
        use="xrootd",
        execute  = True,
        )

    ctx.env.HEPWAF_FOUND_XROOTD = 1
    return

## EOF ##
