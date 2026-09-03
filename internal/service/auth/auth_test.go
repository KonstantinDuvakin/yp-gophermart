package auth

import "testing"

func TestHashAndCheckPassword(t *testing.T) {
	hash, err := HashPassword("s3cret-password")
	if err != nil {
		t.Fatalf("HashPassword: unexpected error %v", err)
	}
	if hash == "" || hash == "s3cret-password" {
		t.Fatalf("HashPassword returned suspicious hash %q", hash)
	}

	ok, err := CheckPasswordHash(hash, "s3cret-password")
	if err != nil || !ok {
		t.Errorf("CheckPasswordHash(correct) = (%v, %v), want (true, nil)", ok, err)
	}

	ok, err = CheckPasswordHash(hash, "wrong-password")
	if err != nil || ok {
		t.Errorf("CheckPasswordHash(wrong) = (%v, %v), want (false, nil)", ok, err)
	}
}

func TestCheckPasswordHash_InvalidHash(t *testing.T) {
	ok, err := CheckPasswordHash("not-a-bcrypt-hash", "whatever")
	if err == nil || ok {
		t.Errorf("CheckPasswordHash(invalid hash) = (%v, %v), want (false, error)", ok, err)
	}
}

func TestTokenManager_Roundtrip(t *testing.T) {
	tm := NewTokenManager("test-secret")

	token, err := tm.BuildJWTString(42)
	if err != nil {
		t.Fatalf("BuildJWTString: %v", err)
	}

	claims, err := tm.ParseJWTString(token)
	if err != nil {
		t.Fatalf("ParseJWTString: %v", err)
	}
	if claims.UserID != 42 {
		t.Errorf("UserId = %d, want 42", claims.UserID)
	}
}

func TestTokenManager_ParseInvalid(t *testing.T) {
	tm := NewTokenManager("test-secret")
	if _, err := tm.ParseJWTString("garbage.token.value"); err == nil {
		t.Error("ParseJWTString(garbage) expected error, got nil")
	}
}

func TestTokenManager_WrongSecret(t *testing.T) {
	token, err := NewTokenManager("secret-one").BuildJWTString(7)
	if err != nil {
		t.Fatalf("BuildJWTString: %v", err)
	}
	if _, err := NewTokenManager("secret-two").ParseJWTString(token); err == nil {
		t.Error("ParseJWTString with wrong secret expected error, got nil")
	}
}
