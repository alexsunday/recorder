package web

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"recorder/proto"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var (
	logger   = slog.New(slog.NewTextHandler(os.Stderr, nil))
	upgrader = websocket.Upgrader{
		ReadBufferSize:  4096,
		WriteBufferSize: 4096,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

type wsReadWrapper struct {
	*websocket.Conn
	// 读取缓存 由于websocket消息按帧发送 可能应用层只需要1字节，但实际上收到了100字节
	// 需要处理这种场景 所以将多余收到的数据缓存起来
	readCache []byte
}

func NewWsConn(c *websocket.Conn) io.ReadWriteCloser {
	return &wsReadWrapper{
		Conn: c,
	}
}

func (m *wsReadWrapper) Write(p []byte) (int, error) {
	err := m.Conn.WriteMessage(websocket.BinaryMessage, p)
	return len(p), err
}

func (m *wsReadWrapper) Close() error {
	return m.Conn.Close()
}

// 应用层获取，只从缓存中拿 如缓存不够 就从ws获取
func (m *wsReadWrapper) Read(p []byte) (int, error) {
	for {
		if len(m.readCache) >= len(p) {
			break
		}

		err := m.readFromLink()
		if err != nil {
			return 0, fmt.Errorf("read from link failed %w", err)
		}
	}

	return m.readFromCache(p)
}

func (m *wsReadWrapper) readFromCache(p []byte) (int, error) {
	appLength := len(p)
	if len(m.readCache) < appLength {
		panic("cache data is not enough")
	}
	// copy 时，长度以 参数中更小的那个 为准
	copy(p, m.readCache)
	m.readCache = m.readCache[appLength:]
	return appLength, nil
}

func (m *wsReadWrapper) readFromLink() error {
	msgType, buf, err := m.Conn.ReadMessage()
	if err != nil {
		return fmt.Errorf("read websocket message failed %w", err)
	}
	if msgType == websocket.TextMessage {
		return fmt.Errorf("unsupport text message now")
	}
	if msgType == websocket.CloseMessage {
		logger.Warn("received a valid close message")
		return m.Conn.Close()
	}
	if msgType == websocket.PingMessage {
		m.Conn.WriteControl(websocket.PongMessage, nil, time.Now().Add(5*time.Second))
		return nil
	}

	m.readCache = append(m.readCache, buf...)
	return nil
}

func WebInit(addr string) error {
	r := gin.Default()
	r.GET("/websocket/link", wsHttpHandler())

	return r.Run(addr)
}

func wsHttpHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		wsConn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			logger.Warn("upgrade websocket request failed", "error", err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		conn := proto.NewConnection(c, NewWsConn(wsConn))
		go conn.ReadLoop()
		go conn.Handle()
	}
}

/*
1. 使用邮箱注册
2. 使用邮箱 发送重置密码验证码
3. 使用邮箱 重置密码
4. 设备登录 提供设备指纹，返回一个设备 nanoid，用于设备侧协议登录时使用，设备侧应保存该ID；
5. 开新会话 返回会话ID 会话主要用于合并音频
*/
