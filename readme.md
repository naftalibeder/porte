# porte

A tool for fixing and organizing a Google Takeout photos export.

## Description

When Google Photos generates a backup of a photo library, the resulting photo files are frequently missing important exif data, like timestamps and geolocation. This metadata is placed into adjacent `.json` files.

porte uses each file's original metadata, any related metadata file, and the filename itself to output a directory of fixed files containing the most accurate information available.

<img width="100%" src="https://github.com/user-attachments/assets/1eda5914-236a-4902-9e24-6e1f5f613a97">

## Features

- Date handling
  - Copies timestamps, if needed, from any related metadata file.
  - If a date cannot be found in the exif data or a related metadata file, attempts to parse a date from the filename.
  - Uses the earliest date found, avoiding errors like assigning the file-modification date as the capture date.
  - Prefixes files with the capture date, for a chronologically ordered output directory.
- Geolocation handling
  - Copies geolocation data, if needed, from any related metadata file.
- File quality
  - Edits image exif data without recompressing files.
  - Edits video exif data and attempts to repackage as `.mp4` without re-encoding, for compatibility. Otherwise, simply renames and copies the file.
  - Fixes incorrect extensions based on the actual file data.
  - Preserves original filename in the output filename and exif title tag.
  - Copies files to an output directory, instead of modifying in-place.
- Understandable output
  - Sorts failed files into a separate folder to inspect manually.
  - Saves a comprehensive log of converting results for each file.
- Speed
  - Distributes work across available cores.

## Known limitations

- Does not preserve album data.
- Editing `.avi` exif data is not currently supported (due to it being unsupported in `ffmpeg`).
- Setting date tags on `.gif` files does not appear to work.
- Does not find metadata files that exist in a different archive than their corresponding image. (This requires more cleverness than just matching by filename or image title, since there are many duplicate filenames in a large photo library.)

## Etymology

_Porte_ comes from the French word _emporte_, meaning _takeout_, in the food sense.

## Installation

You can either use a prebuilt binary or build from source.

#### Prebuilt binary

Go to the [Releases](https://github.com/naftalibeder/porte/releases) page and download a binary for the most recent version.

To run the program on macOS, you need to make it executable:

```sh
chmod +x ./macos
```

#### From source

To build from source, first install these dependencies:

- [go](https://go.dev/doc/install)
- [exiftool](https://exiftool.org/install.html)
- [ffmpeg](https://ffmpeg.org/download.html)

Then, build the project:

```sh
go build -o ./porte
```

## Usage

To convert your photos, run:

```sh
porte srcpath destpath
```

where `srcpath` is any directory containing images or videos, and `destpath` is the desired output directory.

If `destpath` is omitted, a directory will be created by concatenating `srcpath` and `_Export`.

## Development

To run all tests:

```sh
go test ./...
```

## Acknowledgments

Thanks to these useful projects:

- Exif reader/writer: https://github.com/exiftool/exiftool
- Test images: https://github.com/ianare/exif-samples
