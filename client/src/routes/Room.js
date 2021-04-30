import React, { useEffect, useRef, useState } from "react";
import io from "socket.io-client";
import Peer from "simple-peer";
import styled from "styled-components";
import {useParams} from "react-router";
import {Box, ButtonBase, Grid, makeStyles, Menu, MenuItem} from "@material-ui/core";
import GameLovers from "./GameLovers/GameLovers";
import AppBar from '@material-ui/core/AppBar';
import Toolbar from '@material-ui/core/Toolbar';
import Button from '@material-ui/core/Button';
import Remote from '../assets/remote.png';
import VideocamIcon from '@material-ui/icons/Videocam';
import VideocamOffIcon from '@material-ui/icons/VideocamOff';
import MicIcon from '@material-ui/icons/Mic';
import MicOffIcon from '@material-ui/icons/MicOff';

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
    }
}));

const Room = (props) => {
    const [peers, setPeers] = useState([]);
    const [mic, setMic] = useState(true);
    const [camera, setCamera] = useState(true);
    const [anchorEl, setAnchorEl] = React.useState(null);
    const [game, updateGame] = React.useState("SpaceTime");
    const classes = useStyles();
    const socketRef = useRef();
    const userVideo = useRef();
    const peersRef = useRef([]);
    const { id } = useParams();
    const roomID = id;

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
                    peers.push(peer);
                })
                setPeers(peers);
            })

            socketRef.current.on("user joined", payload => {
                const peer = addPeer(payload.signal, payload.callerID, stream);
                peersRef.current.push({
                    peerID: payload.callerID,
                    peer,
                })

                setPeers(users => [...users, peer]);
            });

            socketRef.current.on("receiving returned signal", payload => {
                const item = peersRef.current.find(p => p.peerID === payload.id);
                item.peer.signal(payload.signal);
            });
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

    const isMenuOpen = Boolean(anchorEl);

    const handleMicClick = (e) => {
        userVideo.current.srcObject.getAudioTracks()[0].enabled = !mic;
        setMic(!mic);
    }

    const handleCameraClick = (e) => {
        userVideo.current.srcObject.getVideoTracks()[0].enabled = !camera;
        setCamera(!camera);
    }

    const handleProfileMenuOpen = (event) => {
        setAnchorEl(event.currentTarget);
    };

    const handleMenuClose = () => {
        setAnchorEl(null);
    };

    const selectGame = (game) => {
        const { myValue } = game;
        handleMenuClose();
        updateGame(myValue);
    }

    const menuId = 'primary-search-account-menu';
    const renderMenu = (
      <Menu
        anchorEl={anchorEl}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'right' }}
        id={menuId}
        keepMounted
        transformOrigin={{ vertical: 'bottom', horizontal: 'right' }}
        open={isMenuOpen}
        onClose={handleMenuClose}
      >
          <MenuItem onClick={(e) => selectGame(e.currentTarget.dataset)} className={classes.gameFont} data-my-value={"BroForce"}>BROFORCE</MenuItem>
          <MenuItem onClick={(e) => selectGame(e.currentTarget.dataset)} className={classes.gameFont} data-my-value={"SpaceTime"}>SPACETIME</MenuItem>
      </Menu>
    );

    console.log(peers)

    return (
      <div style={{height: "100vh"}}>
          <AppBar position="static" style={{backgroundColor: "#2b2b2b"}}>
              <Toolbar>
                  <Box className={classes.logo}>
                      <img src={Remote} style={{height: "60px"}}/>
                  </Box>
                  <Button className={classes.button}>
                      START NEW GAME
                  </Button>
                  <Button className={classes.button}>
                      JOIN GAME
                  </Button>
                  <Button
                    className={classes.button}
                    aria-label="account of current user"
                    aria-controls={menuId}
                    aria-haspopup="true"
                    onClick={handleProfileMenuOpen}
                  >
                      SELECT GAME
                  </Button>
              </Toolbar>
          </AppBar>
          {renderMenu}
          <Grid container style={{height: "calc(100% - 64px)"}}>
              <Grid item xs={10} className={classes.centerAlign}>
                  <GameLovers roomId={roomID} gameId={game}/>
              </Grid>
              <Grid item xs={2} container direction={"column"} className={classes.centerAlign}>
                  <StyledVideo muted ref={userVideo} autoPlay playsInline />
                  {peers.map((peer, index) => {
                      return (
                        <Video key={index} peer={peer} />
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
