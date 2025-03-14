package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/mileusna/useragent"
	"github.com/shangsky/zerodrop/model"
	"github.com/shangsky/zerodrop/pkg/randname"
)

func toDevice(os, device, name string) string {
	val := strings.Join([]string{os, device, name}, " ")
	if val == "" {
		return "unknown device"
	}
	return val
}

func (h *handler) handleWS(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	roomID := q.Get("room_id")
	if roomID == "" {
		http.Error(w, "room id is empty", 400)
		return
	}
	c, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("ws upgrade error", "error", err)
		return
	}
	defer c.Close()

	h.lock.Lock()
	room, ok := h.rooms.Get(roomID)
	if !ok {
		room = newWSClientRoom(roomID, randname.New(names))
		h.rooms.Set(roomID, room)
	}
	h.lock.Unlock()

	id := strings.ReplaceAll(uuid.New().String(), "-", "")
	ua := useragent.Parse(r.UserAgent())
	name := room.randNameStore.Pop()
	client := &wsClient{conn: c, user: model.User{ID: id, Name: name, IsMobile: ua.Mobile, Device: toDevice(ua.OS, ua.Device, ua.Name)}}
	h.serveWS(client, room)
	room.randNameStore.Put(name)
	h.lock.Lock()
	if room.isEmpty() {
		h.logger.Info("clear empty room", "id", roomID)
		h.rooms.Delete(roomID)
	}
	h.lock.Unlock()
}

func (h *handler) serveWS(c *wsClient, room *wsClientRoom) {
	h.logger.Info("user join room", "user", c.user, "room", room.id)
	room.register(c.user.ID, c)
	if err := room.freshUsers(); err != nil {
		h.logger.Error("fresh users error", "error", err)
	}

	for {
		var req model.Req
		if err := c.conn.ReadJSON(&req); err != nil {
			h.logger.Error("read json error", "error", err)
			break
		}

		switch req.Method {
		case model.MethodSendMsg:
			raw, _ := req.Data.MarshalJSON()
			var data model.SendMsg
			if err := json.Unmarshal(raw, &data); err != nil {
				h.logger.Error("parse json error", "error", err)
				continue
			}
			h.handleSendMsg(c, room, data)
		case model.MethodReceiveFileREFUSED, model.MethodSendFileCancel:
			raw, _ := req.Data.MarshalJSON()
			var data model.TransferFileCancel
			if err := json.Unmarshal(raw, &data); err != nil {
				h.logger.Error("parse json error", "error", err)
				continue
			}
			h.handleTransferFileCancel(req.Method, data)
		}
	}

	h.logger.Info("unregister one user", "user", c.user)
	room.unRegister(c.user.ID)
	if err := room.freshUsers(); err != nil {
		h.logger.Error("fresh users error", "error", err)
	}
}

func (h *handler) handleSendMsg(c *wsClient, room *wsClientRoom, data model.SendMsg) {
	h.logger.Info("send msg", "from", c.user.ID, "to", data.To, "msg", data.Msg)
	if err := room.sendTo(data.To, model.RespOK(model.MethodReceiveMsg, model.ReceiveMsg{From: c.user.ID, Msg: data.Msg})); err != nil {
		h.logger.Error("send msg error", "error", err)
	}
}

func (h *handler) handleTransferFileCancel(method string, data model.TransferFileCancel) {
	h.logger.Info("transfer file canceld", "download_id", data.DownloadID)
	downloader, ok := h.downloaderStore.Get(data.DownloadID)
	if !ok {
		h.logger.Error("downloader not exists")
		return
	}

	select {
	case downloader.Cancel <- method:
	default:
	}
}
