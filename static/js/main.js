document.addEventListener('DOMContentLoaded', function () {
  const currentRoomId = location.pathname.split('/').pop()
  if (!currentRoomId || Number.isNaN(Number(currentRoomId))) {
    return
  }

  const messages = document.querySelector(`[id^=room-][id$=-messages]`)
  if (!messages) {
    return
  }

  messages.scrollTo(0, messages.scrollHeight)
})

document.addEventListener('htmx:wsAfterMessage', function (event) {
  if (!event.detail.message.includes('data-room-id')) {
    return
  }

  const currentRoomId = location.pathname.split('/').pop()
  if (!currentRoomId || Number.isNaN(Number(currentRoomId))) {
    return
  }

  const wsElt = document.createElement('div')
  wsElt.innerHTML = event.detail.message
  const IncomingMessageRoomId = wsElt.firstChild.getAttribute('data-room-id')

  if (currentRoomId !== IncomingMessageRoomId) {
    incrementUnread(IncomingMessageRoomId)
    const room = document.getElementById(`room-${IncomingMessageRoomId}`)
    const roomList = document.getElementById('room-list')

    if (room) {
      roomList.prepend(room)
    }
    return
  }

  // scroll messages to bottom
  const messages = document.getElementById(`room-${currentRoomId}-messages`)
  if (!messages) {
    return
  }
  // Check if user is near the bottom
  const isAtBottom =
    messages.scrollHeight - messages.scrollTop <= messages.clientHeight + 130

  if (isAtBottom) {
    messages.scrollTo(0, messages.scrollHeight)
    if (!document.hasFocus()) {
      incrementUnread(currentRoomId)
    }
  } else {
    incrementUnread(currentRoomId)
  }

  const room = document.getElementById(`room-${currentRoomId}`)
  const roomList = document.getElementById('room-list')

  if (room) {
    roomList.prepend(room)
  }
})

function incrementUnread(roomId) {
  const room = document.getElementById(`room-${roomId}`)
  if (!room) {
    return
  }

  const unread = room.children[0].children[1]
  if (!unread) {
    return
  }

  const count = parseInt(unread.innerText) || 0
  unread.innerText = count + 1

  const countInTitle = document.title.split(' ')[0]
  if (
    countInTitle[0] === '(' &&
    countInTitle[countInTitle.length - 1] === ')'
  ) {
    const count = parseInt(countInTitle.slice(1, countInTitle.length - 1)) || 0
    document.title = `(${count + 1}) ${document.title
      .split(' ')
      .slice(1)
      .join(' ')}`
  } else {
    document.title = `(1) ${document.title}`
  }
}

function resetUnread(roomId) {
  if (!document.hasFocus()) {
    return
  }
  const room = document.getElementById(`room-${roomId}`)
  if (!room) {
    return
  }

  const unread = room.children[0].children[1]
  if (!unread) {
    return
  }

  const currentCount = parseInt(unread.innerText) || 0
  if (currentCount === 0) {
    return
  }

  unread.innerText = ''

  const countInTitle = document.title.split(' ')[0]
  if (
    countInTitle[0] !== '(' ||
    countInTitle[countInTitle.length - 1] !== ')'
  ) {
    return
  }

  const count = parseInt(countInTitle.slice(1, countInTitle.length - 1)) || 0
  const newCount = count - currentCount

  if (newCount <= 0) {
    document.title = document.title.split(' ').slice(1).join(' ')
    return
  }

  document.title = `(${newCount}) ${document.title
    .split(' ')
    .slice(1)
    .join(' ')}`
}

document.addEventListener('focus', function (event) {
  if (!document.hasFocus()) {
    return
  }
  const currentRoomId = location.pathname.split('/').pop()
  if (!currentRoomId || Number.isNaN(Number(currentRoomId))) {
    return
  }

  const messages = document.getElementById(`room-${currentRoomId}-messages`)
  if (!messages) {
    return
  }

  if (messages.scrollHeight - messages.scrollTop === messages.clientHeight) {
    resetUnread(currentRoomId)
    UpdateLastReadSrv()
  }
})

const sidebar = document.getElementById('sidebar')
const createRoomBtn = document.getElementById('createRoomBtn')
const createRoomModal = document.getElementById('createRoomModal')

function closeSidebar() {
  sidebar.classList.add('-translate-x-full')
}

function openSidebar() {
  sidebar.classList.remove('-translate-x-full')
}
createRoomBtn.addEventListener('click', () => {
  closeSidebar()
  createRoomModal.classList.remove('hidden')
  createRoomModal.classList.add('flex')
})

function closeCreateRoomModal() {
  createRoomModal.classList.add('hidden')
  createRoomModal.classList.remove('flex')
}

function closeRoomInfo() {
  const roomInfoModal = document.getElementById('room-info-modal')
  roomInfoModal.classList.add('hidden')
  roomInfoModal.classList.remove('flex')
}

function toggleRoomMemberActionDropdown(elem) {
  const dropdown = elem.nextElementSibling
  const allDropdowns = document.querySelectorAll('[id^="dropdown"]')

  allDropdowns.forEach((dd) => {
    if (dd.id !== dropdown.id) {
      dd.classList.add('hidden')
    }
  })

  dropdown.classList.toggle('hidden')
}

function handleMessagesScroll(elem) {
  if (elem.scrollHeight - elem.scrollTop === elem.clientHeight) {
    const roomID = location.pathname.split('/').pop()
    resetUnread(roomID)
    UpdateLastReadSrv()
  }
}

function hightlightRoom(roomListItem) {
  const rooms = document.querySelectorAll('.selected-room')
  rooms.forEach((room) => {
    room.classList.remove('selected-room')
  })
  roomListItem.classList.add('selected-room')
}
function handleRoomClick(room) {
  const roomId = room.id.split('-').pop()
  hightlightRoom(room)
  resetUnread(roomId)
  closeSidebar()
}

function getRoomId() {
  const currentRoomId = location.pathname.split('/').pop()
  if (
    currentRoomId &&
    !Number.isNaN(Number(currentRoomId)) &&
    location.pathname.startsWith('/rooms')
  ) {
    return currentRoomId
  }

  console.log(
    'not getting room id',
    currentRoomId,
    Number.isNaN(Number(currentRoomId)),
    location.pathname
  )
  return null
}

htmx.onLoad(function (content) {
  const elem = content.querySelector('#add-member-input')
  if (!elem) return
  new TomSelect(elem, {
    valueField: 'id',
    labelField: 'username',
    searchField: 'username',
    // fetch remote data
    load: function (query, callback) {
      fetch(`/rooms/${getRoomId()}/non-members?q=${encodeURIComponent(query)}`)
        .then((response) => response.json())
        .then((json) => {
          callback(json)
        })
        .catch(() => {
          callback()
        })
    },
    plugins: {
      remove_button: {
        title: 'Remove this member',
      },
    },
  })
})

function UpdateLastReadSrv() {
  const roomId = getRoomId()
  if (!roomId) {
    return
  }

  fetch(`/rooms/${roomId}/mark-as-read`, {
    method: 'POST',
  }).catch((error) => {
    console.error('Error:', error)
  })
}
