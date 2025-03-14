package handler

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/shangsky/zerodrop/model"
	"github.com/shangsky/zerodrop/pkg/downloader"
)

func (h *handler) upload(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	from := q.Get("from")
	to := q.Get("to")
	filename := q.Get("filename")
	roomID := q.Get("room_id")
	downloadID := q.Get("download_id")
	if from == "" || to == "" || filename == "" || roomID == "" || downloadID == "" {
		http.Error(w, "from/to/filename/room_id is empty", http.StatusBadRequest)
		return
	}
	room, ok := h.rooms.Get(roomID)
	if !ok {
		http.Error(w, "房间不存在", http.StatusNotFound)
		return
	}
	_, ok = room.Get(from)
	if !ok {
		http.Error(w, "当前用户不存在", http.StatusNotFound)
		return
	}
	_, ok = room.Get(to)
	if !ok {
		http.Error(w, "目的用户不存在", http.StatusNotFound)
		return
	}

	logger := h.logger.With("from", from, "to", to, "filename", filename)
	defer r.Body.Close()

	var cache []byte
	if r.ContentLength > 1024*1024*10 {
		cachebuf := make([]byte, 1024*1024)
		cacheLength, err := io.ReadFull(r.Body, cachebuf)
		if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) && !errors.Is(err, io.EOF) {
			logger.Error("read file error", "error", err)
			http.Error(w, "读取文件出错", http.StatusBadRequest)
			return
		}
		cache = cachebuf[:cacheLength]
	}

	writetC := make(chan io.Writer, 1)
	stoped := make(chan struct{})
	cancel := make(chan string)
	defer close(stoped)

	downloader := downloader.Downloader{
		WriterC:    writetC,
		Stoped:     stoped,
		Cancel:     cancel,
		FileLength: r.ContentLength,
		Filename:   filename,
		Cache:      cache,
	}
	h.downloaderStore.Set(downloadID, downloader)
	defer h.downloaderStore.Delete(downloadID)
	if err := room.sendTo(to, model.RespOK(
		model.MethodReceiveFileConfirm,
		model.ReceiveFileConfirm{
			From:       from,
			DownloadID: downloadID,
			Filename:   filename,
			FileSize:   r.ContentLength,
		},
	)); err != nil {
		logger.Error("通知接收文件失败", "error", err)
		http.Error(w, "通知接收文件失败", http.StatusInternalServerError)
		return
	}
	timeout := time.NewTimer(time.Second * 90)
	logger = logger.With("download_id", downloadID)
	logger.Info("info", "fileSize", r.ContentLength)
	select {
	case reason := <-cancel:
		if reason == model.MethodReceiveFileREFUSED {
			http.Error(w, "发送文件被拒绝", http.StatusNotAcceptable)
			return
		}
		http.Error(w, "取消传输", http.StatusBadRequest)
		return
	case <-r.Context().Done():
		logger.Error("request cancel")
		return
	case <-timeout.C:
		logger.Error("wait timeout")
		http.Error(w, "等待传输超时", http.StatusRequestTimeout)
		return
	case writer := <-writetC:
		timeout.Stop()
		if err := room.sendTo(
			from,
			model.RespOK(
				model.MethodSendFileStart,
				model.SendFileStartResp{DownloadID: downloadID},
			),
		); err != nil {
			logger.Error("通知发送文件失败", "error", err)
			http.Error(w, "通知发送文件失败", http.StatusInternalServerError)
			return
		}
		if _, err := io.Copy(writer, r.Body); err != nil {
			logger.Error("copy file error", "error", err)
			http.Error(w, "文件传输失败", http.StatusConflict)
			return
		}
		w.WriteHeader(http.StatusOK)
		logger.Info("upload success")
	}
}

func (h *handler) download(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Range") != "" {
		http.Error(w, "not support range", http.StatusPreconditionFailed)
		return
	}
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "id is empty", http.StatusBadRequest)
		return
	}
	logger := h.logger.With("id", id)
	logger.Info("download file", "method", r.Method, "header", r.Header)
	downloader, ok := h.downloaderStore.Get(id)
	if !ok {
		http.Error(w, "传输文件不存在或者已过期", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", strconv.FormatInt(downloader.FileLength, 10))
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", downloader.Filename))
	if r.Method == http.MethodHead {
		w.WriteHeader(http.StatusOK)
		return
	}

	if len(downloader.Cache) > 0 {
		if _, err := io.Copy(w, bytes.NewReader(downloader.Cache)); err != nil {
			logger.Error("传输文件失败", "error", err)
			return
		}
	}

	h.downloaderStore.Delete(id)
	downloader.WriterC <- w
	<-downloader.Stoped
	logger.Info("download end", "filename", downloader.Filename)
}

//xhr abort问题
