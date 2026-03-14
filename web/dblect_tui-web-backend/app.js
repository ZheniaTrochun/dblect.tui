const app = require('express')();
const server = require('http').createServer(app);

const originHost = process.env.ORIGIN_HOST || "http://localhost:5173"

const io = require('socket.io')(server, {
    cors: { origin: originHost }
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
            ssh.shell({ term: 'xterm-256color', cols, rows }, (err, stream) => {
                if (err) {
                    socket.emit('data', '\r\n*** SSH SHELL ERROR: ' + err.message + ' ***\r\n');
                    return;
                }

                socket.on('data', data => stream.write(data));
                socket.on('resize', ({ cols, rows }) => {
                    stream.setWindow(rows, cols, 0, 0)
                });

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