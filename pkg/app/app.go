package app

import (
    "flag"
    "log"
    "net/http"

    "github.com/gorilla/websocket"
    "github.com/20TB-ZipBomb/GGJ_Platform/internal/logger"
)

var addr = flag.String("addr", "localhost:4040", "http service address")
var upgrader = websocket.Upgrader{}

func Echo(w http.ResponseWriter, r *http.Request) {
    c, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        logger.Error("upgrade: %s", err)
        return
    }
    defer c.Close()

    for {
        logger.Debug("waiting...")
        mt, message, err := c.ReadMessage()
        if err != nil {
            logger.Error("read: %s", err)
            break
        }
        log.Printf("recv: %s", message)
        err = c.WriteMessage(mt, message)
        if err != nil {
            logger.Error("write: %s", err)
            break
        }
    }
}

func Run() {
    logger.Init()
    defer logger.Sync()

    http.HandleFunc("/", Echo)
    log.Fatal(http.ListenAndServe(*addr, nil))
}

