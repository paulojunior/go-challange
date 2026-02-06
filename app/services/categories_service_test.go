package services

import (
	"context"
	"errors"
	"testing"

	"github.com/mytheresa/go-hiring-challenge/models"
)

// mockCategoryRepository is a mock implementation of CategoryRepository for testing.
type mockCategoryRepository struct {
	getAllCategoriesFunc func(ctx context.Context) ([]models.Category, error)
	createCategoryFunc   func(ctx context.Context, code, name string) (*models.Category, error)
}

func (m *mockCategoryRepository) GetAllCategories(ctx context.Context) ([]models.Category, error) {
	if m.getAllCategoriesFunc != nil {
		return m.getAllCategoriesFunc(ctx)
	}
	return nil, errors.New("not implemented")
}

func (m *mockCategoryRepository) CreateCategory(ctx context.Context, code, name string) (*models.Category, error) {
	if m.createCategoryFunc != nil {
		return m.createCategoryFunc(ctx, code, name)
	}
	return nil, errors.New("not implemented")
}

func TestListCategories_Success(t *testing.T) {
	mockRepo := &mockCategoryRepository{
		getAllCategoriesFunc: func(ctx context.Context) ([]models.Category, error) {
			return []models.Category{
				{ID: 1, Code: "CLOTHING", Name: "Clothing"},
				{ID: 2, Code: "SHOES", Name: "Shoes"},
				{ID: 3, Code: "ACCESSORIES", Name: "Accessories"},
			}, nil
		},
	}

	svc := NewCategoriesService(mockRepo)

	result, err := svc.ListCategories(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 3 {
		t.Fatalf("expected 3 categories, got %d", len(result))
	}

	if result[0].Code != "CLOTHING" {
		t.Errorf("expected first category code CLOTHING, got %s", result[0].Code)
	}
	if result[0].Name != "Clothing" {
		t.Errorf("expected first category name Clothing, got %s", result[0].Name)
	}

	if result[1].Code != "SHOES" {
		t.Errorf("expected second category code SHOES, got %s", result[1].Code)
	}

	if result[2].Code != "ACCESSORIES" {
		t.Errorf("expected third category code ACCESSORIES, got %s", result[2].Code)
	}
}

func TestListCategories_Empty(t *testing.T) {
	mockRepo := &mockCategoryRepository{
		getAllCategoriesFunc: func(ctx context.Context) ([]models.Category, error) {
			return []models.Category{}, nil
		},
	}

	svc := NewCategoriesService(mockRepo)

	result, err := svc.ListCategories(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("expected empty list, got %d categories", len(result))
	}
}

func TestListCategories_RepositoryError(t *testing.T) {
	mockRepo := &mockCategoryRepository{
		getAllCategoriesFunc: func(ctx context.Context) ([]models.Category, error) {
			return nil, errors.New("database error")
		},
	}

	svc := NewCategoriesService(mockRepo)

	_, err := svc.ListCategories(context.Background())

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCreateCategory_Success(t *testing.T) {
	mockRepo := &mockCategoryRepository{
		createCategoryFunc: func(ctx context.Context, code, name string) (*models.Category, error) {
			return &models.Category{
				ID:   4,
				Code: code,
				Name: name,
			}, nil
		},
	}

	svc := NewCategoriesService(mockRepo)
	input := CreateCategoryInput{
		Code: "ELECTRONICS",
		Name: "Electronics",
	}

	result, err := svc.CreateCategory(context.Background(), input)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Code != "ELECTRONICS" {
		t.Errorf("expected code ELECTRONICS, got %s", result.Code)
	}
	if result.Name != "Electronics" {
		t.Errorf("expected name Electronics, got %s", result.Name)
	}
}

func TestCreateCategory_EmptyCode(t *testing.T) {
	mockRepo := &mockCategoryRepository{}

	svc := NewCategoriesService(mockRepo)
	input := CreateCategoryInput{
		Code: "",
		Name: "Electronics",
	}

	_, err := svc.CreateCategory(context.Background(), input)

	if !errors.Is(err, ErrInvalidCategoryInput) {
		t.Errorf("expected ErrInvalidCategoryInput, got %v", err)
	}
}

func TestCreateCategory_EmptyName(t *testing.T) {
	mockRepo := &mockCategoryRepository{}

	svc := NewCategoriesService(mockRepo)
	input := CreateCategoryInput{
		Code: "ELECTRONICS",
		Name: "",
	}

	_, err := svc.CreateCategory(context.Background(), input)

	if !errors.Is(err, ErrInvalidCategoryInput) {
		t.Errorf("expected ErrInvalidCategoryInput, got %v", err)
	}
}

func TestCreateCategory_BothEmpty(t *testing.T) {
	mockRepo := &mockCategoryRepository{}

	svc := NewCategoriesService(mockRepo)
	input := CreateCategoryInput{
		Code: "",
		Name: "",
	}

	_, err := svc.CreateCategory(context.Background(), input)

	if !errors.Is(err, ErrInvalidCategoryInput) {
		t.Errorf("expected ErrInvalidCategoryInput, got %v", err)
	}
}

func TestCreateCategory_RepositoryError(t *testing.T) {
	mockRepo := &mockCategoryRepository{
		createCategoryFunc: func(ctx context.Context, code, name string) (*models.Category, error) {
			return nil, errors.New("duplicate key violation")
		},
	}

	svc := NewCategoriesService(mockRepo)
	input := CreateCategoryInput{
		Code: "ELECTRONICS",
		Name: "Electronics",
	}

	_, err := svc.CreateCategory(context.Background(), input)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCreateCategory_VerifiesInputPassedToRepo(t *testing.T) {
	var capturedCode, capturedName string

	mockRepo := &mockCategoryRepository{
		createCategoryFunc: func(ctx context.Context, code, name string) (*models.Category, error) {
			capturedCode = code
			capturedName = name
			return &models.Category{
				ID:   1,
				Code: code,
				Name: name,
			}, nil
		},
	}

	svc := NewCategoriesService(mockRepo)
	input := CreateCategoryInput{
		Code: "TEST_CODE",
		Name: "Test Name",
	}

	_, err := svc.CreateCategory(context.Background(), input)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedCode != "TEST_CODE" {
		t.Errorf("expected code TEST_CODE to be passed to repo, got %s", capturedCode)
	}
	if capturedName != "Test Name" {
		t.Errorf("expected name 'Test Name' to be passed to repo, got %s", capturedName)
	}
}
