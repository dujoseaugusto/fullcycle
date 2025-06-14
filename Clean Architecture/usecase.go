package main

type OrderRepository interface {
	List() ([]Order, error)
}

type ListOrdersUseCase struct {
	Repo OrderRepository
}

func (uc *ListOrdersUseCase) Execute() ([]Order, error) {
	return uc.Repo.List()
}
