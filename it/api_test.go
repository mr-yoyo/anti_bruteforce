package main

import (
	"database/sql"
	"log"
	"net/http"
	"testing"

	_ "github.com/lib/pq"
	"gopkg.in/h2non/baloo.v3"
)

var test = baloo.New("http://app_test:8080")

type TestSuite struct {
	db *sql.DB
}

var s TestSuite

func (s *TestSuite) SetupSuite() *TestSuite {
	db, err := sql.Open("postgres", "postgres://user:password@db_test/db?sslmode=disable")
	if err != nil {
		log.Fatalf("couldn't connect to db: %s", err)
	}

	s.db = db

	return s
}

func (s *TestSuite) TearDown() {
	_, _ = s.db.Exec(`TRUNCATE ip_blacklist`)
	_, _ = s.db.Exec(`TRUNCATE ip_whitelist`)
}

func TestAuth(t *testing.T) {
	tests := []struct {
		name   string
		body   map[string]string
		wantOk bool
		times  int
	}{
		{"same login ok", map[string]string{"login": "l", "ip": "192.168.0.1", "password": "p1"}, true, 10},
		{"same login not ok", map[string]string{"login": "l", "ip": "192.168.0.1", "password": "p1"}, false, 3},
		{"same password ok", map[string]string{"login": "l2", "ip": "192.168.0.2", "password": "p2"}, true, 10},
		{"same password ok", map[string]string{"login": "l3", "ip": "192.168.0.2", "password": "p2"}, true, 10},
		{"same password not ok", map[string]string{"login": "l4", "ip": "192.168.0.2", "password": "p2"}, false, 3},
		{"same ip ok", map[string]string{"login": "l5", "ip": "192.168.0.3", "password": "p3"}, true, 10},
		{"same ip ok", map[string]string{"login": "l6", "ip": "192.168.0.3", "password": "p4"}, true, 10},
		{"same ip ok", map[string]string{"login": "l7", "ip": "192.168.0.3", "password": "p5"}, true, 10},
		{"same ip not ok", map[string]string{"login": "l5", "ip": "192.168.0.3", "password": "p3"}, false, 3},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for i := 0; i < tc.times; i++ {
				_ = test.Post("/auth").
					JSON(tc.body).
					Expect(t).
					Status(http.StatusOK).
					Type("json").
					JSON(map[string]bool{
						"ok": tc.wantOk,
					}).Done()
			}
		})
	}
}

func TestAddBlacklist(t *testing.T) {
	tests := []struct {
		ipNet      string
		statusCode int
	}{
		{"192.168.0.1/16", http.StatusOK},
		{"192.168.0.1/32", http.StatusConflict},
		{"192.169.0.1/32", http.StatusOK},
		{"192.169.0.1", http.StatusBadRequest},
	}

	for _, tc := range tests {
		_ = test.Post("/blacklist").
			JSON(map[string]string{
				"subnet": tc.ipNet,
			}).
			Expect(t).
			Status(tc.statusCode).Done()
	}

	s.SetupSuite().TearDown()
}

func TestAddWhitelist(t *testing.T) {
	tests := []struct {
		ipNet      string
		statusCode int
	}{
		{"192.168.0.1/16", http.StatusOK},
		{"192.168.0.1/32", http.StatusConflict},
		{"192.169.0.1/32", http.StatusOK},
		{"192.169.0.1", http.StatusBadRequest},
	}

	for _, tc := range tests {
		_ = test.Post("/whitelist").
			JSON(map[string]string{
				"subnet": tc.ipNet,
			}).
			Expect(t).
			Status(tc.statusCode).Done()
	}

	s.SetupSuite().TearDown()
}

func TestBlackWhitelist(t *testing.T) {
	_, _ = test.Post("/blacklist").
		JSON(map[string]string{
			"subnet": "192.168.0.1/16",
		}).Send()

	_, _ = test.Post("/whitelist").
		JSON(map[string]string{
			"subnet": "192.168.0.1/16",
		}).Send()

	_, _ = test.Post("/whitelist").
		JSON(map[string]string{
			"subnet": "1.1.1.1/16",
		}).Send()

	tests := []struct {
		IPNet  string
		wantOk bool
		times  int
	}{
		{"192.168.0.1", false, 1},
		{"192.168.0.128", false, 1},
		{"192.168.1.1", false, 1},
		{"192.168.1.255", false, 1},
		{"1.1.1.1", true, 50},
	}

	for _, tc := range tests {
		t.Run("", func(t *testing.T) {
			_ = test.Post("/auth").
				JSON(map[string]string{
					"login":    "login",
					"password": "password",
					"ip":       tc.IPNet,
				}).
				Expect(t).
				Status(http.StatusOK).
				Type("json").
				JSON(map[string]bool{
					"ok": tc.wantOk,
				}).Done()
		})
	}

	s.SetupSuite().TearDown()
}

func TestDeleteBlacklist(t *testing.T) {
	_, _ = test.Post("/blacklist").
		JSON(map[string]string{
			"subnet": "192.168.0.1/16",
		}).Send()

	tests := []struct {
		ipNet      string
		statusCode int
	}{
		{"192.168.0.1/16", http.StatusOK},
		{"192.168.0.2/32", http.StatusNoContent},
		{"192.168.0.256/32", http.StatusBadRequest},
	}

	for _, tc := range tests {
		_ = test.Delete("/blacklist").
			JSON(map[string]string{
				"subnet": tc.ipNet,
			}).
			Expect(t).
			Status(tc.statusCode).Done()
	}
}

func TestDeleteWhitelist(t *testing.T) {
	_, _ = test.Post("/whitelist").
		JSON(map[string]string{
			"subnet": "192.168.0.1/16",
		}).Send()

	tests := []struct {
		ipNet      string
		statusCode int
	}{
		{"192.168.0.1/16", http.StatusOK},
		{"192.168.0.2/32", http.StatusNoContent},
		{"192.168.0.256/32", http.StatusBadRequest},
	}

	for _, tc := range tests {
		_ = test.Delete("/whitelist").
			JSON(map[string]string{
				"subnet": tc.ipNet,
			}).
			Expect(t).
			Status(tc.statusCode).Done()
	}
}

func TestResetBucket(t *testing.T) {
	_, _ = test.Post("/auth").
		JSON(map[string]string{
			"login":    "login",
			"password": "password",
			"ip":       "192.168.0.1",
		}).Send()

	tests := []struct {
		login      string
		IP         string
		statusCode int
	}{
		{"login", "192.168.0.2", http.StatusOK},
		{"login", "192.168.0.1", http.StatusOK},
		{"login", "192.168.0.1", http.StatusNoContent},
		{"login", "192.168.0.2", http.StatusNoContent},
		{"", "192.168.0.256", http.StatusBadRequest},
	}

	for _, tc := range tests {
		_ = test.Delete("/bucket").
			JSON(map[string]string{
				"login": tc.login,
				"ip":    tc.IP,
			}).
			Expect(t).
			Status(tc.statusCode).
			Done()
	}

	s.SetupSuite().TearDown()
}
