# Fix Captions

This fixes the `.srt` caption files from Youtube so that there are no overlaping time slices. Teachable requires that
each caption does not overlap, as it does not support incremental caption playback like Youtube does.

## Usage

### Formatting manually generated captions

Manual captions generated using the Youtube UI will generate `.sbv` files. This can be converted to `.srt` files using
the `sbv2srt.py` script:

```
python sbv2srt.py $PATH_TO_SBVFILE
```

### Formatting automatically generated captions

Automatically generated captions can be downloaded as `.srt` files using the Youtube UI. However, these are not in a
format that Teachable accepts since Youtube generates captions that overlap. You can use the `fixcaptions.py` script to
fix these files.

If you have a directory of Youtube `.srt` files, you can run the `fixcaptions.py` script like so:

```
python fixcaptions.py $DIRECTORY_WITH_SRT_FILES
```
