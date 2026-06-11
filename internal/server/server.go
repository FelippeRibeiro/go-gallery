package server

import (
	"net/http"

	"github.com/FelippeRibeiro/go-gallery/internal/handler"
	"github.com/FelippeRibeiro/go-gallery/internal/middleware"
)

func New() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// Auth — público
	mux.HandleFunc("POST /api/auth/register", handler.Register)
	mux.HandleFunc("POST /api/auth/login", handler.Login)
	mux.HandleFunc("POST /api/auth/accept-invite", handler.AcceptInvite)

	// Álbuns públicos — sem autenticação
	mux.HandleFunc("GET /api/public/albums", handler.ListPublicAlbums)
	mux.HandleFunc("GET /api/public/albums/{id}", handler.GetPublicAlbum)

	// Upload legado — protegido
	mux.HandleFunc("POST /api/upload", middleware.RequireAuth(handler.Upload))

	// Álbuns — protegidos
	mux.HandleFunc("GET /api/albums", middleware.RequireAuth(handler.ListAlbums))
	mux.HandleFunc("POST /api/albums", middleware.RequireAuth(handler.CreateAlbum))
	mux.HandleFunc("GET /api/albums/{id}", middleware.RequireAuth(handler.GetAlbum))
	mux.HandleFunc("PATCH /api/albums/{id}/visibility", middleware.RequireAuth(handler.UpdateVisibility))
	mux.HandleFunc("POST /api/albums/{id}/cover", middleware.RequireAuth(handler.UploadCover))

	// Fotos
	mux.HandleFunc("POST /api/albums/{id}/photos", middleware.RequireAuth(handler.UploadPhoto))
	mux.HandleFunc("DELETE /api/albums/{id}/photos/{photoId}", middleware.RequireAuth(handler.DeletePhoto))

	// Convites de clientes
	mux.HandleFunc("POST /api/albums/{id}/invite", middleware.RequireAuth(handler.InviteClient))
	mux.HandleFunc("GET /api/albums/{id}/invites", middleware.RequireAuth(handler.ListInvites))
	mux.HandleFunc("DELETE /api/albums/{id}/invites/{inviteId}", middleware.RequireAuth(handler.RevokeInvite))

	// Fotógrafos colaboradores
	mux.HandleFunc("POST /api/albums/{id}/invite-photographer", middleware.RequireAuth(handler.InvitePhotographer))
	mux.HandleFunc("DELETE /api/albums/{id}/photographers/{faId}", middleware.RequireAuth(handler.RemovePhotographer))

	// Clientes convidados
	mux.HandleFunc("DELETE /api/albums/{id}/clients/{clientId}", middleware.RequireAuth(handler.RemoveClient))

	// WebSocket
	mux.HandleFunc("GET /ws/albums/{id}", handler.ServeAlbumWS)

	return middleware.CORS(mux)
}
