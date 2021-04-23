const pb = require('./proto/signaling_pb');
const input = require('./proto/input_pb');
import adapter from "node_modules/webrtc-adapter";
import { DCLabel } from "./datachannel";
import { InputMap } from "./input";

// const SERVER_ADDR = "34.94.73.231";
const SERVER_ADDR = "127.0.0.1";

(function() {
  let input_dc = null;
  let webSocket = new WebSocket(`ws://${SERVER_ADDR}:8080/demo`);
  webSocket.binaryType = "arraybuffer" // blob or arraybuffer

  webSocket.addEventListener("open", event => {
    console.log("ws open");
  })
  webSocket.addEventListener("message", event => {
    handleWebsocketEvent(event);
  });
  webSocket.addEventListener("close", event => {
    console.log("ws closing");
  });
  webSocket.onerror = function(event) {
    console.error("WebSocket error observed:", event);
  };

  let peerConnection = null;

  let showError = (error) => {
    console.log(error);
    const errorNode = document.querySelector('#error');
    if (errorNode.firstChild) {
      errorNode.removeChild(errorNode.firstChild);
    }
    errorNode.appendChild(document.createTextNode(error.message || error));
  }

  let startSession = (offer) => {
    let pb_sd = new pb.SessionDescription();
    pb_sd.setType(pb.SessionDescription.SDPType.SDP_TYPE_OFFER);
    pb_sd.setSdp(offer);
    let uint8_array = pb_sd.serializeBinary();
    webSocket.send(uint8_array.buffer);
  }

  let createOffer = async (pc) => {
    return new Promise((accept, reject) => {
      pc.onicecandidate = evt => {
        if (!evt.candidate) {

          // ICE Gathering finished
          const { sdp: offer } = pc.localDescription;
          accept(offer);
        }
      };
      pc.createOffer().then(ld => {pc.setLocalDescription(ld)}).catch(reject)
    });
  }

  let startRemoteSession = (remoteVideoNode, playerVideoNodes, stream) => {
    let pc;

    return Promise.resolve().then(() => {
      pc = new RTCPeerConnection({
        iceServers: [{ urls: 'stun:stun.l.google.com:19302' }]
      });

      // game streaming
      pc.addTransceiver('audio', {'direction': 'recvonly'});
      pc.addTransceiver('video', {'direction': 'recvonly'});

      // camera + mic
      // pc.addTransceiver('audio', {'direction': 'sendonly'});
      // pc.addTransceiver('video', {'direction': 'sendonly'});

      pc.ontrack = ({ track, streams }) => {
        console.info('ontrack triggered');
        console.log(track);
        console.log(streams);
        if (!remoteVideoNode.srcObject) {
          remoteVideoNode.srcObject = streams[0];
          remoteVideoNode.play();
        }
      };

      input_dc = pc.createDataChannel(DCLabel.String(DCLabel.Label.GAME_INPUT), {
        negotiated: true,
        id: DCLabel.Id(DCLabel.Label.GAME_INPUT)
      });

      input_dc.onopen = () => { console.log(`Data Channel ${input_dc.label} - ${input_dc.id} is open`); }
      input_dc.onclose = () => { console.log(`Data Channel ${input_dc.label} - ${input_dc.id} is closed`); }
      input_dc.onerror = (event) => { console.log(event); }

      document.addEventListener('keydown', (event) => {
        if (!event.repeat && InputMap.has(event.code)) {
          // console.log(event.code);
          let input_msg = new input.InputEvent();
          let key_press_event = new input.KeyPressEvent();
          key_press_event.setDirection(input.KeyPressEvent.Direction.DIRECTION_DOWN);
          key_press_event.setKey(InputMap.get(event.code));
          input_msg.setKeyPressEvent(key_press_event);
          input_dc.send(input_msg.serializeBinary().buffer);
        }
      });

      document.addEventListener('keyup', (event) => {
        if (InputMap.has(event.code)) {
          // console.log(event.code);
          let input_msg = new input.InputEvent();
          let key_press_event = new input.KeyPressEvent();
          key_press_event.setDirection(input.KeyPressEvent.Direction.DIRECTION_UP);
          key_press_event.setKey(InputMap.get(event.code));
          input_msg.setKeyPressEvent(key_press_event);
          input_dc.send(input_msg.serializeBinary().buffer);
        }
      });

      stream && stream.getTracks().forEach(track => {
        pc.addTrack(track, stream);
      });
      return createOffer(pc);
    }).then(offer => {
      // console.info(offer);
      startSession(offer);
    }).then(() => pc);
  }

  document.addEventListener('DOMContentLoaded', () => {
    let selectedScreen = 0;
    const playerVideos = document.querySelector('#player-videos');
    const remoteVideo = document.querySelector('#remote-video');
    const screenSelect = document.querySelector('#screen-select');
    const startStop = document.querySelector('#start-stop');

    const option = document.createElement('option');
    option.appendChild(document.createTextNode('Screen 1'));
    option.setAttribute('value', 0);
    screenSelect.appendChild(option);

    screenSelect.addEventListener('change', evt => {
      selectedScreen = parseInt(evt.currentTarget.value, 10);
    });

    const enableStartStop = (enabled) => {
      if (enabled) {
        startStop.removeAttribute('disabled');
      } else {
        startStop.setAttribute('disabled', '');
      }
    }

    const setStartStopTitle = (title) => {
      startStop.removeChild(startStop.firstChild);
      startStop.appendChild(document.createTextNode(title));
    }

    startStop.addEventListener('click', () => {
      enableStartStop(false);

      const videoConstraints = {
          height: { ideal: 240, max: 300 },
          width: { ideal: 320, max: 400 }
      };
      const userMediaPromise = Promise.resolve(null);
      // const userMediaPromise = navigator.mediaDevices.getUserMedia({ video: videoConstraints, audio: true }) || Promise.resolve(null);
      if (!peerConnection) {
        userMediaPromise.then(stream => {
          let el = document.createElement("video");
          el.srcObject = stream;
          el.muted = true;
          el.autoplay = true;
          playerVideos.appendChild(el);
          return startRemoteSession(remoteVideo, playerVideos, stream).then(pc => {
            remoteVideo.style.setProperty('visibility', 'visible');
            peerConnection = pc;
          }).catch(showError).then(() => {
            enableStartStop(true);
            setStartStopTitle('Stop');
          });
        })
      } else {
        peerConnection.close();
        peerConnection = null;
        enableStartStop(true);
        setStartStopTitle('Start');
        remoteVideo.style.setProperty('visibility', 'collapse');
      }
    });
  });

  window.addEventListener('beforeunload', () => {
    if (peerConnection) {
      peerConnection.close();
    }
    if (webSocket) {
      webSocket.close();
    }
  });

  let handleAnswer = (sdp) => {
    if (sdp instanceof pb.SessionDescription) {
      // console.info(sdp.getSdp());
      peerConnection.setRemoteDescription({
        type: "answer",
        sdp: sdp.getSdp()
      });
    }
  }

  let handleWebsocketEvent = (event) => {
    if ( Object.getPrototypeOf(event) === MessageEvent.prototype ) {
      let remote_sdp_answer = new pb.SessionDescription.deserializeBinary(event.data);
      if (remote_sdp_answer) {
        console.log("sdp!");
        handleAnswer(remote_sdp_answer);
      } else {
        console.error("not an sdp");
      }
    } else if ( Object.getPrototypeOf(event) === CloseEvent.prototype ) {
      console.log("ws closing");
    } else {
      console.log(`Received event with prototype of ${Object.getPrototypeOf(event)}`);
    }
  }
})()
