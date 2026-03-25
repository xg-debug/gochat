package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"gochat/internal/model"
	"gochat/internal/ws"
)

type chatPayload struct {
	Content     string `json:"content"`
	ContentType string `json:"contentType"`
	TempID      string `json:"tempId"`
}

type ackPayload struct {
	TempID    string `json:"tempId"`
	MessageID int64  `json:"messageId"`
	Error     string `json:"error"`
}

func main() {
	var (
		wsURL       = flag.String("ws-url", "ws://127.0.0.1:8080/ws", "websocket endpoint")
		token       = flag.String("token", "", "jwt token")
		mysqlDSN    = flag.String("mysql-dsn", "", "mysql dsn, e.g. user:pass@tcp(127.0.0.1:3306)/gochat?charset=utf8mb4&parseTime=True&loc=Local")
		toID        = flag.Uint64("to-id", 0, "target user id (single) or group id (group)")
		chatMode    = flag.String("chat-mode", "single", "single or group")
		tempID      = flag.String("temp-id", "idem-stress-1", "fixed client temp id for dedupe")
		content     = flag.String("content", "idempotency stress message", "message content")
		repeat      = flag.Int("repeat", 100, "number of resend attempts with same tempId")
		concurrency = flag.Int("concurrency", 20, "parallel send workers")
		waitAckSec  = flag.Int("wait-ack-sec", 5, "seconds to wait for acks")
		fromID      = flag.Int64("from-id", 0, "sender user id (optional, for DB verification filter)")
	)
	flag.Parse()

	if strings.TrimSpace(*token) == "" || strings.TrimSpace(*mysqlDSN) == "" || *toID == 0 {
		log.Fatal("token, mysql-dsn, to-id are required")
	}
	if *repeat <= 0 {
		*repeat = 1
	}
	if *concurrency <= 0 {
		*concurrency = 1
	}
	if *chatMode != "single" && *chatMode != "group" {
		log.Fatal("chat-mode must be single or group")
	}

	db, err := gorm.Open(mysql.Open(*mysqlDSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("open db failed: %v", err)
	}

	targetURL, err := appendToken(*wsURL, *token)
	if err != nil {
		log.Fatalf("invalid ws-url: %v", err)
	}
	conn, _, err := websocket.DefaultDialer.Dial(targetURL, nil)
	if err != nil {
		log.Fatalf("connect websocket failed: %v", err)
	}
	defer conn.Close()

	ackIDs := make(map[int64]int)
	ackErrs := make(map[string]int)
	var ackMu sync.Mutex
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			var msg ws.WSMessage
			if err := conn.ReadJSON(&msg); err != nil {
				return
			}
			if msg.Type != "ack" {
				continue
			}
			var payload ackPayload
			if err := json.Unmarshal(msg.Payload, &payload); err != nil {
				continue
			}
			if payload.TempID != *tempID {
				continue
			}
			ackMu.Lock()
			if payload.MessageID > 0 {
				ackIDs[payload.MessageID]++
			}
			if payload.Error != "" {
				ackErrs[payload.Error]++
			}
			ackMu.Unlock()
		}
	}()

	payloadRaw, _ := json.Marshal(chatPayload{
		Content:     *content,
		ContentType: "text",
		TempID:      *tempID,
	})

	jobs := make(chan struct{}, *repeat)
	for i := 0; i < *repeat; i++ {
		jobs <- struct{}{}
	}
	close(jobs)

	var wg sync.WaitGroup
	var writeMu sync.Mutex
	for i := 0; i < *concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range jobs {
				out := ws.WSMessage{
					Type:    *chatMode,
					FromID:  0,
					ToID:    *toID,
					Payload: payloadRaw,
				}
				writeMu.Lock()
				err := conn.WriteJSON(out)
				writeMu.Unlock()
				if err != nil {
					ackMu.Lock()
					ackErrs["write_failed"]++
					ackMu.Unlock()
				}
			}
		}()
	}
	wg.Wait()

	time.Sleep(time.Duration(*waitAckSec) * time.Second)
	_ = conn.Close()
	<-done

	var messages []model.Message
	query := db.Where("to_id = ? AND chat_type = ? AND client_msg_id = ?", int64(*toID), modeToChatType(*chatMode), *tempID)
	if *fromID > 0 {
		query = query.Where("from_id = ?", *fromID)
	}
	if err := query.Find(&messages).Error; err != nil {
		log.Fatalf("query messages failed: %v", err)
	}

	ackMu.Lock()
	defer ackMu.Unlock()

	ackIDList := make([]int64, 0, len(ackIDs))
	for id := range ackIDs {
		ackIDList = append(ackIDList, id)
	}
	sort.Slice(ackIDList, func(i, j int) bool { return ackIDList[i] < ackIDList[j] })

	fmt.Println("===== Idempotency Stress Report =====")
	fmt.Printf("mode=%s toID=%d tempId=%s\n", *chatMode, *toID, *tempID)
	fmt.Printf("sent_attempts=%d concurrency=%d\n", *repeat, *concurrency)
	fmt.Printf("ack_unique_message_ids=%v\n", ackIDList)
	fmt.Printf("ack_message_id_hit_count=%v\n", ackIDs)
	fmt.Printf("ack_errors=%v\n", ackErrs)
	fmt.Printf("db_rows_with_same_client_msg_id=%d\n", len(messages))
	if len(messages) > 0 {
		ids := make([]int64, 0, len(messages))
		for _, m := range messages {
			ids = append(ids, m.ID)
		}
		sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
		fmt.Printf("db_message_ids=%v\n", ids)
	}

	if len(messages) != 1 {
		fmt.Fprintln(os.Stderr, "FAIL: expected exactly 1 DB row for same client_msg_id")
		os.Exit(1)
	}
	if len(ackIDList) != 1 {
		fmt.Fprintln(os.Stderr, "FAIL: expected exactly 1 unique ack messageId")
		os.Exit(1)
	}
	fmt.Println("PASS: idempotency verified (single logical message persisted)")
}

func appendToken(rawURL, token string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	q := u.Query()
	q.Set("token", token)
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func modeToChatType(mode string) int8 {
	if mode == "group" {
		return 2
	}
	return 1
}
