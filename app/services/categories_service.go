package services

import (
	"context"

	"github.com/mytheresa/go-hiring-challenge/models"
)

// CreateCategoryInput represents the input for creating a category.
type CreateCategoryInput struct {
	Code string
	Name string
}

// CategoryRepository defines the interface for category data access.
type CategoryRepository interface {
	GetAllCategories(ctx context.Context) ([]models.Category, error)
	CreateCategory(ctx context.Context, code, name string) (*models.Category, error)
}

// CategoriesService handles category business logic.
type CategoriesService struct {
	repo CategoryRepository
}

// NewCategoriesService creates a new CategoriesService instance.
func NewCategoriesService(repo CategoryRepository) *CategoriesService {
	return &CategoriesService{repo: repo}
}

// ListCategories retrieves all categories.
func (s *CategoriesService) ListCategories(ctx context.Context) ([]CategoryDTO, error) {
	categories, err := s.repo.GetAllCategories(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]CategoryDTO, len(categories))
	for i, c := range categories {
		result[i] = CategoryDTO{
			Code: c.Code,
			Name: c.Name,
		}
	}

	return result, nil
}

// CreateCategory creates a new category after validating input.
func (s *CategoriesService) CreateCategory(ctx context.Context, input CreateCategoryInput) (*CategoryDTO, error) {
	if input.Code == "" || input.Name == "" {
		return nil, ErrInvalidCategoryInput
	}

	category, err := s.repo.CreateCategory(ctx, input.Code, input.Name)
	if err != nil {
		return nil, err
	}

	return &CategoryDTO{
		Code: category.Code,
		Name: category.Name,
	}, nil
}
