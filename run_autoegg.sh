#!/bin/bash
if [ ! -e /dev/hidg0 ]; then
	./procon_audio.sh
fi
sleep 6
env PA_ALSA_PLUGHW=1 ./autoegg -o autoegg.log
#sleep 120
#shutdown now
