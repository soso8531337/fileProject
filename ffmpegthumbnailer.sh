#!/bin/sh

ffmpegthumbnailer -i "$1" -o "$2" -c jpeg -s 256 
#ffmpegthumbnailer -i $1 -o $2 -c jpeg -s 256 
