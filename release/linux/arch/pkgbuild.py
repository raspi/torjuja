import io
import logging

log = logging.getLogger(__name__)

import os
import sys
import pathlib
import argparse

__VERSION__ = "0.0.1"
__AUTHOR__ = u"Pekka Järvinen"
__YEAR__ = 2021
__DESCRIPTION__ = u"Generate PKGBUILD. Version {0}.".format(__VERSION__)
__EPILOG__ = u"%(prog)s v{0} (c) {1} {2}-".format(__VERSION__, __AUTHOR__, __YEAR__)

__EXAMPLES__ = [
    u'',
    u'-' * 60,
    u'%(prog)s --checksums filelist.shasums',
    u'-' * 60,
]

if __name__ == '__main__':
    logging.basicConfig(
        format='%(asctime)s [%(levelname)s]: %(message)s',
        stream=sys.stdout,
        level=logging.INFO,
    )

    parser = argparse.ArgumentParser(
        description=__DESCRIPTION__,
        epilog=__EPILOG__,
        usage=os.linesep.join(__EXAMPLES__),
    )

    parser.add_argument('--verbose', '-v', action='count', required=False, default=0, dest='verbose',
                        help="Be verbose. -vvv.. Be more verbose.")
    parser.add_argument('--version', '-V', action='store', required=True, dest='version', type=str, help='')

    args = parser.parse_args()

    if int(args.verbose) > 0:
        logging.getLogger().setLevel(logging.DEBUG)
        log.info("Being verbose")

    checksumsfile = list(pathlib.Path(
        os.path.join('..', '..', 'v' + args.version)
    ).glob('*.shasums'))[0]

    files:dict = {}

    with open(checksumsfile, encoding='utf8') as f:
        for i in f:
            csum, fname = i.strip().split('  ')
            fname = os.path.basename(fname)
            if fname.find('-linux-') == -1:
                # skip
                continue

            _, _, _, arch = fname.split('-')
            archP = pathlib.Path(arch)
            extensions = "".join(archP.suffixes)
            arch = str(archP).removesuffix(extensions)

            # Replace Go architecture names with Arch ones
            if arch == 'arm64':
                arch = 'aarch64'
            elif arch == 'amd64':
                arch = 'x86_64'

            files[fname] = {
                'sum': csum,
                'arch': arch,
            }

    checksumtpl:str = ""

    with io.StringIO() as f:
        archlist = []
        for k, v in files.items():
            archlist.append(v['arch'])

        f.write('arch=(' + ' '.join(map(lambda x: f"'{x}'", archlist)) + ')' + "\n")

        for k, v in files.items():
            f.write(f"sha256sums_{v['arch']}=('{v['sum']}')" + "\n")
            f.write(f'source_{v["arch"]}=("https://github.com/raspi/torjuja/releases/download/$pkgver/{k}")' + "\n")

        checksumtpl = f.getvalue()

    tpl:str = ""

    with open('template.txt', 'r', encoding='utf8') as f:
        tpl = f.read()

    tpl = tpl.replace('%APPNAME%', 'torjuja')
    tpl = tpl.replace('%VERSION%', 'v' + args.version)
    tpl = tpl.replace('%CHECKSUM%', checksumtpl)

    with open('PKGBUILD', 'w', encoding='utf8') as f:
        f.write(tpl)
