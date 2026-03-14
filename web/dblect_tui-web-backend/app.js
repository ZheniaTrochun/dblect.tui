const app = require('express')()
const logger = require('pino')()
const server = require('http').createServer(app)

const originHost = process.env.ORIGIN_HOST || "http://localhost:5173"

const io = require('socket.io')(server, {
    cors: { origin: originHost }
});

const { Client: SSHClient } = require('ssh2')

const SSH_HOST = 'localhost'
const SSH_PORT = 23234
const SERVER_PORT = 5174

const MAX_ALLOWED_SCREEN_SIZE = 500
const MIN_ALLOWED_SCREEN_SIZE = 24

io.on('connection', socket => {
    const mdc = {
        socketId: socket.id,
        address: socket.handshake.address,
        username: socket.handshake.query.username
    }

    logger.info(mdc, 'Frontend connected, opening SSH tunnel...')

    const cols = clampSize(parseInt(socket.handshake.query.cols) || 80)
    const rows = clampSize(parseInt(socket.handshake.query.rows) || 24)

    const username = socket.handshake.query.username
    const key = socket.handshake.query.key

    if (!username || !key) {
        logger.error(mdc, "username and/or ssh key is not provided")
        socket.disconnect(true)
        return
    }

    const ssh = new SSHClient();

    ssh
        .on('banner', msg => socket.emit(msg))
        .on('close', () => socket.disconnect(true))
        .on('end', () => socket.disconnect(true))
        .on('error', err => {
            logger.error({ ...mdc, err }, "SSH error")
            socket.emit('data', '\r\n*** SSH CONNECTION ERROR: ' + err.message + ' ***\r\n')
            ssh.end()
            socket.disconnect(true)
        })
        .on('ready', () => {
            ssh.shell({ term: 'xterm-256color', cols, rows }, (err, stream) => {
                if (err) {
                    logger.error({ ...mdc, err }, "SSH shell error")
                    socket.emit('data', '\r\n*** SSH SHELL ERROR: ' + err.message + ' ***\r\n')
                    ssh.end()
                    socket.disconnect(true)
                    return
                }

                socket.on('data', data => {
                    if (stream.writable) {
                        stream.write(data)
                    }
                })

                socket.on('resize', ({ cols, rows }) => {
                    if (typeof cols === 'number' &&
                        typeof rows === 'number' &&
                        stream.writtable) {

                        stream.setWindow(clampSize(rows), clampSize(cols), 0, 0)
                    }
                })

                stream
                    .on('data', data => socket.emit('data', data))
                    .on('error', err => {
                        logger.error({ ...mdc, err }, "SSH stream error")
                        socket.emit('data', `\r\n*** STREAM ERROR: ${err.message} ***\r\n`)
                        ssh.end()
                        socket.disconnect(true)
                    })
                    .on('close', () => ssh.end())
            });
        })
        .connect({
            host: SSH_HOST,
            port: SSH_PORT,
            username,
            privateKey: key
        })

    socket.on('disconnect', () => {
        logger.info(mdc, "Frontend disconnected, closing SSH tunnel.")
        try {
            ssh.end()
        } catch (err) {
            logger.error({ ...mdc, err }, "Failed to close ssh connection")
        }
    })
})

server.listen(SERVER_PORT, () => {
    logger.info({ port: SERVER_PORT }, "Web backend app started")
})

server.on('error', err => {
    if (err.code === 'EADDRINUSE') {
        logger.info({ port: SERVER_PORT, err }, "Port is already in use")
    } else {
        logger.info({ port: SERVER_PORT, err }, "Failed to start application")
    }
    process.exit(1)
})

function clampSize(size) {
    return Math.max(Math.min(size, MAX_ALLOWED_SCREEN_SIZE), MIN_ALLOWED_SCREEN_SIZE)
}
