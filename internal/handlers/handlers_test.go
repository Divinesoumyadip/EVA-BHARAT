package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"ticket-system/internal/handlers"
	"ticket-system/internal/repository"
)

func newTestServer() *httptest.Server {
	store := repository.NewMemoryStore()
	secret := []byte("test-secret")
	authHandler := handlers.NewAuthHandler(store, secret, time.Hour)
	ticketHandler := handlers.NewTicketHandler(store)
	router := handlers.NewRouter(store, secret, authHandler, ticketHandler)
	return httptest.NewServer(router)
}

func doRequest(t *testing.T, method, url, token string, body interface{}) (*http.Response, map[string]interface{}) {
	t.Helper()

	var reqBody *bytes.Buffer
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("failed to marshal body: %v", err)
		}
		reqBody = bytes.NewBuffer(data)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	var parsed map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&parsed)
	return resp, parsed
}

func registerAndLogin(t *testing.T, server *httptest.Server, email, password string) string {
	t.Helper()

	resp, _ := doRequest(t, http.MethodPost, server.URL+"/auth/register", "", map[string]string{
		"email":    email,
		"password": password,
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 on register, got %d", resp.StatusCode)
	}

	resp, body := doRequest(t, http.MethodPost, server.URL+"/auth/login", "", map[string]string{
		"email":    email,
		"password": password,
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on login, got %d", resp.StatusCode)
	}

	token, ok := body["token"].(string)
	if !ok || token == "" {
		t.Fatalf("expected token in login response, got %v", body)
	}
	return token
}

func TestHealth(t *testing.T) {
	server := newTestServer()
	defer server.Close()

	resp, body := doRequest(t, http.MethodGet, server.URL+"/health", "", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if body["status"] != "ok" {
		t.Fatalf("expected status ok, got %v", body)
	}
}

func TestRegisterDuplicate(t *testing.T) {
	server := newTestServer()
	defer server.Close()

	doRequest(t, http.MethodPost, server.URL+"/auth/register", "", map[string]string{
		"email":    "dup@example.com",
		"password": "password123",
	})
	resp, _ := doRequest(t, http.MethodPost, server.URL+"/auth/register", "", map[string]string{
		"email":    "dup@example.com",
		"password": "password123",
	})
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("expected 409, got %d", resp.StatusCode)
	}
}

func TestRegisterMissingFields(t *testing.T) {
	server := newTestServer()
	defer server.Close()

	resp, _ := doRequest(t, http.MethodPost, server.URL+"/auth/register", "", map[string]string{
		"email": "nopass@example.com",
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestRegisterDoesNotReturnPassword(t *testing.T) {
	server := newTestServer()
	defer server.Close()

	_, body := doRequest(t, http.MethodPost, server.URL+"/auth/register", "", map[string]string{
		"email":    "safe@example.com",
		"password": "password123",
	})
	if _, exists := body["password"]; exists {
		t.Fatal("response should not contain password")
	}
	if _, exists := body["password_hash"]; exists {
		t.Fatal("response should not contain password_hash")
	}
}

func TestLoginWrongPassword(t *testing.T) {
	server := newTestServer()
	defer server.Close()

	doRequest(t, http.MethodPost, server.URL+"/auth/register", "", map[string]string{
		"email":    "wrongpass@example.com",
		"password": "password123",
	})
	resp, _ := doRequest(t, http.MethodPost, server.URL+"/auth/login", "", map[string]string{
		"email":    "wrongpass@example.com",
		"password": "incorrect",
	})
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestLoginUnknownUser(t *testing.T) {
	server := newTestServer()
	defer server.Close()

	resp, _ := doRequest(t, http.MethodPost, server.URL+"/auth/login", "", map[string]string{
		"email":    "ghost@example.com",
		"password": "password123",
	})
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestLoginMalformedRequest(t *testing.T) {
	server := newTestServer()
	defer server.Close()

	req, _ := http.NewRequest(http.MethodPost, server.URL+"/auth/login", bytes.NewBufferString("{invalid json"))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestCreateTicketRequiresAuth(t *testing.T) {
	server := newTestServer()
	defer server.Close()

	resp, _ := doRequest(t, http.MethodPost, server.URL+"/tickets", "", map[string]string{
		"title": "no auth",
	})
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestCreateTicketMalformedAuthHeader(t *testing.T) {
	server := newTestServer()
	defer server.Close()

	req, _ := http.NewRequest(http.MethodGet, server.URL+"/tickets", nil)
	req.Header.Set("Authorization", "NotBearer sometoken")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestCreateTicketInvalidToken(t *testing.T) {
	server := newTestServer()
	defer server.Close()

	resp, _ := doRequest(t, http.MethodGet, server.URL+"/tickets", "invalid.token.value", nil)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestCreateTicketStartsOpen(t *testing.T) {
	server := newTestServer()
	defer server.Close()

	token := registerAndLogin(t, server, "creator@example.com", "password123")
	resp, body := doRequest(t, http.MethodPost, server.URL+"/tickets", token, map[string]string{
		"title":       "First ticket",
		"description": "details",
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
	if body["status"] != "open" {
		t.Fatalf("expected status open, got %v", body["status"])
	}
}

func TestCreateTicketInvalidInput(t *testing.T) {
	server := newTestServer()
	defer server.Close()

	token := registerAndLogin(t, server, "invalidinput@example.com", "password123")
	resp, _ := doRequest(t, http.MethodPost, server.URL+"/tickets", token, map[string]string{
		"title": "",
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestListTicketsOnlyOwnedByUser(t *testing.T) {
	server := newTestServer()
	defer server.Close()

	tokenA := registerAndLogin(t, server, "usera-list@example.com", "password123")
	tokenB := registerAndLogin(t, server, "userb-list@example.com", "password123")

	doRequest(t, http.MethodPost, server.URL+"/tickets", tokenA, map[string]string{"title": "A ticket"})
	doRequest(t, http.MethodPost, server.URL+"/tickets", tokenB, map[string]string{"title": "B ticket"})

	req, _ := http.NewRequest(http.MethodGet, server.URL+"/tickets", nil)
	req.Header.Set("Authorization", "Bearer "+tokenA)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	var tickets []map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&tickets)

	if len(tickets) != 1 {
		t.Fatalf("expected 1 ticket for user A, got %d", len(tickets))
	}
	if tickets[0]["title"] != "A ticket" {
		t.Fatalf("expected user A's own ticket, got %v", tickets[0]["title"])
	}
}

func TestOwnershipProtection(t *testing.T) {
	server := newTestServer()
	defer server.Close()

	tokenA := registerAndLogin(t, server, "usera-own@example.com", "password123")
	tokenB := registerAndLogin(t, server, "userb-own@example.com", "password123")

	_, created := doRequest(t, http.MethodPost, server.URL+"/tickets", tokenA, map[string]string{"title": "A's ticket"})
	ticketID := int64(created["id"].(float64))
	idStr := jsonNumberToString(ticketID)

	resp, _ := doRequest(t, http.MethodGet, server.URL+"/tickets/"+idStr, tokenB, nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 when user B reads user A's ticket, got %d", resp.StatusCode)
	}

	resp, _ = doRequest(t, http.MethodPatch, server.URL+"/tickets/"+idStr+"/status", tokenB, map[string]string{"status": "in_progress"})
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 when user B modifies user A's ticket, got %d", resp.StatusCode)
	}

	resp, _ = doRequest(t, http.MethodGet, server.URL+"/tickets/"+idStr, tokenA, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 when owner reads own ticket, got %d", resp.StatusCode)
	}
}

func TestGetTicketNotFound(t *testing.T) {
	server := newTestServer()
	defer server.Close()

	token := registerAndLogin(t, server, "notfound@example.com", "password123")
	resp, _ := doRequest(t, http.MethodGet, server.URL+"/tickets/999999", token, nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestGetTicketMalformedID(t *testing.T) {
	server := newTestServer()
	defer server.Close()

	token := registerAndLogin(t, server, "malformed@example.com", "password123")
	resp, _ := doRequest(t, http.MethodGet, server.URL+"/tickets/not-a-number", token, nil)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestStatusTransitionFlow(t *testing.T) {
	server := newTestServer()
	defer server.Close()

	token := registerAndLogin(t, server, "flow@example.com", "password123")
	_, created := doRequest(t, http.MethodPost, server.URL+"/tickets", token, map[string]string{"title": "Flow ticket"})
	idStr := jsonNumberToString(int64(created["id"].(float64)))

	resp, body := doRequest(t, http.MethodPatch, server.URL+"/tickets/"+idStr+"/status", token, map[string]string{"status": "in_progress"})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for open -> in_progress, got %d", resp.StatusCode)
	}
	if body["status"] != "in_progress" {
		t.Fatalf("expected in_progress, got %v", body["status"])
	}

	resp, body = doRequest(t, http.MethodPatch, server.URL+"/tickets/"+idStr+"/status", token, map[string]string{"status": "closed"})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for in_progress -> closed, got %d", resp.StatusCode)
	}
	if body["status"] != "closed" {
		t.Fatalf("expected closed, got %v", body["status"])
	}

	resp, _ = doRequest(t, http.MethodPatch, server.URL+"/tickets/"+idStr+"/status", token, map[string]string{"status": "open"})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for closed -> open, got %d", resp.StatusCode)
	}
}

func TestStatusTransitionRejectsOpenToClosed(t *testing.T) {
	server := newTestServer()
	defer server.Close()

	token := registerAndLogin(t, server, "skip@example.com", "password123")
	_, created := doRequest(t, http.MethodPost, server.URL+"/tickets", token, map[string]string{"title": "Skip ticket"})
	idStr := jsonNumberToString(int64(created["id"].(float64)))

	resp, _ := doRequest(t, http.MethodPatch, server.URL+"/tickets/"+idStr+"/status", token, map[string]string{"status": "closed"})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for open -> closed, got %d", resp.StatusCode)
	}
}

func TestStatusTransitionInvalidValue(t *testing.T) {
	server := newTestServer()
	defer server.Close()

	token := registerAndLogin(t, server, "invalidstatus@example.com", "password123")
	_, created := doRequest(t, http.MethodPost, server.URL+"/tickets", token, map[string]string{"title": "Bad status ticket"})
	idStr := jsonNumberToString(int64(created["id"].(float64)))

	resp, _ := doRequest(t, http.MethodPatch, server.URL+"/tickets/"+idStr+"/status", token, map[string]string{"status": "cancelled"})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid status, got %d", resp.StatusCode)
	}
}

func jsonNumberToString(n int64) string {
	return strconv.FormatInt(n, 10)
}
