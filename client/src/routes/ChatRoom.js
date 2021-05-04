import React from "react";
import useChat from "./useChat";
import {useParams} from "react-router";
import {Box, makeStyles} from "@material-ui/core";
import clsx from "clsx";

const useStyles = makeStyles((theme) => ({
  chatRoomContainer: {
    backgroundColor: "black",
    display: "flex",
    alignItems: "center",
    justifyContent: "center",
    width: "35vw",
    height: "45vw",
    border: "solid 5px white",
    overflow: "-moz-hidden-unscrollable",
  },
  messagesContainer: {
    display: "flex",
    overflow: "auto",
    width: "100%",
    border: "1px",
    borderRadius: 0,
    borderColor: "white",
    fontFamily: "'Press Start 2P', cursive",
  },
  messageItem: {
    width: "60%",
    minWidth: "80px",
    margin: "8px",
    padding: "12px 8px",
    wordBreak: "break-word",
    borderRadius: 0,
    color: "white",
  },
  myMessage: {
    backgroundColor: "#008080",
    marginLeft: "auto",
    fontFamily: "'Press Start 2P', cursive",
  },
  receivedMessage: {
    backgroundColor: "#708090",
    marginRight: "auto",
    fontFamily: "'Press Start 2P', cursive",
  },
  newMessageInputField: {
    height: "100px",
    width: "300px",
    fontSize: "20px",
    padding: "8px 12px",
    borderRadius: 0,
    fontFamily: "'Press Start 2P', cursive",

  },
  sendMessageButton: {
    fontWeight: 600,
    color: "white",
    background: "black",
    padding: "24px 12px",
    width: "200px",
    height: "75px",
    fontFamily: "'Press Start 2P', cursive",
    fontSize: "30px",
    margin: "50px",
    border: "dashed 5px white",
    borderRadius: 0,
  },
  boxPadding: {
    padding: "0.5%",
  },
}));

const ChatRoom = (props) => {
  const { roomID } = useParams(); // Gets roomID from URL
  const { messages, sendMessage } = useChat(roomID); // Creates a websocket and manages messaging
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
    <Box className={classes.chatRoomContainer} flexDirection={"column"}>
      <Box flexGrow={1} className={classes.messagesContainer}>
        <ul style={{listStyleType: "none", padding: 0, width: "100%"}}>
          {messages.map((message, i) => (
            <li
              key={i}
              className={clsx(classes.messageItem,
                message.ownedByCurrentUser ? classes.myMessage : classes.receivedMessage
              )}
            >
              {message.body}
            </li>
          ))}
        </ul>
      </Box>
      <Box>
        <textarea
          value={newMessage}
          onChange={handleNewMessageChange}
          placeholder="Write message..."
          className= {classes.newMessageInputField}
        />
      </Box>
      <Box>
        <button onClick={handleSendMessage} className= {classes.sendMessageButton}>
          Send
        </button>
      </Box>
    </Box>
  );
};

export default ChatRoom;