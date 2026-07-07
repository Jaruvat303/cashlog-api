package router

import (
	"github.com/Jaruvat303/cashlog/internal/delivery/http/v1/handler"
	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App, txHandler *handler.TransactionHandler, catHandler *handler.CategoryHandler) {

	// กำหนดและผูกเส้นทาง API Endpoint ตามสัญญา RESTful Spec
	v1 := app.Group("/api/v1")

	tx := v1.Group("/transactions")
	tx.Get("/", txHandler.GetMonthlyHistory)
	v1.Get("/summary", txHandler.GetDashboardSummary)
	v1.Post("/upload-slip", txHandler.UplaodSlipAndLog)
	tx.Patch("/:id", txHandler.UpdateTransaction)
	tx.Delete("/:id", txHandler.DeleteTransaction)

	cat := v1.Group("/categories")
	cat.Post("/", catHandler.CreateCategory)
	cat.Get("/", catHandler.FetchCategoriesByType)
	cat.Patch("/:id", catHandler.UpdateCategory)
	cat.Delete("/:id", catHandler.DeleteCategory)

}
