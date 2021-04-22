# ffmpeg -loglevel debug -threads 2 -filter_threads 2 \
ffmpeg -hwaccel cuda -hwaccel_output_format cuda -threads 2 -filter_threads 2 \
-f x11grab -draw_mouse 0 -s 1280x720 -framerate 60 -i :100.0 \
-b:v 2400k -minrate:v 2400k -maxrate:v 2400k -bufsize:v 2400k \
-c h264_nvenc -preset p4 -tune ll -profile high \
-f rtp rtp://127.0.0.1:5004
