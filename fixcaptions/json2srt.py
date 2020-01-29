import click
import os
import json
import srt
import glob
import datetime


@click.command()
@click.argument('jsonfile')
def json2srt(jsonfile):
    if os.path.isdir(jsonfile):
        json_files = glob.glob(os.path.join(jsonfile, '*.json'))
    else:
        json_files = [jsonfile]

    for jsonfile in json_files:
        convert(jsonfile)


def convert(jsonfile):
    target_srt_base, _ = os.path.splitext(jsonfile)
    target_srt_file = target_srt_base + '.srt'

    print('Converting json file "{}" to srt file "{}"'.format(jsonfile, target_srt_file))

    new_srt = []

    with open(jsonfile, 'r') as f:
        data = json.load(f)

    for title in data['titles']:
        start_time_milliseconds = title['s']
        start_time = datetime.timedelta(milliseconds=start_time_milliseconds)
        end_time_milliseconds = title['e']
        end_time = datetime.timedelta(milliseconds=end_time_milliseconds)
        text = title['t']
        new_srt.append(srt.Subtitle(len(new_srt) + 1, start_time, end_time, text))

    with open(target_srt_file, 'w') as f:
        f.write(srt.compose(new_srt))


if __name__ == '__main__':
    json2srt()
