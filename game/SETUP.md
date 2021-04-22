### Installing packages

- `sudo apt-get install build-essential unzip xserver-xorg-core x11-utils kubuntu-desktop pkg-config libglvnd-dev yasm cmake libtool libc6 libc6-dev wget libnuma1 libnuma-dev libpulse-dev libopus-dev docker.io`

### User permission for uinput

- `echo KERNEL==\"uinput\", GROUP=\"$USER\", MODE:=\"0660\" | sudo tee /etc/udev/rules.d/99-$USER.rules` Linux uinput
- `sudo udevadm trigger`

### Compiling ffmpeg

- `git clone https://git.videolan.org/git/ffmpeg/nv-codec-headers.git`
- `cd nv-codec-headers && sudo make install && cd -`
- `git clone https://git.ffmpeg.org/ffmpeg.git ffmpeg/`
- `cd ffmpeg`
- `./configure --enable-cuda --enable-cuvid --enable-nvdec --enable-nvenc --enable-nonfree --enable-libopus --enable-libpulse --enable-opengl --extra-cflags=-I/usr/local/cuda/include  --extra-ldflags=-L/usr/local/cuda/lib64`
- `make -j 2`
- `sudo make install`

### Configuration X server

- `sudo nvidia-xconfig --mode-list=1280x720 --separate-x-screens`
