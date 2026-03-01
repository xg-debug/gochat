package service

type UserService struct {
}

func NewUserService() *UserService {
	return &UserService{}
}


func (s *UserService) Login(username, password string) (string, error) {
	return "", nil
}