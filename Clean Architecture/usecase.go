package main

import (
	"fmt"
	"strings"
)

type OrderRepository interface {
	List() ([]Order, error)
	GetByID(id string) (*Order, error)
	Create(order Order) error
	Update(order Order) error
	Delete(id string) error
}

type ListOrdersUseCase struct {
	Repo OrderRepository
}

func (uc *ListOrdersUseCase) Execute() ([]Order, error) {
	return uc.Repo.List()
}

type GetOrderUseCase struct {
	Repo OrderRepository
}

func (uc *GetOrderUseCase) Execute(id string) (*Order, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("order ID cannot be empty")
	}
	return uc.Repo.GetByID(id)
}

type CreateOrderUseCase struct {
	Repo OrderRepository
}

func (uc *CreateOrderUseCase) Execute(order Order) error {
	if strings.TrimSpace(order.ID) == "" {
		return fmt.Errorf("order ID cannot be empty")
	}
	if order.Value < 0 {
		return fmt.Errorf("order value cannot be negative")
	}
	return uc.Repo.Create(order)
}

type UpdateOrderUseCase struct {
	Repo OrderRepository
}

func (uc *UpdateOrderUseCase) Execute(order Order) error {
	if strings.TrimSpace(order.ID) == "" {
		return fmt.Errorf("order ID cannot be empty")
	}
	if order.Value < 0 {
		return fmt.Errorf("order value cannot be negative")
	}
	return uc.Repo.Update(order)
}

type DeleteOrderUseCase struct {
	Repo OrderRepository
}

func (uc *DeleteOrderUseCase) Execute(id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("order ID cannot be empty")
	}
	return uc.Repo.Delete(id)
}
