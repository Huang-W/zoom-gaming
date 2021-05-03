
import React from "react";

import "./ChatRoom.css";
import useChat from "./useChat";
import {useParams} from "react-router";
import {Box, Grid, makeStyles} from "@material-ui/core";

const useStyles = makeStyles((theme) => ({
  centerAlign: {
    display: "flex",
    justifyContent: "center",
    padding: "5px"
  },
  button: {
    color: "white",
    fontFamily: "'Press Start 2P', cursive",
    marginLeft: "40px",
  },
  gameFont: {
    fontFamily: "'Press Start 2P', cursive",
  },
  chatRoomContainer: {
    width: "50vw",
    height: "60vw",
  },
  messagesContainer: {
    display: "flex",
    flexGrow: "1",
    minHeight: "20%",
    overflow: "auto",
    border: "1px",
    borderRadius: "7px",
  },
  myMessage: {
    backgroundColor: "#008080",
    marginLeft: "auto",
  },
  receivedMessage: {
    backgroundColor: "#3f4042",
    marginRight: "auto",
  },
  newMessageInputField: {
    height: "20%",
    maxHeight: "50%",
    fontSize: "20px",
    padding: "8px 12px",
    resize: "none",
  },
  sendMessageButton: {
    fontSize: "28px",
    fontWeight: 600,
    color: "white",
    background: "#31a24c",
    padding: "24px 12px",
    border: "none",
    width: "200px",
    height: "75px",
  },
}));

const ChatRoom = (props) => {
  const { id } = useParams(); // Gets roomId from URL
  const roomId = id;
  const { messages, sendMessage } = useChat(roomId); // Creates a websocket and manages messaging
  const [newMessage, setNewMessage] = React.useState(""); // Message to be sent
  const classes = useStyles();

  const handleNewMessageChange = (event) => {
    setNewMessage(event.target.value);
  };

  const handleSendMessage = () => {
    sendMessage(newMessage);
    setNewMessage("");
  };

  return (
    <Box className={classes.chatRoomContainer}>
      <Grid container spacing={1} direction={"column"}>
        <Grid container item xs={9} className={classes.messagesContainer}>
          <ol>
            {messages.map((message, i) => (
              <li
                key={i}
                className={`message-item ${
                  message.ownedByCurrentUser ? classes.myMessage : classes.receivedMessage
                }`}
              >
                {message.body}
              </li>
            ))}
          </ol>
        </Grid>
        <Grid container item xs={2}>
        <textarea
          value={newMessage}
          onChange={handleNewMessageChange}
          placeholder="Write message..."
          className= {classes.newMessageInputField}
        />
        </Grid>
        <Grid container item xs={1}>
          <button onClick={handleSendMessage} className= {classes.sendMessageButton}>
            Send
          </button>
        </Grid>
      </Grid>
    </Box>
  );
};

export default ChatRoom;