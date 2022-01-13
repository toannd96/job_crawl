package repository

type Repository interface {
	Save(data interface{}, collection string) error
}
