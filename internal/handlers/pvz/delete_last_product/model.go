package delete_last_product

import "github.com/go-playground/validator/v10"

type DeleteProductRequest struct {
	PvzID string `param:"pvzId" validate:"required,uuid"`
}

func validateDeleteProductRequest(req DeleteProductRequest) error {
	v := validator.New()
	return v.Struct(req)
}
