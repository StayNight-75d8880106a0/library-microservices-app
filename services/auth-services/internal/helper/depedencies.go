package helper

import "auth-services/internal/repository"

type Depedencies struct {
	IsblacklistToken repository.AuthRepositoryInterface
}
