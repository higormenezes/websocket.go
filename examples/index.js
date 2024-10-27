/** @type {HTMLTextAreaElement} */
const wsMessageTextAreaElement = document.getElementById("ws-message-textarea");
const wsConnectButtonElement = document.getElementById("ws-connect-button");
const wsSendMessageButtonElement = document.getElementById(
  "ws-send-message-button"
);

/** @type {WebSocket | null} */
let exampleSocket = null;

function connect() {
  console.log("Connecting...");
  exampleSocket = new WebSocket("ws://localhost:8090/ws");
  wsConnectButtonElement.innerHTML = "Disconnect";
  wsMessageTextAreaElement.disabled = false;
  wsSendMessageButtonElement.disabled = false;
}

function disconnect() {
  console.log("Disconnecting...");
  exampleSocket.close();
  exampleSocket = null;
  wsConnectButtonElement.innerHTML = "Connect";
  wsMessageTextAreaElement.disabled = true;
  wsSendMessageButtonElement.disabled = true;
}

wsConnectButtonElement.addEventListener("click", () => {
  const isConnected = exampleSocket !== null;
  isConnected ? disconnect() : connect();
});

wsSendMessageButtonElement.addEventListener("click", () => {
  const message = wsMessageTextAreaElement.value;
  if (!message || !exampleSocket) {
    return;
  }

  console.log("Sending the message: ", message);
  exampleSocket.send(message);
});
