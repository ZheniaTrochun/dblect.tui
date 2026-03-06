// var express = require('express');
// var logger = require('morgan');
//
// const http = require("http");
// var SSHClient = require("ssh2").Client;
// var utf8 = require("utf8");
//
// const app = express();
//
//
// var serverPort = 5174;
//
// var server = http.createServer(app);
//
// app.use(logger('dev'));
// app.use(express.json());
// app.use(express.urlencoded({ extended: false }));
//
// server.listen(serverPort);
//
// //socket.io instantiation
// const io = require("socket.io")(server);
//
// //Socket Connection
//
// io.on("connection", function(socket) {
//   var ssh = new SSHClient();
//   ssh
//       .on("ready", function() {
//         socket.emit("data", "\r\n*** SSH CONNECTION ESTABLISHED ***\r\n");
//         connected = true;
//         ssh.shell(function(err, stream) {
//           if (err)
//             return socket.emit(
//                 "data",
//                 "\r\n*** SSH SHELL ERROR: " + err.message + " ***\r\n"
//             );
//           socket.on("data", function(data) {
//             stream.write(data);
//           });
//           stream
//               .on("data", function(d) {
//                 socket.emit("data", utf8.decode(d.toString("binary")));
//               })
//               .on("close", function() {
//                 ssh.end();
//               });
//         });
//       })
//       .on("close", function() {
//         socket.emit("data", "\r\n*** SSH CONNECTION CLOSED ***\r\n");
//       })
//       .on("error", function(err) {
//         console.log(err);
//         socket.emit(
//             "data",
//             "\r\n*** SSH CONNECTION ERROR: " + err.message + " ***\r\n"
//         );
//       })
//       .connect({
//         host: "0.0.0.0",
//         port: "23234", // Generally 22 but some server have diffrent port for security Reson
//         username: "test_from_portal", // user name
//           password: "pass"
//       });
// });




const app = require('express')();
const server = require('http').createServer(app);
const io = require('socket.io')(server, {
    cors: { origin: "http://localhost:5173" }
});
const { Client: SSHClient } = require('ssh2');

const SSH_HOST = 'localhost';
const SSH_PORT = 23234;
const SERVER_PORT = 5174;

io.on('connection', socket => {
    console.log('Frontend connected, opening SSH tunnel...');

    const cols = parseInt(socket.handshake.query.cols) || 80;
    const rows = parseInt(socket.handshake.query.rows) || 24;

    const ssh = new SSHClient();

    ssh
        .on('banner', msg => console.log('SSH banner:', msg))
        .on('close', () => console.log('SSH close'))
        .on('end', () => console.log('SSH end'))
        .on('error', err => console.error('SSH error:', err))
        .on('ready', () => {
            socket.emit('data', '\r\n*** SSH CONNECTION ESTABLISHED ***\r\n');
            ssh.shell({ term: 'xterm-256color', cols, rows }, (err, stream) => {
                if (err) {
                    socket.emit('data', '\r\n*** SSH SHELL ERROR: ' + err.message + ' ***\r\n');
                    return;
                }

                socket.on('data', data => stream.write(data));
                socket.on('resize', ({ cols, rows }) => stream.setWindow(rows, cols, 0, 0));

                stream
                    .on('data', data => socket.emit('data', data))
                    .on('close', () => ssh.end());
            });
        })
        .connect({
            host: SSH_HOST,
            port: SSH_PORT,
            username: 'user',
            password: 'asd123',
        });

    socket.on('disconnect', () => {
        console.log('Frontend disconnected, closing SSH tunnel.');
        ssh.end();
    });
});

server.listen(SERVER_PORT, () => {
    console.log(`App ready on :${SERVER_PORT}`);
});