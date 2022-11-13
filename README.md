# remotepi

This is a PoC of controlling my Marantz PM6007 stereo receiver over the Remote Control port.  
It can also be changed to work with an IR LED, though this was only tested initially.

It expect it will work for most Marantz receivers.

It is actively used on my Raspberry Pi with an IQAudio sound card HAT.  

It simply sends an ON signal to the receiver, when the ALSA device starts streaming.  
When the stream stops, an OFF signal is sent to the receiver.  

It just loops forever (with sleeps of course), while reading `/proc/asound/card0/pcm*/sub*/status`.  
State is on, when `state: RUNNING` in found in `status`.

The rc5 package can create and send signals for both wired and IR signals.
