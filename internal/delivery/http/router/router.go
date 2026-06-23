package router

import (
	"github.com/Jaruvat303/cashlog/internal/delivery/http/v1/handler"
	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App, txHandler *handler.TransactionHandler, catHandler *handler.CategoryHandler) {

	// กำหนดและผูกเส้นทาง API Endpoint ตามสัญญา RESTful Spec
	v1 := app.Group("/api/v1")

	v1.Post("/log", txHandler.UplaodSlipAndLog)
	v1.Get("/dashboard/summary", txHandler.GetDashboardSummary)
	v1.Get("/transactions", txHandler.GetMonthlyHistory)
	v1.Patch("/transactions/:id", txHandler.UpdateTransaction)
	v1.Delete("/transactions/:id", txHandler.DeleteTransaction)

	cat := v1.Group("/categories")
	cat.Post("/", catHandler.CreateCategory)
	cat.Get("/", catHandler.FetchCategories)
	cat.Patch("/:id", catHandler.UpdateCategory)
	cat.Delete("/:id", catHandler.DeleteCategory)

}
