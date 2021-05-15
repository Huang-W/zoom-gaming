ffplay:
	ffplay -protocol_whitelist file,udp,rtp sdp

video-run:
	ffmpeg -re -f lavfi -i "testsrc=size=640x480:rate=30" -vcodec libvpx -cpu-used 5 -deadline 1 -g 10 -error-resilient 1 -auto-alt-ref 1 -f rtp rtp://127.0.0.1:5004

audio-run:
	ffmpeg -f lavfi -i "sine=frequency=1000" -c:a libopus -b:a 48000 -sample_fmt s16p -ssrc 1 -payload_type 111 -f rtp -max_delay 0 -application lowdelay rtp:/127.0.0.1:4004

info:
	xdpyinfo | more

go-start:
	go run main.go > game.log 2>&1 &

h264_nvenc_help:
	ffmpeg -h encoder=h264_nvenc

gcp-start:
	gcloud beta compute --project=zoomgaming-311019 instances create nvidiap4 --zone=us-central1-a --machine-type=n1-standard-1 --subnet=public --network-tier=PREMIUM --maintenance-policy=TERMINATE --service-account=256980823626-compute@developer.gserviceaccount.com --scopes=https://www.googleapis.com/auth/devstorage.read_only,https://www.googleapis.com/auth/logging.write,https://www.googleapis.com/auth/monitoring.write,https://www.googleapis.com/auth/servicecontrol,https://www.googleapis.com/auth/service.management.readonly,https://www.googleapis.com/auth/trace.append --accelerator=type=nvidia-tesla-p4,count=1 --tags=game-server --image=ubuntu-2004-focal-v20210415 --image-project=ubuntu-os-cloud --boot-disk-size=100GB --boot-disk-type=pd-balanced --boot-disk-device-name=nvidiap4 --no-shielded-secure-boot --shielded-vtpm --shielded-integrity-monitoring --reservation-affinity=any

instances-list:
	gcloud compute instances list

ssh-instance:
	gcloud beta compute ssh --zone "us-west2-c" "nvidiap4-2" --project "zoomgaming-311019"

scp-to-instance:
	gcloud beta compute scp --zone "us-west2-c" --project "zoomgaming-311019" $(file) $(user)@$(name):~/

scp-from-instance:
	gcloud compute scp $(user)@$(name):~/ $(file)

get-credentials:
	gcloud container clusters get-credentials --zone us-central1-$(z) $(cluster)

ssh-tunnel:
	gcloud compute ssh $(name) -- -L $(l):localhost:$(r)

docker-credentials:
	sudo usermod -a -G docker ${USER}
	docker-credential-gcr configure-docker --registries=us-central1-docker.pkg.dev

list-pulse-src:
	pacmd list-sink-inputs
