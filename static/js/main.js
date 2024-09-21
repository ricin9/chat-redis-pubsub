document.addEventListener("DOMContentLoaded", function () {
  const currentRoomId = location.pathname.split("/").pop();
  if (!currentRoomId || Number.isNaN(Number(currentRoomId))) {
    return;
  }

  const messages = document.getElementById(`room-${currentRoomId}-messages`);
  if (!messages) {
    return;
  }

  messages.scrollTo(0, messages.scrollHeight);
});
document.addEventListener("htmx:wsAfterMessage", function (event) {
  if (!event.detail.message.includes("data-room-id")) {
    return;
  }

  const currentRoomId = location.pathname.split("/").pop();
  if (!currentRoomId || Number.isNaN(Number(currentRoomId))) {
    return;
  }

  const wsElt = document.createElement("div");
  wsElt.innerHTML = event.detail.message;
  const IncomingMessageRoomId = wsElt.firstChild.getAttribute("data-room-id");

  if (currentRoomId !== IncomingMessageRoomId) {
    incrementUnread(IncomingMessageRoomId);
    return;
  }

  // scroll messages to bottom
  const messages = document.getElementById(`room-${currentRoomId}-messages`);
  if (!messages) {
    return;
  }
  // Check if user is near the bottom
  const isAtBottom =
    messages.scrollHeight - messages.scrollTop <= messages.clientHeight + 100;

  if (isAtBottom) {
    messages.scrollTo(0, messages.scrollHeight);
    if (!document.hasFocus()) {
      incrementUnread(currentRoomId);
    }
  } else {
    incrementUnread(currentRoomId);
  }

  const room = document.getElementById(`room-${currentRoomId}`);
  const roomList = document.getElementById("room-list");

  if (room) {
    roomList.prepend(room);
  }
});

function incrementUnread(roomId) {
  const room = document.getElementById(`room-${roomId}`);
  if (!room) {
    return;
  }

  const unread = room.children[0].children[1];
  if (!unread) {
    return;
  }

  const count = parseInt(unread.innerText) || 0;
  unread.innerText = count + 1;

  const countInTitle = document.title.split(" ")[0];
  if (
    countInTitle[0] === "(" &&
    countInTitle[countInTitle.length - 1] === ")"
  ) {
    const count = parseInt(countInTitle.slice(1, countInTitle.length - 1)) || 0;
    document.title = `(${count + 1}) ${document.title
      .split(" ")
      .slice(1)
      .join(" ")}`;
  } else {
    document.title = `(1) ${document.title}`;
  }
}

function resetUnread(roomId) {
  if (!document.hasFocus()) {
    return;
  }
  const room = document.getElementById(`room-${roomId}`);
  if (!room) {
    return;
  }

  const unread = room.children[0].children[1];
  if (!unread) {
    return;
  }

  const currentCount = parseInt(unread.innerText) || 0;
  if (currentCount === 0) {
    return;
  }

  unread.innerText = "";

  const countInTitle = document.title.split(" ")[0];
  if (
    countInTitle[0] !== "(" ||
    countInTitle[countInTitle.length - 1] !== ")"
  ) {
    return;
  }

  const count = parseInt(countInTitle.slice(1, countInTitle.length - 1)) || 0;
  const newCount = count - currentCount;

  if (newCount <= 0) {
    document.title = document.title.split(" ").slice(1).join(" ");
    return;
  }

  document.title = `(${newCount}) ${document.title
    .split(" ")
    .slice(1)
    .join(" ")}`;
}

document.addEventListener("focus", function (event) {
  if (!document.hasFocus()) {
    return;
  }
  const currentRoomId = location.pathname.split("/").pop();
  if (!currentRoomId || Number.isNaN(Number(currentRoomId))) {
    return;
  }

  const messages = document.getElementById(`room-${currentRoomId}-messages`);
  if (!messages) {
    return;
  }

  if (messages.scrollHeight - messages.scrollTop === messages.clientHeight) {
    resetUnread(currentRoomId);
  }
});
