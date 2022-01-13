package repository

type Repository interface {
	Save(data interface{}, collection string) error
	FindByUrl(url string, collection string) (int, error)
}
