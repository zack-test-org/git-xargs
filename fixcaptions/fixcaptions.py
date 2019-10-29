import click
import os
import srt
import shutil
import glob


@click.command()
@click.argument('srtfile')
def fixcaptions(srtfile):
    if os.path.isdir(srtfile):
        srt_files = glob.glob(os.path.join(srtfile, '*.srt'))
    else:
        srt_files = [srtfile]

    for srtfile in srt_files:
        print('Fixing captions in {}'.format(srtfile))
        _fix_caption_file(srtfile)


def _fix_caption_file(srtfile):
    # Copy to a backup
    shutil.copyfile(srtfile, srtfile + '.bak')

    with open(srtfile, 'r') as f:
        parsed = list(srt.parse(f.read()))

    new_srt = []
    for i in range(0, len(parsed), 2):
        if i + 1 < len(parsed):
            parsed[i].content += ' ' + parsed[i + 1].content
        new_srt.append(parsed[i])

    for i, sub in enumerate(new_srt):
        sub.index = i

    with open(srtfile, 'w') as f:
        f.write(srt.compose(new_srt))


if __name__ == '__main__':
    fixcaptions()
