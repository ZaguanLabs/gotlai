package cache

import (
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
)

func TestRedisCache_Get_Hit(t *testing.T) {
	db, mock := redismock.NewClientMock()
	defer db.Close()

	cache := NewRedisCacheFromClient(db, 3600, "test:")

	mock.ExpectGet("test:mykey").SetVal("myvalue")

	val, ok := cache.Get("mykey")
	if !ok {
		t.Error("Expected cache hit")
	}
	if val != "myvalue" {
		t.Errorf("Expected 'myvalue', got %q", val)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet expectations: %v", err)
	}
}

func TestRedisCache_Get_Miss(t *testing.T) {
	db, mock := redismock.NewClientMock()
	defer db.Close()

	cache := NewRedisCacheFromClient(db, 3600, "test:")

	mock.ExpectGet("test:mykey").RedisNil()

	val, ok := cache.Get("mykey")
	if ok {
		t.Error("Expected cache miss")
	}
	if val != "" {
		t.Errorf("Expected empty string, got %q", val)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet expectations: %v", err)
	}
}

func TestRedisCache_Set(t *testing.T) {
	db, mock := redismock.NewClientMock()
	defer db.Close()

	cache := NewRedisCacheFromClient(db, 3600, "test:")

	mock.ExpectSet("test:mykey", "myvalue", 3600*time.Second).SetVal("OK")

	err := cache.Set("mykey", "myvalue")
	if err != nil {
		t.Errorf("Set failed: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet expectations: %v", err)
	}
}

func TestRedisCache_Set_NoTTL(t *testing.T) {
	db, mock := redismock.NewClientMock()
	defer db.Close()

	cache := NewRedisCacheFromClient(db, 0, "test:")

	mock.ExpectSet("test:mykey", "myvalue", 0).SetVal("OK")

	err := cache.Set("mykey", "myvalue")
	if err != nil {
		t.Errorf("Set failed: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet expectations: %v", err)
	}
}

func TestRedisCache_KeyPrefix(t *testing.T) {
	db, mock := redismock.NewClientMock()
	defer db.Close()

	cache := NewRedisCacheFromClient(db, 3600, "gotlai:v1:")

	// Verify prefix is applied
	mock.ExpectGet("gotlai:v1:hash123").SetVal("translated")

	val, ok := cache.Get("hash123")
	if !ok || val != "translated" {
		t.Errorf("Expected 'translated', got %q (ok=%v)", val, ok)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet expectations: %v", err)
	}
}

func TestRedisCache_Ping(t *testing.T) {
	db, mock := redismock.NewClientMock()
	defer db.Close()

	cache := NewRedisCacheFromClient(db, 3600, "test:")

	mock.ExpectPing().SetVal("PONG")

	err := cache.Ping()
	if err != nil {
		t.Errorf("Ping failed: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet expectations: %v", err)
	}
}

func TestRedisCache_Close(t *testing.T) {
	db, mock := redismock.NewClientMock()

	cache := NewRedisCacheFromClient(db, 3600, "test:")

	// Close should work without error
	err := cache.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}

	_ = mock // Silence unused warning
}
