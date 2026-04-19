const winston = require('winston')

const logger = winston.createLogger({
    level: 'info',
    format: winston.format.combine(
        winston.format.timestamp({ format: 'HH:mm:ss' }),
        winston.format.errors({ stack: true }),
        winston.format.colorize(),
        winston.format.printf(({ timestamp, level, message, stack, ...meta }) => {
            const metaStr = Object.keys(meta).length ? ' ' + JSON.stringify(meta) : ''
            return `${timestamp} [${level}] ${message}${metaStr}${stack ? '\n' + stack : ''}`
        })
    ),
    transports: [new winston.transports.Console()]
})

const server = require('http').createServer()

const ORIGIN_HOST = process.env.ORIGIN_HOST || "http://localhost:5173"

const io = require('socket.io')(server, {
    cors: { origin: ORIGIN_HOST }
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
        address: socket.handshake.headers["x-real-ip"],
        username: socket.handshake.auth.username
    }

    logger.info('Frontend connected, opening SSH tunnel...', mdc)

    const cols = clampSize(parseInt(socket.handshake.query.cols) || 80)
    const rows = clampSize(parseInt(socket.handshake.query.rows) || 24)

    const username = socket.handshake.auth.username
    const key = socket.handshake.auth.key

    if (!username || !key) {
        logger.error("username and/or ssh key is not provided", mdc)
        socket.disconnect(true)
        return
    }

    const ssh = new SSHClient();

    ssh
        .on('banner', msg => socket.emit('data', msg))
        .on('close', () => socket.disconnect(true))
        .on('end', () => socket.disconnect(true))
        .on('error', err => {
            logger.error("SSH error", { ...mdc, err })
            socket.emit('data', '\r\n*** SSH CONNECTION ERROR: ' + err.message + ' ***\r\n')
            ssh.end()
            socket.disconnect(true)
        })
        .on('ready', () => {
            ssh.shell({ term: 'xterm-256color', clicolor_force: '1', cols, rows }, (err, stream) => {
                if (err) {
                    logger.error("SSH shell error", { ...mdc, err })
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
                        stream.writable) {

                        stream.setWindow(clampSize(rows), clampSize(cols), 0, 0)
                    }
                })

                stream
                    .on('data', data => socket.emit('data', data))
                    .on('error', err => {
                        logger.error("SSH stream error", { ...mdc, err })
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
        logger.info("Frontend disconnected, closing SSH tunnel.", mdc)
        try {
            ssh.end()
        } catch (err) {
            logger.error("Failed to close ssh connection", { ...mdc, err })
        }
    })
})

server.listen(SERVER_PORT, () => {
    logger.info("Web backend app started", { port: SERVER_PORT })
})

server.on('error', err => {
    if (err.code === 'EADDRINUSE') {
        logger.error("Port is already in use", { port: SERVER_PORT, err })
    } else {
        logger.error("Failed to start application", { port: SERVER_PORT, err })
    }
    process.exit(1)
})

function clampSize(size) {
    return Math.max(Math.min(size, MAX_ALLOWED_SCREEN_SIZE), MIN_ALLOWED_SCREEN_SIZE)
}
