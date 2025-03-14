package handler

import (
	"log/slog"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/shangsky/zerodrop/frontend"
	"github.com/shangsky/zerodrop/pkg/downloader"
	"github.com/shangsky/zerodrop/pkg/httpwrap"
)

type handler struct {
	upgrader        *websocket.Upgrader
	logger          *slog.Logger
	rooms           *wsClientRooms
	downloaderStore *downloader.DownloaderStore
	lock            sync.Mutex
}

func newHandler(logger *slog.Logger) *handler {
	upgrader := &websocket.Upgrader{}
	// upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	rooms := newWSClientRooms()
	downloaderStore := downloader.New()
	return &handler{upgrader: upgrader, logger: logger, rooms: rooms, downloaderStore: downloaderStore}
}

func New(logger *slog.Logger) http.Handler {
	handler := newHandler(logger)
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", handler.handleWS)
	mux.HandleFunc("/upload", httpwrap.POST(handler.upload))
	mux.HandleFunc("/download", httpwrap.Methods([]string{http.MethodHead, http.MethodGet}, handler.download))
	mux.Handle("/", http.FileServer(http.FS(frontend.StaticFS)))
	return mux
}
