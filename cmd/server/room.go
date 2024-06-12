package server

import "sync"

type rooms struct {
	rooms map[int64]map[int64]*LiveChatSocketMiddleware

	mutex sync.RWMutex
}

func newRooms() *rooms {
	return &rooms{
		rooms: make(map[int64]map[int64]*LiveChatSocketMiddleware),
	}
}

func (r *rooms) joinRoom(roomID int64, conn *LiveChatSocketMiddleware) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// initialize new room if not existed before
	if _, ok := r.rooms[roomID]; !ok {
		r.rooms[roomID] = map[int64]*LiveChatSocketMiddleware{}
	}

	if _, ok := r.rooms[roomID][conn.UserID]; !ok {
		r.rooms[roomID][conn.UserID] = conn
	}

}

func (r *rooms) leaveRoom(roomID int64, conn *LiveChatSocketMiddleware) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if room, ok := r.rooms[roomID]; ok {
		delete(room, conn.UserID) // remove connection from room

		if len(room) == 0 { // if room is empty, also remove room from pool
			delete(r.rooms, roomID)
		}
	}
}

func (r *rooms) getRoom(roomID int64) map[int64]*LiveChatSocketMiddleware {
	return r.rooms[roomID]
}
