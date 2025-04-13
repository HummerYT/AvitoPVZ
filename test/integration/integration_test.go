package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"AvitoPVZ/internal/config"
	"AvitoPVZ/internal/handlers/dummy_login"
	"AvitoPVZ/internal/handlers/login"
	"AvitoPVZ/internal/handlers/products"
	"AvitoPVZ/internal/handlers/pvz/close_last_reception"
	deleteLastProduct "AvitoPVZ/internal/handlers/pvz/delete_last_product"
	pvzGet "AvitoPVZ/internal/handlers/pvz/get"
	pvzPost "AvitoPVZ/internal/handlers/pvz/post"
	"AvitoPVZ/internal/handlers/receptions"
	"AvitoPVZ/internal/handlers/register"
	"AvitoPVZ/internal/middleware/jwt"
	"AvitoPVZ/internal/models"
	authPool "AvitoPVZ/internal/repository/auth"
	productsRepository "AvitoPVZ/internal/repository/products"
	pvzRepository "AvitoPVZ/internal/repository/pvz"
	receptionsRepository "AvitoPVZ/internal/repository/receptions"
	loginUseCase "AvitoPVZ/internal/usecase/login"
	productsUseCase "AvitoPVZ/internal/usecase/products"
	pvzUseCase "AvitoPVZ/internal/usecase/pvz"
	receptionsUseCase "AvitoPVZ/internal/usecase/receptions"
	registerUseCase "AvitoPVZ/internal/usecase/register"

	"github.com/gofiber/fiber/v2"
)

func dummyLogin(app *fiber.App, role string) (string, error) {
	reqBody, _ := json.Marshal(map[string]string{"role": role})
	req := httptest.NewRequest("POST", "/dummyLogin", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("dummyLogin returned status %d: %s", resp.StatusCode, string(body))
	}

	var token string
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return "", err
	}

	return token, nil
}

func createPVZ(app *fiber.App, token, city string) (*models.PVZ, error) {
	reqBody, _ := json.Marshal(map[string]string{"city": city})
	req := httptest.NewRequest("POST", "/pvz", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req, -1)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("createPVZ status %d: %s", resp.StatusCode, string(body))
	}
	var pvz models.PVZ
	if err := json.NewDecoder(resp.Body).Decode(&pvz); err != nil {
		return nil, err
	}
	return &pvz, nil
}

func createReception(app *fiber.App, token, pvzID string) (*models.Reception, error) {
	reqBody, _ := json.Marshal(map[string]string{"pvzId": pvzID})
	req := httptest.NewRequest("POST", "/receptions", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req, -1)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("createReception status %d: %s", resp.StatusCode, string(body))
	}
	var rec models.Reception
	if err := json.NewDecoder(resp.Body).Decode(&rec); err != nil {
		return nil, err
	}
	return &rec, nil
}

func addProduct(app *fiber.App, token, pvzID, productType string) (*models.Product, error) {
	reqBody, _ := json.Marshal(map[string]string{
		"pvzId": pvzID,
		"type":  productType,
	})
	req := httptest.NewRequest("POST", "/products", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req, -1)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("addProduct status %d: %s", resp.StatusCode, string(body))
	}
	var prod models.Product
	if err := json.NewDecoder(resp.Body).Decode(&prod); err != nil {
		return nil, err
	}
	return &prod, nil
}

func closeReception(app *fiber.App, token, pvzID string) (*models.Reception, error) {
	url := "/pvz/" + pvzID + "/close_last_reception"
	req := httptest.NewRequest("POST", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req, -1)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("closeReception status %d: %s", resp.StatusCode, string(body))
	}
	var rec models.Reception
	if err := json.NewDecoder(resp.Body).Decode(&rec); err != nil {
		return nil, err
	}
	return &rec, nil
}

func TestIntegrationFullFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode.")
	}
	app := fiber.New()

	pathToConfig := "../../config_prod.yml"
	cfg := config.MustConfig(&pathToConfig)

	sourceURL := "file://../../internal/migrations/up"
	if err := cfg.Postgres.MigrationsUp(sourceURL); err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	pool := config.NewPostgres(ctx, cfg.Postgres)
	defer pool.Close()

	// repository group
	registerPool := authPool.NewInsertRepo(pool)
	pvzRepo := pvzRepository.NewPVZRepositoryPostgres(pool)
	receptionsRepo := receptionsRepository.NewReceptionRepositoryPg(pool)
	productsRepo := productsRepository.NewProductRepositoryPg(pool)

	// usecase group
	registerUC := registerUseCase.NewUseCase(registerPool)
	loginUC := loginUseCase.NewUseCase(registerPool)
	pvzUC := pvzUseCase.NewPVZUseCase(pvzRepo)
	receptionsUC := receptionsUseCase.NewReceptionUseCase(receptionsRepo)
	productsUC := productsUseCase.NewProductUseCase(productsRepo)

	// handlers group
	registerHandler := register.NewHandler(registerUC)
	loginHandler := login.NewHandler(loginUC)
	pvzCreateHandler := pvzPost.NewCreatePVZHandler(pvzUC)
	receptionsHandler := receptions.NewReceptionHandler(receptionsUC)
	productsHandler := products.NewProductHandler(productsUC)
	deleteLastProductHandler := deleteLastProduct.NewProductHandler(productsUC)
	closeLastReceptionHandler := close_last_reception.NewReceptionHandler(receptionsUC)
	pvzGetHandler := pvzGet.NewPVZDataHandler(pvzUC)

	// middleware group
	jwtToken := jwt.NewMiddleware(cfg.JWT.Secret)

	app.Post("/dummyLogin", dummy_login.DummyLoginHandler, jwtToken.SignedToken)
	app.Post("/register", registerHandler.Register)
	app.Post("/login", loginHandler.Register, jwtToken.SignedToken)

	app.Post("/pvz", jwtToken.CompareToken, pvzCreateHandler.Handle)
	app.Get("/pvz", jwtToken.CompareToken, pvzGetHandler.GetPVZData)
	app.Post("/pvz/:pvzId/close_last_reception", jwtToken.CompareToken, closeLastReceptionHandler.CloseLastReception)
	app.Post("/pvz/:pvzId/delete_last_product", jwtToken.CompareToken, deleteLastProductHandler.DeleteLastProduct)

	app.Post("/receptions", jwtToken.CompareToken, receptionsHandler.CreateReception)

	app.Post("/products", jwtToken.CompareToken, productsHandler.CreateProduct)

	moderatorToken, err := dummyLogin(app, "moderator")
	if err != nil {
		t.Fatalf("Не удалось получить токен модератора: %v", err)
	}

	pvz, err := createPVZ(app, moderatorToken, "Москва")
	if err != nil {
		t.Fatalf("Не удалось создать ПВЗ: %v", err)
	}
	t.Logf("Создан ПВЗ с ID: %s", pvz.ID)

	employeeToken, err := dummyLogin(app, "employee")
	if err != nil {
		t.Fatalf("Не удалось получить токен сотрудника: %v", err)
	}

	reception, err := createReception(app, employeeToken, pvz.ID)
	if err != nil {
		t.Fatalf("Не удалось создать приёмку: %v", err)
	}
	t.Logf("Создана приёмка с ID: %s", reception.ID)

	for i := 1; i <= 50; i++ {
		prod, err := addProduct(app, employeeToken, pvz.ID, "электроника")
		if err != nil {
			t.Fatalf("Ошибка при добавлении товара #%d: %v", i, err)
		}
		t.Logf("Добавлен товар #%d с ID: %s", i, prod.ID)
	}

	closedReception, err := closeReception(app, employeeToken, pvz.ID)
	if err != nil {
		t.Fatalf("Не удалось закрыть приёмку: %v", err)
	}
	if closedReception.Status != "close" {
		t.Errorf("Приёмка не закрыта, статус: %s", closedReception.Status)
	}
	t.Logf("Приёмка закрыта, статус: %s", closedReception.Status)
}
