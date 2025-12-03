package http

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/romashah-741992/auto-sender/internal/messages"
	"github.com/romashah-741992/auto-sender/internal/scheduler"
)

type SchedulerRequest struct {
	Action string `json:"action"` // "start" or "stop"
}

type schedulerResponse struct {
	Status string `json:"status"`
}

type httpError struct {
	Error string `json:"error"`
}

func RegisterRoutes(mux *http.ServeMux, sched *scheduler.Scheduler, svc *messages.Service) {
	mux.HandleFunc("/scheduler", SchedulerHandler(sched))
	mux.HandleFunc("/messages/sent", ListSentMessagesHandler(svc))
}

// SchedulerHandler handles starting and stopping the background scheduler.
//
// @Summary      Start or stop automatic message sending
// @Description  Controls the background scheduler that sends up to 2 pending messages every 2 minutes.
// @Tags         scheduler
// @Accept       json
// @Produce      json
// @Param        body  body      SchedulerRequest  true  "Scheduler action (start|stop)"
// @Success      200   {object}  schedulerResponse
// @Failure      400   {object}  httpError
// @Failure      405   {object}  httpError
// @Router       /scheduler [post]
func SchedulerHandler(s *scheduler.Scheduler) http.HandlerFunc {
	// POST /scheduler { "action": "start" | "stop" }
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}

		var req SchedulerRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
			return
		}

		action := strings.ToLower(strings.TrimSpace(req.Action))
		switch action {
		case "start":
			s.Start()
			writeJSON(w, http.StatusOK, map[string]string{"status": "started"})
		case "stop":
			s.Stop()
			writeJSON(w, http.StatusOK, map[string]string{"status": "stopped"})
		default:
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "action must be 'start' or 'stop'"})
		}
	}
}

// ListSentMessagesHandler returns sent messages.
//
// @Summary      List sent messages
// @Description  Returns messages with status=sent, ordered by sent time.
// @Tags         messages
// @Produce      json
// @Param        limit   query    int  false  "Limit"  default(50)
// @Param        offset  query    int  false  "Offset" default(0)
// @Success      200     {array}  messages.Message
// @Failure      405     {object} httpError
// @Failure      500     {object} httpError
// @Router       /messages/sent [get]
func ListSentMessagesHandler(svc *messages.Service) http.HandlerFunc {
	// GET /messages/sent?limit=&offset=
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}

		q := r.URL.Query()
		limit, _ := strconv.Atoi(q.Get("limit"))
		offset, _ := strconv.Atoi(q.Get("offset"))
		if limit <= 0 {
			limit = 50
		}

		msgs, err := svc.ListSentMessages(context.Background(), limit, offset)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, msgs)
	}
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
