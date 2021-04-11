ffmpeg -f x11grab -draw_mouse 0 -s 1920x1080 -framerate 15 -i :0.0 \
  -c:v libx264 \
  -preset fast \
  -pix_fmt yuv420p \
  -s 1280x720 \
  -threads 0 \
  -f rtp \
  rtp://127.0.0.1:5004
