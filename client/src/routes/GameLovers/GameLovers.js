import React, {useEffect, useState} from "react";
import { DCLabel } from "./datachannel";
import { InputMap } from "./input";
import Button from "@material-ui/core/Button";
import {gamecode} from "../TabooGame/src/__fixtures__/game";
import {useHistory} from "react-router";
import DialogContent from "@material-ui/core/DialogContent";
import {DialogActions, Grid, TextField, Typography} from "@material-ui/core";
import Dialog from "@material-ui/core/Dialog";
import {Transition, useStyles} from "../CreateRoom";
const pb = require('./proto/signaling_pb');
const input = require('./proto/input_pb');

// const SERVER_ADDR = "34.94.73.231";
const SERVER_ADDR = "w2.zoomgaming.app";

const GameLovers = (props) => {

  const classes = useStyles();
  const history = useHistory();

  const [failed, updateFailed] = useState(false);

  const [open, setOpen] = React.useState(true);

  const handleClickOpen = () => {
    setOpen(true);
  };

  const handleClose = () => {
    setOpen(false);
  };

 useEffect(() => {
   let peerConnection = null; // webrtc connection
   let input_dc = null; // keyboard events are sent to the server using this
   let webSocket = new WebSocket(`wss://${SERVER_ADDR}/demo/${props.roomId}/${props.gameId}`); // session description is sent/received via websocket
   webSocket.binaryType = "arraybuffer" // blob or arraybuffer
   // webSocket.addEventListener("open", event => { console.log("ws open"); });
   webSocket.addEventListener("message", event => { handleWebsocketEvent(event); });
   webSocket.addEventListener("close", event => { console.log("ws closing"); updateFailed(true) });
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
       updateFailed(false);
     }).then(() => pc);
   }

   webSocket.addEventListener("open", event => {
     var remoteVideo = document.querySelector('#remote-video');
     if (!peerConnection) {
       console.log("ws open");
       startRemoteSession(remoteVideo).then(pc => {
         remoteVideo.style.setProperty('visibility', 'visible');
         remoteVideo.volume = 0.15;
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

  const getControls = (gameId) => {
   switch (gameId) {
     case ("SpaceTime"):
       return [
         {action: "Move Up", control: "Up Arrow"},
         {action: "Move Down", control: "Down Arrow"},
         {action: "Move Left", control: "Left Arrow"},
         {action: "Move Right", control: "Right Arrow"},
         {action: "Jump", control: "Space"},
         {action: "Fire/Use", control: "D"},
         {action: "Back/Cancel", control: "S"},
         {action: "Space-Set", control: "A"},
       ]
     case ("Broforce"):
       return [
         {action: "Move Up", control: "Up Arrow"},
         {action: "Move Down", control: "Down Arrow"},
         {action: "Move Left", control: "Left Arrow"},
         {action: "Move Right", control: "Right Arrow"},
         {action: "Jump", control: "Up Arrow"},
         {action: "Fire", control: "D"},
         {action: "Grenade", control: "A"},
         {action: "Melle", control: "S"},
         {action: "Flex", control: "Space"},
       ]
     default:
       return []
   }
  }


  return failed ?
    (
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
      <>
        <video id="remote-video" autoPlay playsInline style={{width: "100%", height: "100%"}}></video>
        <Dialog
          open={open && getControls(props.gameId).length}
          TransitionComponent={Transition}
          className={classes.mobileDialog}
          keepMounted
          onClose={handleClose}
          aria-labelledby="alert-dialog-slide-title"
          aria-describedby="alert-dialog-slide-description"
        >
          <DialogContent className={classes.centerAlign} style={{margin: "20px"}}>
            <Grid container>
              <Grid container item xs={12} style={{paddingBottom: "20px"}}>
                <Grid container xs={6}>
                  <Typography className={classes.typography}>Action:</Typography>
                </Grid>
                <Grid container xs={6}>
                  <Typography className={classes.typography}>Control:</Typography>
                </Grid>
              </Grid>
              {getControls(props.gameId).map(control => (
                <Grid container item xs={12}>
                  <Grid container xs={6}>
                    <Typography className={classes.typography}>
                      {control.action}
                    </Typography>
                  </Grid>
                  <Grid container xs={6}>
                    <Typography className={classes.typography}>
                      {control.control}
                    </Typography>
                  </Grid>
                </Grid>
              ))}
            </Grid>
          </DialogContent>
          <DialogActions className={classes.centerAlign} style={{paddingBottom: "20px"}}>
            <Button className={classes.buttonGame} onClick={() => setOpen(false)}>
              Close
            </Button>
          </DialogActions>
        </Dialog>
      </>
  )
}

export default GameLovers;