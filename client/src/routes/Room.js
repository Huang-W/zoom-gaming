import React, { useEffect, useRef, useState } from "react";
import io from "socket.io-client";
import Peer from "simple-peer";
import styled from "styled-components";
import {useParams} from "react-router";
import {Box, ButtonBase, Grid, makeStyles } from "@material-ui/core";
import GameLovers from "./GameLovers/GameLovers";
import AppBar from '@material-ui/core/AppBar';
import Toolbar from '@material-ui/core/Toolbar';
import Button from '@material-ui/core/Button';
import Remote from '../assets/remote.png';
import VideocamIcon from '@material-ui/icons/Videocam';
import VideocamOffIcon from '@material-ui/icons/VideocamOff';
import MicIcon from '@material-ui/icons/Mic';
import MicOffIcon from '@material-ui/icons/MicOff';
import ChatRoom from "./ChatRoom";
import Dialog from '@material-ui/core/Dialog';
import DialogContent from '@material-ui/core/DialogContent';

const Video = (props) => {
    const ref = useRef();

    useEffect(() => {
        props.peer.on("stream", stream => {
            ref.current.srcObject = stream;
        })
    }, []);

    return (
        <StyledVideo playsInline autoPlay ref={ref} />
    );
}


const videoConstraints = {
    "width": 640,
    "height": 480
};

const StyledVideo = styled.video`
    width: 100%;
    height: 480;
`;

const useStyles = makeStyles((theme) => ({
    centerAlign: {
        display: "flex",
        justifyContent: "center",
        padding: "5px"
    },
    videoOptions: {
        display: "flex",
        justifyContent: "center",
        width: "100%",
    },
    logo: {
        flexGrow: 1,
    },
    button: {
        color: "white",
        fontFamily: "'Press Start 2P', cursive",
        marginLeft: "40px",
    },
    gameFont: {
        fontFamily: "'Press Start 2P', cursive",
    },
    chatDialog: {
        display: "flex",
        justifyContent: "center",
        width: "50vw",
        height: "60vw",
        // "& .MuiPaper-root": {
        //     backgroundColor: "white",
        //     border: "solid 5px white",
        //     borderRadius: 0,
        //     fontFamily: "'Press Start 2P', cursive",
        // },
    },
}));

const Room = (props) => {
    const [peers, setPeers] = useState([]);
    const [mic, setMic] = useState(true);
    const [camera, setCamera] = useState(true);
    const classes = useStyles();
    const socketRef = useRef();
    const userVideo = useRef();
    const peersRef = useRef([]);
    const { roomID, gameID } = useParams();

    useEffect(() => {
        socketRef.current = io.connect("/");
        navigator.mediaDevices.getUserMedia({ video: videoConstraints, audio: true }).then(stream => {
            userVideo.current.srcObject = stream;
            socketRef.current.emit("join room", roomID);
            socketRef.current.on("all users", users => {
                const peers = [];
                users.forEach(userID => {
                    const peer = createPeer(userID, socketRef.current.id, stream);
                    peersRef.current.push({
                        peerID: userID,
                        peer,
                    })
                    peers.push({
                        peerID: userID,
                        peer,
                    });
                })
                setPeers(peers);
            })

            socketRef.current.on("user joined", payload => {
                const peer = addPeer(payload.signal, payload.callerID, stream);
                peersRef.current.push({
                    peerID: payload.callerID,
                    peer,
                })

                const peerObj = {
                    peerID: payload.callerID,
                    peer,
                }

                setPeers(users => [...users, peerObj]);
            });

            socketRef.current.on("receiving returned signal", payload => {
                const item = peersRef.current.find(p => p.peerID === payload.id);
                item.peer.signal(payload.signal);
            });

            socketRef.current.on("user left", id => {
                const peerObj = peersRef.current.find(p => p.peerID === id);
                if (peerObj) {
                    peerObj.peer.destroy();
                }
                const peers = peersRef.current.filter(p => p.peerID !== id);
                peersRef.current = peers;
                setPeers(peers);
            })
        })
    }, []);

    function createPeer(userToSignal, callerID, stream) {
        const peer = new Peer({
            initiator: true,
            trickle: false,
            stream,
        });

        peer.on("signal", signal => {
            socketRef.current.emit("sending signal", { userToSignal, callerID, signal })
        })

        return peer;
    }

    function addPeer(incomingSignal, callerID, stream) {
        const peer = new Peer({
            initiator: false,
            trickle: false,
            stream,
        })

        peer.on("signal", signal => {
            socketRef.current.emit("returning signal", { signal, callerID })
        })

        peer.signal(incomingSignal);

        return peer;
    }

    const handleMicClick = (e) => {
        userVideo.current.srcObject.getAudioTracks()[0].enabled = !mic;
        setMic(!mic);
    }

    const handleCameraClick = (e) => {
        userVideo.current.srcObject.getVideoTracks()[0].enabled = !camera;
        setCamera(!camera);
    }

    const [openChat, setOpenChat] = React.useState(false);
    const handleClickOpenChat = () => {
        console.log("open");
        setOpenChat(true);
    };
    const handleCloseChat = () => {
        setOpenChat(false);
    };

    return (
      <div style={{height: "100vh"}}>
          <AppBar position="static" style={{backgroundColor: "#2b2b2b"}}>
              <Toolbar>
                  <Box className={classes.logo}>
                      <ButtonBase onClick={() => props.history.push(`/`)}>
                          <img src={Remote} style={{height: "60px"}}/>
                      </ButtonBase>
                  </Box>
                  <Button className={classes.button} onClick={() => props.history.push(`/`)}>
                      START NEW GAME
                  </Button>
                  <Button className={classes.button} onClick={handleClickOpenChat}>
                      CHAT
                  </Button>
                  <Dialog
                    open={openChat}
                    keepMounted
                    onClose={handleCloseChat}
                    className={classes.chatDialog}
                    aria-labelledby="alert-dialog-slide-title"
                    aria-describedby="alert-dialog-slide-description"
                  >
                      <DialogContent >
                        <ChatRoom/>
                      </DialogContent>
                  </Dialog>
              </Toolbar>
          </AppBar>
          <Grid container style={{height: "calc(100% - 64px)"}}>
              <Grid item xs={10} className={classes.centerAlign}>
                  <GameLovers roomId={roomID} gameId={gameID}/>
              </Grid>
              <Grid item xs={2} container direction={"column"} className={classes.centerAlign}>
                  <StyledVideo muted ref={userVideo} autoPlay playsInline />
                  {peers.map((peer) => {
                      return (
                        <Video key={peer.peerID} peer={peer.peer} />
                      );
                  })}
              </Grid>
              <Grid className={classes.videoOptions}>
                  {
                      mic ? (
                        <MicIcon onClick={handleMicClick} style={{ color: 'white', fontSize: 45, marginRight: 15 }}/>
                      ) : (
                        <MicOffIcon onClick={handleMicClick} style={{ color: 'red', fontSize: 45, marginRight: 15 }}/>
                      )
                  }
                  {
                      camera ? (
                        <VideocamIcon onClick={handleCameraClick} style={{ color: 'white', fontSize: 45, marginLeft: 15 }}/>
                      ) : (
                        <VideocamOffIcon onClick={handleCameraClick} style={{ color: 'red', fontSize: 45, marginLeft: 15 }}/>
                      )
                  }
              </Grid>
          </Grid>
      </div>
    );
};

export default Room;
