package handler

import (
	"errors"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/shangsky/zerodrop/model"
	"github.com/shangsky/zerodrop/pkg/randname"
)

var ErrConnNotExist = errors.New("ws client not exists")

type wsClient struct {
	user model.User
	conn *websocket.Conn
}

type wsClientRoom struct {
	m             sync.RWMutex
	id            string
	clients       map[string]*wsClient
	randNameStore *randname.Store
}

func newWSClientRoom(id string, randNameStore *randname.Store) *wsClientRoom {
	return &wsClientRoom{id: id, clients: make(map[string]*wsClient), randNameStore: randNameStore}
}

func (r *wsClientRoom) register(key string, c *wsClient) {
	r.m.Lock()
	defer r.m.Unlock()
	r.clients[key] = c
}

func (r *wsClientRoom) unRegister(key string) {
	r.m.Lock()
	defer r.m.Unlock()
	delete(r.clients, key)
}

func (r *wsClientRoom) Get(key string) (*wsClient, bool) {
	r.m.Lock()
	defer r.m.Unlock()
	c, ok := r.clients[key]
	return c, ok
}

func (r *wsClientRoom) sendTo(key string, msg any) error {
	r.m.RLock()
	defer r.m.RUnlock()
	client, ok := r.clients[key]
	if !ok {
		return ErrConnNotExist
	}
	return client.conn.WriteJSON(msg)
}

func (r *wsClientRoom) sendAll(msg any) error {
	r.m.RLock()
	defer r.m.RUnlock()
	var errs []error
	for _, client := range r.clients {
		err := client.conn.WriteJSON(msg)
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

func (r *wsClientRoom) users() []model.User {
	var users []model.User
	for _, client := range r.clients {
		users = append(users, client.user)
	}
	return users
}

func (r *wsClientRoom) isEmpty() bool {
	r.m.RLock()
	defer r.m.RUnlock()
	return len(r.clients) == 0
}

func filterUsers(users []model.User, id string) []model.User {
	var us []model.User
	for _, u := range users {
		if u.ID != id {
			us = append(us, u)
		}
	}
	return us
}

func (r *wsClientRoom) freshUsers() error {
	r.m.RLock()
	defer r.m.RUnlock()
	users := r.users()
	var errs []error
	for _, client := range r.clients {
		me := client.user
		others := filterUsers(users, client.user.ID)
		if err := client.conn.WriteJSON(
			model.RespOK(model.MethodFreshUsers, model.RegisterResp{Me: me, Users: others}),
		); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

type wsClientRooms struct {
	m     sync.RWMutex
	rooms map[string]*wsClientRoom
}

func newWSClientRooms() *wsClientRooms {
	return &wsClientRooms{rooms: make(map[string]*wsClientRoom)}
}

func (rms *wsClientRooms) Set(id string, room *wsClientRoom) {
	rms.m.Lock()
	defer rms.m.Unlock()
	rms.rooms[id] = room
}

func (rms *wsClientRooms) Get(id string) (*wsClientRoom, bool) {
	rms.m.RLock()
	defer rms.m.RUnlock()
	room, ok := rms.rooms[id]
	return room, ok
}

func (rms *wsClientRooms) Delete(id string) {
	rms.m.Lock()
	defer rms.m.Unlock()
	delete(rms.rooms, id)
}
