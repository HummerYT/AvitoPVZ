package main

import (
	"AvitoPVZ/internal/handlers/pvz/close_last_reception"
	"context"
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"

	"AvitoPVZ/internal/config"
	"AvitoPVZ/internal/handlers/dummy_login"
	"AvitoPVZ/internal/handlers/login"
	"AvitoPVZ/internal/handlers/products"
	deleteLastProduct "AvitoPVZ/internal/handlers/pvz/delete_last_product"
	pvzGet "AvitoPVZ/internal/handlers/pvz/get"
	pvzPost "AvitoPVZ/internal/handlers/pvz/post"
	"AvitoPVZ/internal/handlers/receptions"
	"AvitoPVZ/internal/handlers/register"
	"AvitoPVZ/internal/middleware/jwt"
	authPool "AvitoPVZ/internal/repository/auth"
	productsRepository "AvitoPVZ/internal/repository/products"
	pvzRepository "AvitoPVZ/internal/repository/pvz"
	receptionsRepository "AvitoPVZ/internal/repository/receptions"
	loginUseCase "AvitoPVZ/internal/usecase/login"
	productsUseCase "AvitoPVZ/internal/usecase/products"
	pvzUseCase "AvitoPVZ/internal/usecase/pvz"
	receptionsUseCase "AvitoPVZ/internal/usecase/receptions"
	registerUseCase "AvitoPVZ/internal/usecase/register"
)

func main() {
	ctx := context.Background()
	app := fiber.New()
	app.Use(logger.New())
	app.Use(
		cors.New(
			cors.Config{
				Next:             nil,
				AllowOriginsFunc: nil,
				AllowOrigins:     "*",
				AllowMethods: strings.Join([]string{
					fiber.MethodGet,
					fiber.MethodPost,
					fiber.MethodHead,
					fiber.MethodPut,
					fiber.MethodDelete,
					fiber.MethodPatch,
				}, ","),
				AllowCredentials: false,
				MaxAge:           0,
				AllowHeaders:     "Authorization, Reset",
				ExposeHeaders:    "Authorization, Reset",
			},
		),
	)

	cfg := config.MustConfig(nil)

	if err := cfg.Postgres.MigrationsUp(); err != nil {
		panic(err)
	}

	pool := config.NewPostgres(ctx, cfg.Postgres)
	defer pool.Close()

	registerPool := authPool.NewInsertRepo(pool)
	pvzRepo := pvzRepository.NewPVZRepositoryPostgres(pool)
	receptionsRepo := receptionsRepository.NewReceptionRepositoryPg(pool)
	productsRepo := productsRepository.NewProductRepositoryPg(pool)

	registerUC := registerUseCase.NewUseCase(registerPool)
	loginUC := loginUseCase.NewUseCase(registerPool)
	pvzUC := pvzUseCase.NewPVZUseCase(pvzRepo)
	receptionsUC := receptionsUseCase.NewReceptionUseCase(receptionsRepo)
	productsUC := productsUseCase.NewProductUseCase(productsRepo)

	registerHandler := register.NewHandler(registerUC)
	loginHandler := login.NewHandler(loginUC)
	pvzCreateHandler := pvzPost.NewCreatePVZHandler(pvzUC)
	receptionsHandler := receptions.NewReceptionHandler(receptionsUC)
	productsHandler := products.NewProductHandler(productsUC)
	deleteLastProductHandler := deleteLastProduct.NewProductHandler(productsUC)
	closeLastReceptionHandler := close_last_reception.NewReceptionHandler(receptionsUC)
	pvzGetHandler := pvzGet.NewPVZDataHandler(pvzUC)

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

	log.Println(cfg.App.String())
	if err := app.Listen(cfg.App.String()); err != nil {
		panic("app not start")
	}
}
