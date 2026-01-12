package domain

import postgre_model "github.com/Vilamuzz/yota-backend/domain/model/postgre"

func GetAllModels() []interface{} {
	return []interface{}{
		&postgre_model.User{},
	}
}
