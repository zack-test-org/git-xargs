# Fix Captions

This fixes the `.srt` caption files from Youtube so that there are no overlaping time slices. Teachable requires that
each caption does not overlap, as it does not support incremental caption playback like Youtube does.

## Usage

If you have a directory of `.srt` files, you can run the `fixcaptions.py` script like so:

```
python fixcaptions.py $DIRECTORY_WITH_SRT_FILES
```
