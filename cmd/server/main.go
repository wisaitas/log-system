package main

import (
	"encoding/json"
	"log-system/pkg"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

type Request struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type Response struct {
	FullName string `json:"full_name"`
}

func main() {
	app := fiber.New()

	app.Use(pkg.NewLogger("server"))

	app.Post("/do", func(c *fiber.Ctx) error {
		client := &http.Client{}
		payload := map[string]any{
			"first_name": "John",
			"last_name":  "Doe",
		}

		pkg.DownStreamHttp(c, http.MethodPost, "http://localhost:8082/do/b", payload, &response)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		if response.StatusCode != http.StatusOK {
			return c.Status(response.StatusCode).JSON(fiber.Map{
				"error": response.Status,
			})
		}
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return c.Status(resp.StatusCode).JSON(fiber.Map{
				"error": resp.Status,
			})
		}

		response := Response{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"full_name": response.FullName,
		})
	})

	app.Listen(":8081")
}

// import (
// 	"context"
// 	"errors"
// 	"log/slog"
// 	"net/http"
// 	"os"
// 	"time"

// 	"log-system/pkg"

// 	"github.com/gofiber/contrib/otelfiber/v2"
// 	"github.com/gofiber/fiber/v2"
// 	"go.opentelemetry.io/otel"
// 	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
// 	"go.opentelemetry.io/otel/sdk/resource"
// 	sdktrace "go.opentelemetry.io/otel/sdk/trace"
// 	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
// )

// func main() {
// 	// --- Logger ---
// 	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo, AddSource: true})))

// 	// --- Minimal OTel (stdout exporter) ---
// 	shutdown := mustSetupTracer("svc-a")
// 	defer shutdown()

// 	app := fiber.New(fiber.Config{ErrorHandler: pkg.ErrorHandlerJSON()})
// 	app.Use(otelfiber.Middleware()) // create spans + inject/extract trace context
// 	app.Use(pkg.NewLogging("svc-a", &pkg.Options{EnableCPU: false}))

// 	app.Post("/do", func(c *fiber.Ctx) error {
// 		// Example: call downstream (svc-b)
// 		lb := pkg.DoDownstream(c.UserContext(), http.MethodPost, "http://localhost:8082/do", map[string]any{
// 			"first_name": "John", "last_name": "Doe",
// 		})
// 		if lb != nil {
// 			pkg.AppendDownstream(c, *lb)
// 		}

// 		// Simulate business error
// 		return errors.New("error in current service")
// 	})

// 	slog.Info("server starting", slog.String("addr", ":8081"))
// 	if err := app.Listen(":8081"); err != nil {
// 		panic(err)
// 	}
// }

// // mustSetupTracer configures a basic stdout trace exporter so trace_id becomes valid.
// func mustSetupTracer(service string) func() {
// 	exp, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
// 	if err != nil {
// 		panic(err)
// 	}
// 	r, err := resource.New(context.Background(), resource.WithAttributes(semconv.ServiceNameKey.String(service)))
// 	if err != nil {
// 		panic(err)
// 	}
// 	tp := sdktrace.NewTracerProvider(
// 		sdktrace.WithBatcher(exp, sdktrace.WithMaxExportBatchSize(10), sdktrace.WithBatchTimeout(500*time.Millisecond)),
// 		sdktrace.WithResource(r),
// 	)
// 	otel.SetTracerProvider(tp)
// 	return func() { _ = tp.Shutdown(context.Background()) }
// }
