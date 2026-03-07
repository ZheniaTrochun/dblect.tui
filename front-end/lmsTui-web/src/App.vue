<script setup>
import {Terminal} from "@xterm/xterm";
import '@xterm/xterm/css/xterm.css';
import { FitAddon } from '@xterm/addon-fit';
import {onMounted} from "vue";
// import io from "socket.io/lib/client.js";
import {io} from 'socket.io-client'

onMounted(() => {

  const socket = io('http://localhost:5174', { autoConnect: false });

  const term = new Terminal({ cursorBlink: true, theme: { background: '#333' } });
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
  socket.on('disconnect', () => term.dispose());

  fitAddon.fit();
  term.focus();

  socket.connect();

  // term.resize(120, 30);



  // var socket;
  // var terminalContainer = document.getElementById("terminal-container");
  // var term = new Terminal({ cursorBlink: true });
  // term.open(terminalContainer);
  // // term.fit();
  //
  // const sock = io("ws://localhost:5173")
  //
  // socket = sock.connect();
  // socket.on("connect", function() {
  //   term.write("\r\n*** Connected to backend***\r\n");
  //
  //   // Browser -> Backend
  //   term.on("data", function(data) {
  //     //console.log(data);
  //     //                        alert("Not allowd to write. Please don't remove this alert without permission of Ankit or Samir sir. It will be a problem for server'");
  //     socket.emit("data", data);
  //   });
  //
  //   // Backend -> Browser
  //   socket.on("data", function(data) {
  //     term.write(data);
  //   });
  //
  //   socket.on("disconnect", function() {
  //     term.write("\r\n*** Disconnected from backend***\r\n");
  //   });
  // });

})

</script>

<template>
<!--  <header>-->
<!--    <img alt="Vue logo" class="logo" src="./assets/logo.svg" width="125" height="125" />-->

<!--    <div class="wrapper">-->
<!--      <HelloWorld msg="You did it!" />-->
<!--    </div>-->
<!--  </header>-->

<!--  <main>-->
<!--    <TheWelcome />-->

    <div id="terminal-container"></div>
<!--  </main>-->
</template>

<style>
#terminal-container {
  width: 100%;
  height: 90vh;
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}
</style>

<!--
<style scoped>
header {
  line-height: 1.5;
}

.logo {
  display: block;
  margin: 0 auto 2rem;
}

@media (min-width: 1024px) {
  header {
    display: flex;
    place-items: center;
    padding-right: calc(var(--section-gap) / 2);
  }

  .logo {
    margin: 0 2rem 0 0;
  }

  header .wrapper {
    display: flex;
    place-items: flex-start;
    flex-wrap: wrap;
  }
}
</style>
-->