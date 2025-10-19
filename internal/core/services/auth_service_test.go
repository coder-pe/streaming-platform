// Copyright (c) 2024 Miguel Mamani
// Email: miguel.coder.per@gmail.com
// License: MIT

package services

import (
	"context"
	"testing"
	"time"

	"streaming-platform/internal/core/domain"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Mock del UserRepository
type mockUserRepository struct {
	users map[string]*domain.User // email -> User
}

func newMockUserRepository() *mockUserRepository {
	return &mockUserRepository{
		users: make(map[string]*domain.User),
	}
}

func (m *mockUserRepository) Create(ctx context.Context, user *domain.User) error {
	m.users[user.Email] = user
	return nil
}

func (m *mockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	for _, user := range m.users {
		if user.ID == id {
			return user, nil
		}
	}
	return nil, domain.ErrUserNotFound
}

func (m *mockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	user, exists := m.users[email]
	if !exists {
		return nil, domain.ErrUserNotFound
	}
	return user, nil
}

func (m *mockUserRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *mockUserRepository) UpdatePassword(ctx context.Context, id uuid.UUID, hashedPassword string) error {
	for _, user := range m.users {
		if user.ID == id {
			user.PasswordHash = hashedPassword
			return nil
		}
	}
	return domain.ErrUserNotFound
}

func (m *mockUserRepository) GetPublicProfile(ctx context.Context, id uuid.UUID) (*domain.UserProfile, error) {
	user, err := m.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &domain.UserProfile{
		ID:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      user.Role,
		Avatar:    user.Avatar,
		CreatedAt: user.CreatedAt,
	}, nil
}

func (m *mockUserRepository) UpdateProfile(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	return nil
}

func (m *mockUserRepository) GetUserStats(ctx context.Context, id uuid.UUID) (*domain.UserStats, error) {
	return nil, nil
}

func (m *mockUserRepository) GetSettings(ctx context.Context, id uuid.UUID) (map[string]interface{}, error) {
	return nil, nil
}

func (m *mockUserRepository) UpdateSettings(ctx context.Context, id uuid.UUID, settings map[string]interface{}) error {
	return nil
}

func (m *mockUserRepository) Search(ctx context.Context, query string, page, limit int) ([]*domain.User, int64, error) {
	return nil, 0, nil
}

func (m *mockUserRepository) UpdateStatus(ctx context.Context, id uuid.UUID, isActive bool) error {
	return nil
}

func (m *mockUserRepository) UpdateRole(ctx context.Context, id uuid.UUID, role string) error {
	return nil
}

func (m *mockUserRepository) Update(ctx context.Context, user *domain.User) error {
	m.users[user.Email] = user
	return nil
}

func (m *mockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	for email, user := range m.users {
		if user.ID == id {
			delete(m.users, email)
			return nil
		}
	}
	return domain.ErrUserNotFound
}

func (m *mockUserRepository) List(ctx context.Context, page, limit int) ([]*domain.User, int64, error) {
	users := make([]*domain.User, 0, len(m.users))
	for _, user := range m.users {
		users = append(users, user)
	}
	return users, int64(len(users)), nil
}

// Mock del CacheRepository
type mockCacheRepository struct {
	cache map[string]interface{}
}

func newMockCacheRepository() *mockCacheRepository {
	return &mockCacheRepository{
		cache: make(map[string]interface{}),
	}
}

func (m *mockCacheRepository) Set(ctx context.Context, key string, value interface{}, exp time.Duration) error {
	m.cache[key] = value
	return nil
}

func (m *mockCacheRepository) Get(ctx context.Context, key string, dest interface{}) error {
	value, exists := m.cache[key]
	if !exists {
		return domain.ErrCacheKeyNotFound
	}
	// En un mock simple, simplemente asignamos el valor
	// En producción, el cache hace JSON marshal/unmarshal
	*dest.(*interface{}) = value
	return nil
}

func (m *mockCacheRepository) Delete(ctx context.Context, key string) error {
	delete(m.cache, key)
	return nil
}

func (m *mockCacheRepository) Exists(ctx context.Context, key string) (bool, error) {
	_, exists := m.cache[key]
	return exists, nil
}

func (m *mockCacheRepository) SetTTL(ctx context.Context, key string, exp time.Duration) error {
	return nil
}

func (m *mockCacheRepository) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	return 0, nil
}

func (m *mockCacheRepository) DeleteByPattern(ctx context.Context, pattern string) error {
	return nil
}

func (m *mockCacheRepository) SetString(ctx context.Context, key string, value string, exp time.Duration) error {
	m.cache[key] = value
	return nil
}

func (m *mockCacheRepository) GetString(ctx context.Context, key string) (string, error) {
	value, exists := m.cache[key]
	if !exists {
		return "", domain.ErrCacheKeyNotFound
	}
	return value.(string), nil
}

func (m *mockCacheRepository) CacheUser(ctx context.Context, user *domain.User) error {
	m.cache["user:"+user.ID.String()] = user
	return nil
}

func (m *mockCacheRepository) GetCachedUser(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	value, exists := m.cache["user:"+userID.String()]
	if !exists {
		return nil, domain.ErrCacheKeyNotFound
	}
	return value.(*domain.User), nil
}

func (m *mockCacheRepository) InvalidateUserCache(ctx context.Context, userID uuid.UUID) error {
	delete(m.cache, "user:"+userID.String())
	return nil
}

func (m *mockCacheRepository) CacheVideo(ctx context.Context, video *domain.Video) error {
	m.cache["video:"+video.ID.String()] = video
	return nil
}

func (m *mockCacheRepository) GetCachedVideo(ctx context.Context, videoID uuid.UUID) (*domain.Video, error) {
	value, exists := m.cache["video:"+videoID.String()]
	if !exists {
		return nil, domain.ErrCacheKeyNotFound
	}
	return value.(*domain.Video), nil
}

func (m *mockCacheRepository) InvalidateVideoCache(ctx context.Context, videoID uuid.UUID) error {
	delete(m.cache, "video:"+videoID.String())
	return nil
}

func (m *mockCacheRepository) SetSession(ctx context.Context, sessionID string, userID uuid.UUID, exp time.Duration) error {
	m.cache["session:"+sessionID] = userID.String()
	return nil
}

func (m *mockCacheRepository) GetSession(ctx context.Context, sessionID string) (uuid.UUID, error) {
	value, exists := m.cache["session:"+sessionID]
	if !exists {
		return uuid.Nil, domain.ErrCacheKeyNotFound
	}
	return uuid.Parse(value.(string))
}

func (m *mockCacheRepository) InvalidateSession(ctx context.Context, sessionID string) error {
	delete(m.cache, "session:"+sessionID)
	return nil
}

func (m *mockCacheRepository) Ping(ctx context.Context) error {
	return nil
}

func (m *mockCacheRepository) Increment(ctx context.Context, key string) (int64, error) {
	return 0, nil
}

func (m *mockCacheRepository) IncrementBy(ctx context.Context, key string, value int64) (int64, error) {
	return 0, nil
}

func (m *mockCacheRepository) Decrement(ctx context.Context, key string) (int64, error) {
	return 0, nil
}

func (m *mockCacheRepository) DecrementBy(ctx context.Context, key string, value int64) (int64, error) {
	return 0, nil
}

func (m *mockCacheRepository) SetMultiple(ctx context.Context, items map[string]interface{}, exp time.Duration) error {
	for key, value := range items {
		m.cache[key] = value
	}
	return nil
}

func (m *mockCacheRepository) GetMultiple(ctx context.Context, keys []string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	for _, key := range keys {
		if value, exists := m.cache[key]; exists {
			result[key] = value
		}
	}
	return result, nil
}

func (m *mockCacheRepository) DeleteMultiple(ctx context.Context, keys []string) error {
	for _, key := range keys {
		delete(m.cache, key)
	}
	return nil
}

func (m *mockCacheRepository) GetKeys(ctx context.Context, pattern string) ([]string, error) {
	keys := make([]string, 0, len(m.cache))
	for key := range m.cache {
		keys = append(keys, key)
	}
	return keys, nil
}

func (m *mockCacheRepository) ListPush(ctx context.Context, key string, values ...interface{}) error {
	return nil
}

func (m *mockCacheRepository) ListPop(ctx context.Context, key string) (interface{}, error) {
	return nil, nil
}

func (m *mockCacheRepository) ListLength(ctx context.Context, key string) (int64, error) {
	return 0, nil
}

func (m *mockCacheRepository) SetAdd(ctx context.Context, key string, member string) error {
	return nil
}

func (m *mockCacheRepository) SetRemove(ctx context.Context, key string, member string) error {
	return nil
}

func (m *mockCacheRepository) SetMembers(ctx context.Context, key string) ([]interface{}, error) {
	return nil, nil
}

func (m *mockCacheRepository) SetExists(ctx context.Context, key string, member interface{}) (bool, error) {
	return false, nil
}

func (m *mockCacheRepository) HashSet(ctx context.Context, key string, field string, value interface{}) error {
	return nil
}

func (m *mockCacheRepository) HashGet(ctx context.Context, key string, field string) (interface{}, error) {
	return nil, nil
}

func (m *mockCacheRepository) HashGetAll(ctx context.Context, key string) (map[string]interface{}, error) {
	return nil, nil
}

func (m *mockCacheRepository) HashDelete(ctx context.Context, key string, fields ...string) error {
	return nil
}

// Tests

func TestAuthService_Register_Success(t *testing.T) {
	// Arrange (Preparar)
	mockUserRepo := newMockUserRepository()
	mockCacheRepo := newMockCacheRepository()
	authService := NewAuthService(mockUserRepo, mockCacheRepo, "test-secret-key")

	email := "test@example.com"
	password := "SecurePass123!"
	firstName := "Test"
	lastName := "User"
	role := "user"

	// Act (Actuar)
	accessToken, refreshToken, profile, err := authService.Register(context.Background(), email, password, firstName, lastName, role)

	// Assert (Verificar)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if accessToken == "" {
		t.Error("Expected access token to be generated")
	}

	if refreshToken == "" {
		t.Error("Expected refresh token to be generated")
	}

	if profile == nil {
		t.Fatal("Expected user profile to be returned")
	}

	if profile.Email != email {
		t.Errorf("Expected email %s, got %s", email, profile.Email)
	}

	if profile.FirstName != firstName {
		t.Errorf("Expected first name %s, got %s", firstName, profile.FirstName)
	}

	if profile.LastName != lastName {
		t.Errorf("Expected last name %s, got %s", lastName, profile.LastName)
	}

	// Verificar que el usuario fue guardado en el repositorio
	savedUser, err := mockUserRepo.GetByEmail(context.Background(), email)
	if err != nil {
		t.Errorf("Expected user to be saved in repository, got error: %v", err)
	}

	if savedUser.Email != email {
		t.Errorf("Expected saved user email %s, got %s", email, savedUser.Email)
	}
}

func TestAuthService_Register_InvalidEmail(t *testing.T) {
	// Arrange
	mockUserRepo := newMockUserRepository()
	mockCacheRepo := newMockCacheRepository()
	authService := NewAuthService(mockUserRepo, mockCacheRepo, "test-secret-key")

	invalidEmail := "not-an-email"
	password := "SecurePass123"
	firstName := "Test"
	lastName := "User"
	role := "user"

	// Act
	accessToken, refreshToken, profile, err := authService.Register(context.Background(), invalidEmail, password, firstName, lastName, role)

	// Assert
	if err == nil {
		t.Error("Expected error for invalid email, got none")
	}

	if accessToken != "" {
		t.Error("Expected no access token for invalid email")
	}

	if refreshToken != "" {
		t.Error("Expected no refresh token for invalid email")
	}

	if profile != nil {
		t.Error("Expected no profile for invalid email")
	}
}

func TestAuthService_Register_DuplicateEmail(t *testing.T) {
	// Arrange
	mockUserRepo := newMockUserRepository()
	mockCacheRepo := newMockCacheRepository()
	authService := NewAuthService(mockUserRepo, mockCacheRepo, "test-secret-key")

	email := "duplicate@example.com"
	password := "SecurePass123!"
	firstName := "User"
	lastName := "One"
	role := "user"

	// Crear el primer usuario
	_, _, _, err := authService.Register(context.Background(), email, password, firstName, lastName, role)
	if err != nil {
		t.Fatalf("Expected first registration to succeed, got error: %v", err)
	}

	// Intentar crear el segundo usuario con el mismo email
	_, _, _, err = authService.Register(context.Background(), email, password, "User", "Two", role)

	// Assert
	if err != domain.ErrUserAlreadyExists {
		t.Errorf("Expected ErrUserAlreadyExists, got %v", err)
	}
}

func TestAuthService_Login_Success(t *testing.T) {
	// Arrange
	mockUserRepo := newMockUserRepository()
	mockCacheRepo := newMockCacheRepository()
	authService := NewAuthService(mockUserRepo, mockCacheRepo, "test-secret-key")

	email := "login@example.com"
	password := "SecurePass123!"
	firstName := "Login"
	lastName := "User"
	role := "user"

	// Primero registrar un usuario
	_, _, _, err := authService.Register(context.Background(), email, password, firstName, lastName, role)
	if err != nil {
		t.Fatalf("Expected registration to succeed, got error: %v", err)
	}

	// Act - Intentar hacer login
	accessToken, refreshToken, profile, err := authService.Login(context.Background(), email, password)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if accessToken == "" {
		t.Error("Expected access token to be generated")
	}

	if refreshToken == "" {
		t.Error("Expected refresh token to be generated")
	}

	if profile == nil {
		t.Fatal("Expected user profile to be returned")
	}

	if profile.Email != email {
		t.Errorf("Expected email %s, got %s", email, profile.Email)
	}
}

func TestAuthService_Login_InvalidCredentials(t *testing.T) {
	// Arrange
	mockUserRepo := newMockUserRepository()
	mockCacheRepo := newMockCacheRepository()
	authService := NewAuthService(mockUserRepo, mockCacheRepo, "test-secret-key")

	email := "test@example.com"
	password := "SecurePass123!"
	wrongPassword := "WrongPassword!"
	firstName := "Test"
	lastName := "User"
	role := "user"

	// Registrar un usuario
	_, _, _, err := authService.Register(context.Background(), email, password, firstName, lastName, role)
	if err != nil {
		t.Fatalf("Expected registration to succeed, got error: %v", err)
	}

	// Act - Intentar login con password incorrecta
	accessToken, refreshToken, profile, err := authService.Login(context.Background(), email, wrongPassword)

	// Assert
	if err != domain.ErrInvalidCredentials {
		t.Errorf("Expected ErrInvalidCredentials, got %v", err)
	}

	if accessToken != "" {
		t.Error("Expected no access token for invalid credentials")
	}

	if refreshToken != "" {
		t.Error("Expected no refresh token for invalid credentials")
	}

	if profile != nil {
		t.Error("Expected no profile for invalid credentials")
	}
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	// Arrange
	mockUserRepo := newMockUserRepository()
	mockCacheRepo := newMockCacheRepository()
	authService := NewAuthService(mockUserRepo, mockCacheRepo, "test-secret-key")

	email := "nonexistent@example.com"
	password := "SomePassword123!"

	// Act - Intentar login con usuario que no existe
	accessToken, refreshToken, profile, err := authService.Login(context.Background(), email, password)

	// Assert
	if err != domain.ErrInvalidCredentials {
		t.Errorf("Expected ErrInvalidCredentials, got %v", err)
	}

	if accessToken != "" {
		t.Error("Expected no access token for non-existent user")
	}

	if refreshToken != "" {
		t.Error("Expected no refresh token for non-existent user")
	}

	if profile != nil {
		t.Error("Expected no profile for non-existent user")
	}
}

func TestAuthService_ValidateToken_Success(t *testing.T) {
	// Arrange
	mockUserRepo := newMockUserRepository()
	mockCacheRepo := newMockCacheRepository()
	jwtSecret := "test-secret-key-for-validation"
	authService := NewAuthService(mockUserRepo, mockCacheRepo, jwtSecret)

	// Crear un usuario
	userID := uuid.New()
	email := "validate@example.com"
	password := "Password123!"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	user := &domain.User{
		ID:           userID,
		Email:        email,
		PasswordHash: string(hashedPassword),
		FirstName:    "Validate",
		LastName:     "User",
		Role:         "user",
		IsActive:     true,
	}
	mockUserRepo.Create(context.Background(), user)

	// Generar un token válido
	accessToken, _, _, err := authService.Login(context.Background(), email, password)
	if err != nil {
		t.Fatalf("Expected login to succeed, got error: %v", err)
	}

	// Act - Validar el token
	tokenClaims, err := authService.ValidateToken(context.Background(), accessToken)

	// Assert
	if err != nil {
		t.Errorf("Expected no error validating token, got %v", err)
	}

	if tokenClaims == nil {
		t.Fatal("Expected token claims to be returned")
	}

	if tokenClaims.UserID != userID {
		t.Errorf("Expected user ID %s, got %s", userID, tokenClaims.UserID)
	}

	if tokenClaims.Email != email {
		t.Errorf("Expected email %s, got %s", email, tokenClaims.Email)
	}
}

func TestAuthService_ValidateToken_InvalidToken(t *testing.T) {
	// Arrange
	mockUserRepo := newMockUserRepository()
	mockCacheRepo := newMockCacheRepository()
	authService := NewAuthService(mockUserRepo, mockCacheRepo, "test-secret-key")

	invalidToken := "invalid.jwt.token"

	// Act
	tokenClaims, err := authService.ValidateToken(context.Background(), invalidToken)

	// Assert
	if err == nil {
		t.Error("Expected error for invalid token, got none")
	}

	if tokenClaims != nil {
		t.Error("Expected nil token claims for invalid token")
	}
}
