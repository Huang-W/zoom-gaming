require('dotenv').config();
const express = require("express");
const http = require("http");
const app = express();
const server = http.createServer(app);
const socket = require("socket.io");
const io = socket(server);
const path = require("path");
let roomIDLocal = "";

const users = {};

const socketToRoom = {};
const NEW_CHAT_MESSAGE_EVENT = "newChatMessage";

io.on('connection', socket => {
    socket.on("join room", roomID => {
        roomIDLocal = roomID
        if (users[roomID]) {
            users[roomID].push(socket.id);
        } else {
            users[roomID] = [socket.id];
        }
        socketToRoom[socket.id] = roomID;
        const usersInThisRoom = users[roomID].filter(id => id !== socket.id);

        socket.emit("all users", usersInThisRoom);
    });

    // Join a conversation
    // const { roomId } = socket.handshake.query;
    socket.join(roomIDLocal);

    // Listen for new messages
    socket.on(NEW_CHAT_MESSAGE_EVENT, (data) => {
        io.in(roomIDLocal).emit(NEW_CHAT_MESSAGE_EVENT, data);
    });

    socket.on("sending signal", payload => {
        io.to(payload.userToSignal).emit('user joined', { signal: payload.signal, callerID: payload.callerID });
    });

    socket.on("returning signal", payload => {
        io.to(payload.callerID).emit('receiving returned signal', { signal: payload.signal, id: socket.id });
    });

    socket.on('disconnect', () => {
        const roomID = socketToRoom[socket.id];
        let room = users[roomID];
        if (room) {
            room = room.filter(id => id !== socket.id);
            users[roomID] = room;
            socket.broadcast.emit('user left', socket.id)
        }
    });

});

if (process.env.PROD) {
    app.use(express.static(path.join(__dirname, './client/build')))
    app.get('*', (req, res) => {
        res.sendFile(path.join(__dirname, './client/build/index.html'))
    })
}

server.listen(process.env.PORT || 8000, () => console.log('server is running on port 8000'));


