#!/bin/bash

# This script checks for the existence of exiftool, ffmpeg, and ffprobe.
# If any library is missing, it downloads and installs the program locally
# inside the project.

echo "Running install script from '$PWD'"

EXIFTOOL_URL="https://exiftool.org/Image-ExifTool-12.92.tar.gz"
EXIFTOOL_ARCHIVE="$CACHE_DIR/exiftool.tar.gz"

FFMPEG_URL="https://evermeet.cx/ffmpeg/ffmpeg-7.0.2.zip"
FFMPEG_ARCHIVE="$CACHE_DIR/ffmpeg.zip"

FFPROBE_URL="https://evermeet.cx/ffmpeg/ffprobe-7.0.2.zip"
FFPROBE_ARCHIVE="$CACHE_DIR/ffprobe.zip"

if [ ! -f $EXIFTOOL_BIN ]; then
  mkdir -p $EXIFTOOL_DIR
  curl --output $EXIFTOOL_ARCHIVE $EXIFTOOL_URL
  tar -xvzf $EXIFTOOL_ARCHIVE -C $EXIFTOOL_DIR --strip-components=1
  rm -rf $EXIFTOOL_ARCHIVE
else
  echo "exiftool already exists at $EXIFTOOL_BIN"
fi

if [ ! -f $FFMPEG_BIN ]; then
  mkdir -p $FFMPEG_DIR
  curl --output $FFMPEG_ARCHIVE $FFMPEG_URL
  unzip $FFMPEG_ARCHIVE -d $FFMPEG_DIR
  rm -rf $FFMPEG_ARCHIVE
  $FFMPEG_BIN -version
else
  echo "ffmpeg already exists at $FFMPEG_BIN"
fi

if [ ! -f $FFPROBE_BIN ]; then
  mkdir -p $FFPROBE_DIR
  curl --output $FFPROBE_ARCHIVE $FFPROBE_URL
  unzip $FFPROBE_ARCHIVE -d $FFPROBE_DIR
  rm -rf $FFPROBE_ARCHIVE
else
  echo "ffprobe already exists at $FFPROBE_BIN"
fi
