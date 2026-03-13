<script setup>
  import {Terminal} from "@xterm/xterm";
  import '@xterm/xterm/css/xterm.css';
  import { FitAddon } from '@xterm/addon-fit';
  import {onMounted} from "vue";
  import {io} from 'socket.io-client'

  onMounted(() => {

    const socket = io({ autoConnect: false });

    // const term = new Terminal({ cursorBlink: true, theme: { background: '#333' }, lineHeight: 1, fontSize: 10});
    const term = new Terminal({ cursorBlink: true, theme: { background: '#333' }, lineHeight: 1, fontSize: 14});
    const fitAddon = new FitAddon();
    term.loadAddon(fitAddon);
    const terminalContainer = document.getElementById('terminal-container');
    term.open(terminalContainer);

    term.onResize(({ cols, rows }) => socket.emit('resize', { cols, rows }));
    window.addEventListener('resize', () => fitAddon.fit());

    term.onData(data => socket.emit('data', data));

    socket.on('data', data => {
      if (data instanceof ArrayBuffer) {
        term.write(new Uint8Array(data));
      } else {
        console.log(data)
        term.write(data);
      }
    });

    socket.on('connect', () => socket.emit('resize', { cols: term.cols, rows: term.rows }));

    socket.on('disconnect', () => {
      term.clear()
      term.write("Disconnected")
      // term.dispose()
      socket.connect()
    });

    term.focus();
    socket.connect();

    setTimeout(() => {
      fitAddon.fit();
    }, 250);

    setTimeout(() => {
      socket.emit('resize', { cols: term.cols, rows: term.rows })
    }, 250);
  })
</script>

<template>
    <div id="terminal-container"></div>
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
