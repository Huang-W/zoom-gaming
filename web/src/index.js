const pb = require('./proto/signaling_pb');
const echo = require('./proto/echo_pb');
import adapter from "node_modules/webrtc-adapter";
import { DCLabel } from "./datachannel";

(function() {
  let echo_dc = null;
  let webSocket = new WebSocket("ws://localhost:8080/demo");
  webSocket.binaryType = "arraybuffer" // blob or arraybuffer

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
    let signaling_event = new pb.SignalingEvent();
    signaling_event.setSessionDescription(pb_sd);
    let uint8_array = signaling_event.serializeBinary();
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

  let startRemoteSession = (remoteVideoNode, remoteAudioNode, stream) => {
    let pc;

    return Promise.resolve().then(() => {
      pc = new RTCPeerConnection({
        iceServers: [{ urls: 'stun:stun.l.google.com:19302' }]
      });

      pc.addTransceiver('audio', {'direction': 'recvonly'})
      pc.addTransceiver('video', {'direction': 'recvonly'})

      pc.ontrack = (event) => {
        console.info('ontrack triggered');
        console.log(event);
        if (event.track.kind == "video") {
          remoteVideoNode.srcObject = event.streams[0]
          remoteVideoNode.play()
        }
        if (event.track.kind == "audio") {
          remoteAudioNode.srcObject = event.streams[0]
          remoteAudioNode.play()
        }
      };

      echo_dc = pc.createDataChannel(DCLabel.String(DCLabel.Label.ECHO), {
        negotiated: true,
        id: DCLabel.Id(DCLabel.Label.ECHO)
      });

      echo_dc.onopen = () => { console.log(`Data Channel ${echo_dc.label} - ${echo_dc.id} is open`); }
      echo_dc.onclose = () => { console.log(`Data Channel ${echo_dc.label} - ${echo_dc.id} is closed`); }
      echo_dc.onerror = (event) => { console.log(event); }
      echo_dc.onmessage = (event) => {
        let pb_echo = DCLabel.Label.PbMessageType(DCLabel.Label.ECHO);
        let msg = new pb_echo.deserializeBinary(event.data);
        console.log(`New data channel message on ${label} has arrived: `);
        console.log(msg);
      };

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
    const remoteVideo = document.querySelector('#remote-video');
    const remoteAudio = document.querySelector('#remote-audio');
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

      const userMediaPromise =  (adapter.browserDetails.browser === 'safari') ?
        navigator.mediaDevices.getUserMedia({ video: true }) :
        Promise.resolve(null);
      if (!peerConnection) {
        userMediaPromise.then(stream => {
          return startRemoteSession(remoteVideo, remoteAudio, stream).then(pc => {
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
      let signaling_event = new pb.SignalingEvent.deserializeBinary(event.data);
      switch (signaling_event.getEventCase()) {
        case pb.SignalingEvent.EventCase.SESSION_DESCRIPTION:
          console.log("sdp!");
          let pb_sdp = signaling_event.getSessionDescription();
          handleAnswer(pb_sdp);
          break;
        case pb.SignalingEvent.EventCase.EVENT_NOT_SET:
          console.log("SignalingEvent's event field is empty");
          break;
        default:
          console.log("Unable to deserialize into pb.SignalingEvent");
          break;
      }
    } else if ( Object.getPrototypeOf(event) === CloseEvent.prototype ) {
      console.log("ws closing");
    } else {
      console.log(`Received event with prototype of ${Object.getPrototypeOf(event)}`);
    }
  }
})()
