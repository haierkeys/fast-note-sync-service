package main

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/lxzan/gws"
)

const (
	baseURL = "http://127.0.0.1:9000"
	wsURL   = "ws://127.0.0.1:9000/api/user/sync"
)

type Handler struct {
	gws.BuiltinEventHandler
	recv chan []byte
}

func (h *Handler) OnMessage(socket *gws.Conn, message *gws.Message) {
	defer message.Close()
	data := message.Data.Bytes()
	// Copy data because message is closed after return
	buf := make([]byte, len(data))
	copy(buf, data)
	h.recv <- buf
}

func main() {
	// 1. Register & Login
	token := loginOrRegister()
	fmt.Println("Got token:", token)

	// 2. Connect WS
	handler := &Handler{recv: make(chan []byte, 10)}
	u, _ := url.Parse(wsURL)
	socket, _, err := gws.NewClient(handler, &gws.ClientOption{
		Addr: u.String(),
	})
	if err != nil {
		log.Fatal("dial:", err)
	}
	go socket.ReadLoop()
	defer socket.WriteClose(1000, []byte("bye"))

	// 3. Auth
	sendJSON(socket, "Authorization", []byte(token))

	// Read response
	readResponse(handler) // Expect Authorization success

	// 4. Init Upload
	filename := "test.bin"
	size := int64(1024 * 1024) // 1MB
	data := make([]byte, size)
	data[0] = 1
	data[size-1] = 2

	hash := md5.Sum(data)
	hashStr := hex.EncodeToString(hash[:])

	initData := map[string]interface{}{
		"filename":    filename,
		"hash":        hashStr,
		"size":        size,
		"totalChunks": 1,
	}
	initBytes, _ := json.Marshal(initData)
	sendJSON(socket, "FileChunkUploadInit", initBytes)

	// Read response
	resp := readResponse(handler)
	sessionID := extractsessionID(resp)
	fmt.Println("Upload ID:", sessionID)

	// 5. Send Binary
	// Protocol: [sessionID (36 bytes)][ChunkIndex (4 bytes BigEndian)][Data...]
	buf := new(bytes.Buffer)
	buf.WriteString(sessionID)                     // 36 bytes
	binary.Write(buf, binary.BigEndian, uint32(0)) // Index 0
	buf.Write(data)                                // Data

	err = socket.WriteMessage(gws.OpcodeBinary, buf.Bytes())
	if err != nil {
		log.Fatal("write binary:", err)
	}
	fmt.Println("Sent binary chunk")

	// 6. Complete
	completeData := map[string]interface{}{
		"sessionID": sessionID,
	}
	completeBytes, _ := json.Marshal(completeData)
	sendJSON(socket, "FileChunkUploadComplete", completeBytes)

	// Read response
	finalResp := readResponse(handler)
	fmt.Println("Final response:", finalResp)
}

func loginOrRegister() string {
	client := &http.Client{Timeout: 5 * time.Second}

	// Generate random username
	randName := fmt.Sprintf("user%d", time.Now().UnixNano())
	regBody := []byte(fmt.Sprintf(`{"username":"%s","password":"password123","confirmPassword":"password123","email":"%s@example.com","nickname":"Tester"}`, randName, randName))
	resp, err := client.Post(baseURL+"/api/user/register", "application/json", bytes.NewBuffer(regBody))
	if err != nil {
		log.Fatal("register req:", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	fmt.Println("Register response:", string(body))

	if resp.StatusCode != 200 {
		fmt.Println("Register failed:", string(body))
	}

	var res struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	json.Unmarshal(body, &res)
	if res.Data.Token != "" {
		return res.Data.Token
	}

	// Login (fallback)
	loginBody := []byte(fmt.Sprintf(`{"credentials":"%s","password":"password123"}`, randName))
	resp, err = client.Post(baseURL+"/api/user/login", "application/json", bytes.NewBuffer(loginBody))
	if err != nil {
		log.Fatal("login req:", err)
	}
	defer resp.Body.Close()

	body, _ = io.ReadAll(resp.Body)
	var loginRes struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	json.Unmarshal(body, &loginRes)
	if loginRes.Data.Token == "" {
		fmt.Println("Login failed response:", string(body))
	}
	return loginRes.Data.Token
}

func sendJSON(socket *gws.Conn, typeStr string, data []byte) {
	payload := fmt.Sprintf("%s|%s", typeStr, string(data))
	socket.WriteMessage(gws.OpcodeText, []byte(payload))
}

func readResponse(h *Handler) string {
	select {
	case msg := <-h.recv:
		fmt.Println("Recv:", string(msg))
		return string(msg)
	case <-time.After(5 * time.Second):
		log.Fatal("timeout waiting for response")
		return ""
	}
}

func extractsessionID(resp string) string {
	// resp format: Action|JSON
	// e.g. FileChunkUploadInit|{"code":200,"data":{"sessionID":"...","chunkSize":...}}
	// But wait, `ToResponse` does:
	// `c.send(actionType, ResResult{...})`
	// `send` does: `fmt.Sprintf("%s|%s", actionType, string(responseBytes))`

	// However, `ToResponse` logic:
	// if actionType != "" ...
	// `c.send(actionType, ...)`

	// In `ws_upload.go`:
	// `c.ToResponse(code.Success.WithData(response), "FileChunkUploadInit")`
	// So it sends "FileChunkUploadInit|JSON"

	// But wait, `ToResponse` implementation in `pkg/app/websocket.go`:
	// `if global.Config.App.IsReturnSussess || actionType != "" || code.Code() > 200 || code.HaveData() { ... }`

	// So yes, it sends "FileChunkUploadInit|JSON".

	// But wait, `ResResult` struct:
	// type ResResult struct {
	// 	Code   int         `json:"code"`
	// 	Status bool        `json:"status"`
	// 	Msg    string      `json:"msg"`
	// 	Data   interface{} `json:"data"`
	// }

	// So JSON is `{"code":..., "data":...}`.

	// Wait, `strings.SplitN(resp, "|", 2)` might fail if `resp` doesn't have `|` (e.g. error without action type).
	// But `ToResponse` usually sends action type if provided.

	parts := bytes.SplitN([]byte(resp), []byte("|"), 2)
	var jsonPart []byte
	if len(parts) < 2 {
		jsonPart = []byte(resp)
	} else {
		jsonPart = parts[1]
	}

	var res struct {
		Data struct {
			sessionID string `json:"sessionID"`
		} `json:"data"`
	}
	json.Unmarshal(jsonPart, &res)
	return res.Data.sessionID
}
