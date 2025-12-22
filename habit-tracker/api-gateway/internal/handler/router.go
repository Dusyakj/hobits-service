package handler

import (
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger"

	"api-gateway/internal/middleware"
)

// Router sets up HTTP routes
type Router struct {
	userHandler    *UserHandler
	habitHandler   *HabitHandler
	authMiddleware *middleware.AuthMiddleware
	mux            *http.ServeMux
}

// NewRouter creates a new router
func NewRouter(userHandler *UserHandler, habitHandler *HabitHandler, authMiddleware *middleware.AuthMiddleware) *Router {
	return &Router{
		userHandler:    userHandler,
		habitHandler:   habitHandler,
		authMiddleware: authMiddleware,
		mux:            http.NewServeMux(),
	}
}

// Setup configures all routes
func (r *Router) Setup() http.Handler {

	r.mux.HandleFunc("/api/v1/auth/register", r.userHandler.Register)
	r.mux.HandleFunc("/api/v1/auth/login", r.userHandler.Login)
	r.mux.HandleFunc("/api/v1/auth/refresh", r.userHandler.RefreshToken)
	r.mux.HandleFunc("/api/v1/auth/verify-email", r.userHandler.VerifyEmail)
	r.mux.HandleFunc("/api/v1/auth/resend-verification", r.userHandler.ResendVerificationEmail)
	r.mux.HandleFunc("/api/v1/auth/forgot-password", r.userHandler.ForgotPassword)
	r.mux.HandleFunc("/api/v1/auth/reset-password", r.userHandler.ResetPassword)

	r.mux.HandleFunc("/api/v1/auth/logout", r.authMiddleware.Auth(r.userHandler.Logout))
	r.mux.HandleFunc("/api/v1/users/profile", r.authMiddleware.Auth(r.userHandler.GetProfile))
	r.mux.HandleFunc("/api/v1/users/change-password", r.authMiddleware.Auth(r.userHandler.ChangePassword))
	r.mux.HandleFunc("/api/v1/users/deactivate", r.authMiddleware.Auth(r.userHandler.DeactivateAccount))

	// Habit routes (all require authentication)
	r.mux.HandleFunc("/api/v1/habits/create", r.authMiddleware.Auth(r.habitHandler.CreateHabit))
	r.mux.HandleFunc("/api/v1/habits/list", r.authMiddleware.Auth(r.habitHandler.ListHabits))
	r.mux.HandleFunc("/api/v1/habits/get", r.authMiddleware.Auth(r.habitHandler.GetHabit))
	r.mux.HandleFunc("/api/v1/habits/update", r.authMiddleware.Auth(r.habitHandler.UpdateHabit))
	r.mux.HandleFunc("/api/v1/habits/delete", r.authMiddleware.Auth(r.habitHandler.DeleteHabit))
	r.mux.HandleFunc("/api/v1/habits/confirm", r.authMiddleware.Auth(r.habitHandler.ConfirmHabit))
	r.mux.HandleFunc("/api/v1/habits/history", r.authMiddleware.Auth(r.habitHandler.GetHabitHistory))
	r.mux.HandleFunc("/api/v1/habits/stats", r.authMiddleware.Auth(r.habitHandler.GetHabitStats))

	r.mux.HandleFunc("/swagger/", httpSwagger.WrapHandler)

	r.mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	var handler http.Handler = r.mux

	handler = middleware.Logging(handler)

	handler = middleware.RateLimit(60)(handler)

	return handler
}
