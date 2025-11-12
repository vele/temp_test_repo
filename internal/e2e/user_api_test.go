package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/vele/temp_test_repo/internal/domain"
	"github.com/vele/temp_test_repo/internal/event"
	"github.com/vele/temp_test_repo/internal/service"
	"github.com/vele/temp_test_repo/internal/testutil"
	httptransport "github.com/vele/temp_test_repo/internal/transport/http"
	"github.com/vele/temp_test_repo/internal/transport/http/handler"
	"github.com/vele/temp_test_repo/internal/transport/http/middleware"
)

func TestUserAPIEndToEnd(t *testing.T) {
	dsn := testutil.StartPostgres(t)
	repo := testutil.ConnectRepository(t, dsn)
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = repo.Truncate(ctx)
		require.NoError(t, repo.Close())
	})

	publisher := event.NewInMemoryPublisher()
	userSvc := service.NewUserService(repo, repo, publisher)

	userHandler := handler.NewUserHandler(userSvc)
	authHandler := handler.NewAuthHandler("secret", "admin", "password", time.Minute*15)
	authMW := middleware.NewAuth("secret")

	router := httptransport.NewRouter(httptransport.RouterDeps{
		UserHandler: userHandler,
		AuthHandler: authHandler,
		Auth:        authMW,
	})

	server := httptest.NewServer(router)
	t.Cleanup(server.Close)

	client := server.Client()
	baseURL := server.URL

	token := login(t, client, baseURL+"/auth/login")
	user := createUser(t, client, baseURL+"/api/v1/users", token)
	require.NotZero(t, user.ID)

	updated := updateUser(t, client, baseURL+"/api/v1/users", token, user.ID)
	require.Equal(t, 31, updated.Age)

	users := listUsers(t, client, baseURL+"/api/v1/users", token)
	require.Len(t, users, 1)

	f := addFile(t, client, baseURL+"/api/v1/users", token, user.ID)
	require.Equal(t, "passport", f.Name)

	files := listFiles(t, client, baseURL+"/api/v1/users", token, user.ID)
	require.Len(t, files, 1)

	deleteFiles(t, client, baseURL+"/api/v1/users", token, user.ID)
	files = listFiles(t, client, baseURL+"/api/v1/users", token, user.ID)
	require.Len(t, files, 0)

	deleteUser(t, client, baseURL+"/api/v1/users", token, user.ID)

	events := publisher.Events()
	require.Len(t, events, 3)
	require.Equal(t, event.UserCreated, events[0].Type)
	require.Equal(t, event.UserUpdated, events[1].Type)
	require.Equal(t, event.UserDeleted, events[2].Type)
}

func login(t *testing.T, client *http.Client, url string) string {
	resp := doRequest(t, client, http.MethodPost, url, "", map[string]string{
		"username": "admin",
		"password": "password",
	})
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var body map[string]string
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	return body["token"]
}

func createUser(t *testing.T, client *http.Client, base, token string) domain.User {
	resp := doRequest(t, client, http.MethodPost, base, token, service.CreateUserInput{
		Name:  "Jane",
		Email: "jane@example.com",
		Age:   30,
	})
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	var user domain.User
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&user))
	return user
}

func listUsers(t *testing.T, client *http.Client, base, token string) []domain.User {
	resp := doRequest(t, client, http.MethodGet, base, token, nil)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var users []domain.User
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&users))
	return users
}

func updateUser(t *testing.T, client *http.Client, base, token string, userID uint) domain.User {
	url := base + "/" + itoa(userID)
	newAge := 31
	resp := doRequest(t, client, http.MethodPut, url, token, service.UpdateUserInput{
		Age: &newAge,
	})
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var user domain.User
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&user))
	return user
}

func addFile(t *testing.T, client *http.Client, base, token string, userID uint) domain.File {
	url := base + "/" + itoa(userID) + "/files"
	resp := doRequest(t, client, http.MethodPost, url, token, service.FileInput{
		Name: "passport",
		Path: "/tmp/passport.pdf",
	})
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	var file domain.File
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&file))
	return file
}

func listFiles(t *testing.T, client *http.Client, base, token string, userID uint) []domain.File {
	url := base + "/" + itoa(userID) + "/files"
	resp := doRequest(t, client, http.MethodGet, url, token, nil)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var files []domain.File
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&files))
	return files
}

func deleteFiles(t *testing.T, client *http.Client, base, token string, userID uint) {
	url := base + "/" + itoa(userID) + "/files"
	resp := doRequest(t, client, http.MethodDelete, url, token, nil)
	defer resp.Body.Close()
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func deleteUser(t *testing.T, client *http.Client, base, token string, userID uint) {
	url := base + "/" + itoa(userID)
	resp := doRequest(t, client, http.MethodDelete, url, token, nil)
	defer resp.Body.Close()
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func doRequest(t *testing.T, client *http.Client, method, url, token string, payload interface{}) *http.Response {
	t.Helper()

	var body []byte
	var err error
	if payload != nil {
		body, err = json.Marshal(payload)
		require.NoError(t, err)
	}

	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := client.Do(req)
	require.NoError(t, err)
	return resp
}

func itoa(v uint) string {
	return strconv.FormatUint(uint64(v), 10)
}
