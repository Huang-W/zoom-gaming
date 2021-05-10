require('dotenv').config();
const express = require("express");
const http = require("http");
const socket = require("socket.io");
const path = require("path");
const wrtc = require('wrtc');
const Peer = require("simple-peer");

const app = express();
const server = http.createServer(app);
const io = socket(server);

let roomIDLocal = "";

// map of RoomID -> { id: socketId, peer: Peer }
let users = new Map();

// map of RoomID -> { id: socketId, stream: MediaStream }
let streams = new Map();

// map of socketId -> roomID
const socketToRoom = new Map();
const NEW_CHAT_MESSAGE_EVENT = "newChatMessage";

io.on('connection', socket => {
    socket.on("join room", roomID => {
        roomIDLocal = roomID;
        socket.join(roomID);

        if (!users.has(roomID)) {
            users.set(roomID, []);
            streams.set(roomID, []);
        }

        const peer = new Peer({
            initiator: true,
            trickle: false,
            wrtc: wrtc,
        });

        // sending an offer
        peer.on('signal', signal => {
            socket.emit("sending signal", signal);
        });

        peer.addTransceiver("audio", { direction: "sendrecv" });
        peer.addTransceiver("video", { direction: "sendrecv" });

        let existingStreams = streams.get(roomID);
        for (let obj of existingStreams) {
            peer.addStream(obj.stream);
        }

        // A new audio/video stream has arrived from the remote peer (should only happen once per connection)
        peer.on('stream', stream => {

            console.log(stream.id);

            // add this stream to the array of existing streams, any new players get initialized with this array of streams
            streams.get(roomID).push({ id: socket.id, stream: stream });

            // notify all current players of a new stream
            const usersInThisRoom = users.get(roomID).filter(obj => obj.id !== socket.id);
            for (let obj of usersInThisRoom) {
                try{
                    obj.peer.addStream(stream);
                } catch (error) {
                    console.log("Error: ", error)
                }
            }
        });

        peer.on('error', (err) => { console.error(err); } );

        // a returning answer
        socket.on('returning signal', signal => {
            peer.signal(signal);
        });

        users.get(roomID).push({ id: socket.id, peer: peer });
        socketToRoom.set(socket.id, roomID);
    });

    // Join a conversation
    // const { roomId } = socket.handshake.query;
    socket.join(roomIDLocal);

    // Listen for new messages
    socket.on(NEW_CHAT_MESSAGE_EVENT, (data) => {
        io.in(roomIDLocal).emit(NEW_CHAT_MESSAGE_EVENT, data);
    });

    // remove traces of this player
    socket.on('disconnect', () => {

        console.log('socket disconnect');

        const roomID = socketToRoom.get(socket.id);

        if (roomID && users.has(roomID)) {

            let currentUsers = users.get(roomID);
            let currentStreams = streams.get(roomID);

            let updatedUsers = currentUsers.filter(conn => conn.id !== socket.id);
            let updatedStreams = currentStreams.filter(s => s.id !== socket.id);

            let leavingUser = currentUsers.find(conn => conn.id === socket.id);
            if (leavingUser) {
                leavingUser.peer.destroy();
            }

            let leavingStream = currentStreams.find(s => s.id === socket.id);
            if (leavingStream) {
                console.log(leavingStream.stream.id);
                updatedUsers.forEach(user => user.peer.removeStream(leavingStream.stream) );
                io.to(roomID).emit('user left', leavingStream.stream.id);
            }

            users.set(roomID, updatedUsers);
            streams.set(roomID, updatedStreams);
            socketToRoom.delete(socket.id);
        }
    });
});

if (process.env.PROD) {
    app.use(express.static(path.join(__dirname, './client/build')));
    app.get('*', (req, res) => {
        res.sendFile(path.join(__dirname, './client/build/index.html'))
    });
}

server.listen(process.env.PORT || 8000, () => console.log('server is running on port 8000'));