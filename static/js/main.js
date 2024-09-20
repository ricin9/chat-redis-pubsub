document.addEventListener("htmx:wsBeforeMessage", function (event) {
  if (!event.detail.message.includes("data-room-id")) {
    return;
  }

  const currentRoom = document.querySelector(".selected-room");
  const currentRoomId = currentRoom
    ? currentRoom.id.replace("room-", "")
    : null;

  if (!currentRoomId) {
    return;
  }

  const wsElt = document.createElement("div");
  wsElt.innerHTML = event.detail.message;
  const IncomingMessageRoomId = wsElt.firstChild.getAttribute("data-room-id");

  if (currentRoomId !== IncomingMessageRoomId) {
    console.log("Message from another room, not swapping", event.detail);
    //event.detail.shouldSwap = false;
  }
});

document.addEventListener("htmx:wsAfterMessage", function (event) {
  if (!event.detail.message.includes("data-room-id")) {
    return;
  }

  const currentRoom = document.querySelector(".selected-room");
  const currentRoomId = currentRoom
    ? currentRoom.id.replace("room-", "")
    : null;

  if (!currentRoomId) {
    // show notification actually
    return;
  }

  const wsElt = document.createElement("div");
  wsElt.innerHTML = event.detail.message;
  const IncomingMessageRoomId = wsElt.firstChild.getAttribute("data-room-id");

  if (currentRoomId !== IncomingMessageRoomId) {
    // show notification actually
    return;
  }

  // scroll messages to bottom
  const messages = document.querySelector(`room-${currentRoomId}-messages`);
  if (!messages) {
    console.log("not found messaages element");
    return;
  }
  // Check if user is near the bottom
  const isAtBottom =
    messages.scrollHeight - messages.scrollTop <= messages.clientHeight + 50;

  if (isAtBottom) {
    messages.scrollTo(0, messages.scrollHeight);
  }
  // TODO, show notification if user is not at the bottom

  const room = document.getElementById(`room-${currentRoomId}`);
  const roomList = document.getElementById("room-list");

  if (room) {
    roomList.prepend(room);
  }
});
