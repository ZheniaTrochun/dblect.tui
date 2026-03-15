<script setup>
  import {Terminal} from "@xterm/xterm";
  import '@xterm/xterm/css/xterm.css';
  import { FitAddon } from '@xterm/addon-fit';
  import {onMounted, onUnmounted, ref} from "vue";
  import {io} from 'socket.io-client'
  import ssh from 'micro-key-producer/ssh.js'
  import { randomBytes } from 'micro-key-producer/utils.js';

  const SSH_KEYS_KEY = "DBLECT.tui_keys"

  const terminalContainer = ref(null)

  const term = new Terminal({ cursorBlink: true, theme: { background: '#333' }, lineHeight: 1, fontSize: 14})
  const fitAddon = new FitAddon()
  term.loadAddon(fitAddon)

  let socket
  let backgroundReconnect

  onMounted(() => {
    term.open(terminalContainer.value)
    fitAddon.fit()
    window.addEventListener('resize', handleWindowResize)

    const creds = checkIfKeysExist() ? getKeys() : generateKeysPair()

    socket = setupSocket(term, creds)

    socket.on("disconnect", () => {
      console.log("Socket disconnected, retry connection in 1 sec")
      backgroundReconnect = setTimeout(() => {
        console.log("Retrying socket connection")
        socket = setupSocket(term, creds)
        socket.connect()
      }, 1000)
    })

    term.focus();
    socket.connect();
  })

  onUnmounted(() => {
    if (socket) {
      socket.disconnect()
    }

    if (backgroundReconnect) {
      clearTimeout(backgroundReconnect)
    }

    window.removeEventListener('resize', handleWindowResize)

    term.dispose()
  })

  function setupSocket(term, credentials) {
    const socket = io({ autoConnect: false, query: {
        cols: term.cols,
        rows: term.rows
    },
    auth: {
        username: credentials.username,
        key: credentials.privateKey
    }})

    term.onResize(({ cols, rows }) => socket.emit('resize', { cols, rows }))
    term.onData(data => socket.emit('data', data))

    socket.on('data', data => {
      if (data instanceof ArrayBuffer) {
        term.write(new Uint8Array(data))
      } else {
        term.write(data)
      }
    })

    socket.on("connect_error", (err) => {
      console.error(`Connection failed: ${err.message}`)

      term.write(`\r\nConnection failed: ${err.message}\r\n`)
    })

    socket.on('disconnect', () => {
      term.clear()
      term.write("Connection lost, retrying...")
    })

    return socket
  }

  function checkIfKeysExist() {
    return !!localStorage.getItem(SSH_KEYS_KEY)
  }

  function generateKeysPair() {
    const username = prompt("Please let me know who are you.")

    const seed = randomBytes(32)
    const {fingerprint, publicKey, privateKey} = ssh(seed, username)

    const creds = {
      username,
      fingerprint,
      publicKey,
      privateKey
    }

    const serializedCreds = JSON.stringify(creds)

    localStorage.setItem(SSH_KEYS_KEY, serializedCreds)

    return creds
  }

  function getKeys() {
    const serializedCreds = localStorage.getItem(SSH_KEYS_KEY)
    return JSON.parse(serializedCreds)
  }

  function handleWindowResize() {
    fitAddon.fit()
  }
</script>

<template>
    <div id="terminal-container" ref="terminalContainer"></div>
</template>

<style>
#terminal-container {
  width: 100%;
  height: 100vh;
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}
</style>
