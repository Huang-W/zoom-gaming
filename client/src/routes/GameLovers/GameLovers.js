import React, {useEffect, useState} from "react";
import { DCLabel } from "./datachannel";
import { InputMap } from "./input";
import Button from "@material-ui/core/Button";
import {gamecode} from "../TabooGame/src/__fixtures__/game";
import {useHistory} from "react-router";
const pb = require('./proto/signaling_pb');
const input = require('./proto/input_pb');

// const SERVER_ADDR = "34.94.73.231";
const SERVER_ADDR = "w2.zoomgaming.app";

const GameLovers = (props) => {

  const history = useHistory();

  const [failed, updateFailed] = useState(true);

 useEffect(() => {
   let peerConnection = null; // webrtc connection
   let input_dc = null; // keyboard events are sent to the server using this
   let webSocket = new WebSocket(`wss://${SERVER_ADDR}/demo/${props.roomId}/${props.gameId}`); // session description is sent/received via websocket
   webSocket.binaryType = "arraybuffer" // blob or arraybuffer
   // webSocket.addEventListener("open", event => { console.log("ws open"); });
   webSocket.addEventListener("message", event => { handleWebsocketEvent(event); });
   webSocket.addEventListener("close", event => { console.log("ws closing") });
   webSocket.onerror = function(event) { console.error("WebSocket error observed:", event); };

   let startSession = (offer) => {
     let pb_sd = new pb.SessionDescription();
     pb_sd.setType(pb.SessionDescription.SDPType.SDP_TYPE_OFFER);
     pb_sd.setSdp(offer);
     let uint8_array = pb_sd.serializeBinary();
     console.log("sending local offer to the server...");
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

   let handleWebsocketEvent = (event) => {
     if ( Object.getPrototypeOf(event) === MessageEvent.prototype ) {
       let remote_sdp_answer = new pb.SessionDescription.deserializeBinary(event.data);
       if (remote_sdp_answer) {
         console.log("received a remote answer from server...");
         peerConnection.setRemoteDescription({
           type: "answer",
           sdp: remote_sdp_answer.getSdp()
         });
       } else {
         console.error("not an sdp");
       }
     } else {
       console.log(`Received event with prototype of ${Object.getPrototypeOf(event)}`);
     }
   }

   let startRemoteSession = (remoteVideoNode) => {
     let pc;

     return Promise.resolve().then(() => {
       pc = new RTCPeerConnection({
         iceServers: [{ urls: 'stun:stun.l.google.com:19302' }]
       });

       // game streaming
       pc.addTransceiver('audio', {'direction': 'recvonly'});
       pc.addTransceiver('video', {'direction': 'recvonly'});

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
       return createOffer(pc);
     }).then(offer => {
       startSession(offer);
       updateFailed(true);
     }).then(() => pc);
   }

   webSocket.addEventListener("open", event => {
     var remoteVideo = document.querySelector('#remote-video');
     if (!peerConnection) {
       console.log("ws open");
       startRemoteSession(remoteVideo).then(pc => {
         remoteVideo.style.setProperty('visibility', 'visible');
         peerConnection = pc;
       }).catch((error) => { console.error(error); });
     }
   });

   window.addEventListener('beforeunload', () => {
     if (peerConnection) {
       peerConnection.close();
     }
     if (webSocket) {
       webSocket.close();
     }
   });
   return () => {
     if (peerConnection) {
       peerConnection.close();
     }
     if (webSocket) {
       webSocket.close();
     }
   }
 }, [props.gameId])

  return failed ? (
    <div style={{
      border: "dashed 5px white",
      fontFamily: "'Press Start 2P', cursive",
      color: "white",
      textAlign: "center",
      fontSize: "20px",
      width: "100%",
      display: "flex",
      justifyContent: "center",
      alignItems: "center",
      flexDirection: "column",
      padding: "30px"

    }}>
      <a>Sorry, this game is not available now.</a>
      <br/>
      <br/>
      <a>Looks like our free server is taking a toll!</a>
      <br/>
      <a>You can still play
        <Button style={{
          color: "white",
          marginLeft: "20px",
          marginRight: "5px",
          fontSize: "20px",
          fontFamily: "'Press Start 2P', cursive",
          padding: "20px",
          border: "dashed 5px white",
          height: "20px",
          borderRadius: 0,
        }} onClick={() => history.push(`/${props.roomId}/Taboo`)}>
        Taboo
        </Button>
        .
      </a>
      <br/>
      <a>It's one of our favorites.</a>
    </div>
    ) :
    (
    <video id="remote-video" autoPlay muted playsInline style={{width: "100%", height: "100%"}}></video>
  )
}

export default GameLovers;