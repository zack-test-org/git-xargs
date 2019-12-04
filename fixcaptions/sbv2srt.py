import click
import os
import srt
import glob
import datetime
import re
from collections import deque

_TIMEDELTA_RE = r'(\d+):(\d+):(\d+).(\d+)'


@click.command()
@click.argument('sbvfile')
def sbv2srt(sbvfile):
    if os.path.isdir(sbvfile):
        sbv_files = glob.glob(os.path.join(sbvfile, '*.sbv'))
    else:
        sbv_files = [sbvfile]

    for sbvfile in sbv_files:
        convert(sbvfile)


def convert(sbvfile):
    target_srt_base, _ = os.path.splitext(sbvfile)
    target_srt_file = target_srt_base + '.srt'

    print('Converting sbv file "{}" to srt file "{}"'.format(
        sbvfile, target_srt_file))

    with open(sbvfile, 'r') as f:
        contents = f.read()

    new_srt = []

    lines = deque(contents.splitlines())
    # Keep inspecting each line until we parse all contents of the file
    while lines:
        cur = lines.popleft()
        if _is_start_of_caption(cur):
            # We are at the start of a caption, so parse it into a Subtitle object.

            # First parse the timestamps
            raw_start_time, raw_end_time = cur.split(',')
            start_time = _parse_timedelta(raw_start_time)
            end_time = _parse_timedelta(raw_end_time)

            # Now parse the text. All text until a blank line are considered a part of this block, so we keep
            # popleft-ing until we reach an empty line.
            text = lines.popleft()
            while lines and lines[0]:
                text += ' ' + lines.popleft()

            # We now have everything we need to create the Subtitle object
            new_srt.append(
                srt.Subtitle(len(new_srt) + 1, start_time, end_time, text))

    with open(target_srt_file, 'w') as f:
        f.write(srt.compose(new_srt))


def _parse_timedelta(timedelta_str):
    """
    _parse_timedelta takes a line that looks like HH:MM:SS.milliseconds and converts it to a datetime.timedelta object.
    """
    regex = re.compile(_TIMEDELTA_RE)
    parts = regex.match(timedelta_str)
    return datetime.timedelta(
        hours=int(parts.group(1)),
        minutes=int(parts.group(2)),
        seconds=int(parts.group(3)),
        milliseconds=int(parts.group(4)),
    )


def _is_start_of_caption(line):
    """
    _is_start_of_caption returns true if this line marks the start of a caption section, which is a timestamp line.
    """
    regex = re.compile(_TIMEDELTA_RE + r',' + _TIMEDELTA_RE)
    return regex.match(line) is not None


if __name__ == '__main__':
    sbv2srt()
