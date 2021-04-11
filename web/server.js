const path = require("path");

//set up express server
const express = require('express')
const http = require('http')

const PORT = process.env.PORT || 4000

const app = express()

// Serve static html and javascript
app.use(express.static(path.join(__dirname, "dist")));
const server = http.createServer(app)

server.listen(PORT, ()=>console.log(`Listening on port ${PORT}`))
