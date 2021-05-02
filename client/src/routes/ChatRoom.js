
import React from "react";

import "./ChatRoom.css";
import useChat from "./useChat";
import {useParams} from "react-router";
import {Box, makeStyles} from "@material-ui/core";

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
      <div className="messages-container">
        <ol className="messages-list">
          {messages.map((message, i) => (
            <li
              key={i}
              className={`message-item ${
                message.ownedByCurrentUser ? "my-message" : "received-message"
              }`}
            >
              {message.body}
            </li>
          ))}
        </ol>
      </div>
      <textarea
        value={newMessage}
        onChange={handleNewMessageChange}
        placeholder="Write message..."
        className="new-message-input-field"
      />
      <button onClick={handleSendMessage} className="send-message-button">
        Send
      </button>
    </Box>
  );
};

export default ChatRoom;