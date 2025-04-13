package integration_test

import (
	"net/http"
	"testing"

	. "github.com/Eun/go-hit"
	"github.com/google/uuid"

	"github.com/spanwalla/pvz/internal/entity"
)

// Scenario
// 1. POST /pvz
// 2. POST /receptions
// 3. POST /products x50
// 4. POST /pvz/:pvzId/close_last_reception
func TestReceptionScenario(t *testing.T) {
	const products = 50
	const city = "Казань"
	allowedProducts := []entity.ProductType{entity.ProductTypeShoes, entity.ProductTypeClothes, entity.ProductTypeElectronics}

	moderatorToken, err := dummyLogin(entity.RoleTypeModerator)
	if err != nil {
		t.Fatal(err)
	}

	employeeToken, err := dummyLogin(entity.RoleTypeEmployee)
	if err != nil {
		t.Fatal(err)
	}

	pvzId, err := createPvz(moderatorToken, city)
	if err != nil {
		t.Fatal(err)
	}

	err = openReception(employeeToken, pvzId)
	if err != nil {
		t.Fatal(err)
	}

	for i := range products {
		err = createProduct(employeeToken, pvzId, string(allowedProducts[i%len(allowedProducts)]))
		if err != nil {
			t.Fatal(err)
		}
	}

	err = closeReception(employeeToken, pvzId)
	if err != nil {
		t.Fatal(err)
	}
}

// POST /pvz
func createPvz(token string, city string) (uuid.UUID, error) {
	var id uuid.UUID

	body := map[string]string{
		"city": city,
	}
	if err := Do(
		Post(basePath+"/pvz"),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Headers("Authorization").Add("Bearer "+token),
		Send().Body().JSON(body),
		Expect().Status().Equal(http.StatusCreated),
		Store().Response().Body().JSON().JQ(".id").In(&id),
	); err != nil {
		return uuid.Nil, err
	}

	return id, nil
}

// POST /receptions
func openReception(token string, pvzId uuid.UUID) error {
	body := map[string]string{
		"pvzId": pvzId.String(),
	}
	if err := Do(
		Post(basePath+"/receptions"),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Headers("Authorization").Add("Bearer "+token),
		Send().Body().JSON(body),
		Expect().Status().Equal(http.StatusCreated),
	); err != nil {
		return err
	}

	return nil
}

// POST /products
func createProduct(token string, pvzId uuid.UUID, productType string) error {
	body := map[string]string{
		"pvzId": pvzId.String(),
		"type":  productType,
	}
	if err := Do(
		Post(basePath+"/products"),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Headers("Authorization").Add("Bearer "+token),
		Send().Body().JSON(body),
		Expect().Status().Equal(http.StatusCreated),
	); err != nil {
		return err
	}

	return nil
}

// POST /pvz/:pvzId/close_last_reception
func closeReception(token string, pvzId uuid.UUID) error {
	if err := Do(
		Post(basePath+"/pvz/"+pvzId.String()+"/close_last_reception"),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Headers("Authorization").Add("Bearer "+token),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".status").Equal("close"),
	); err != nil {
		return err
	}

	return nil
}
