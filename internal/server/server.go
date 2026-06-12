package server

import (
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/FelippeRibeiro/go-gallery/internal/handler"
	"github.com/FelippeRibeiro/go-gallery/internal/middleware"
)

// spaHandler serves static files from root. Unknown paths fall back to index.html
// so React Router can handle client-side navigation.
type spaHandler struct{ root string }

func (h spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	upath := path.Clean("/" + r.URL.Path)
	full := filepath.Join(h.root, filepath.FromSlash(upath))
	if _, err := os.Stat(full); os.IsNotExist(err) {
		http.ServeFile(w, r, filepath.Join(h.root, "index.html"))
		return
	}
	http.FileServer(http.Dir(h.root)).ServeHTTP(w, r)
}

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
	mux.HandleFunc("POST /api/auth/logout", handler.Logout)

	// Usuário autenticado (decodificado do cookie JWT)
	mux.HandleFunc("GET /api/me", middleware.RequireAuth(handler.Me))

	// Álbuns públicos — sem autenticação
	mux.HandleFunc("GET /api/public/albums", handler.ListPublicAlbums)
	mux.HandleFunc("GET /api/public/albums/{id}", handler.GetPublicAlbum)

	// Fotógrafos públicos
	mux.HandleFunc("GET /api/public/photographers", handler.ListPublicPhotographers)
	mux.HandleFunc("GET /api/public/photographers/{id}", handler.GetPublicPhotographer)

	// Perfil do fotógrafo autenticado
	mux.HandleFunc("GET /api/profile", middleware.RequireAuth(handler.GetProfile))
	mux.HandleFunc("PATCH /api/profile", middleware.RequireAuth(handler.UpdateProfile))
	mux.HandleFunc("POST /api/profile/photo", middleware.RequireAuth(handler.UploadProfilePhoto))

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
	mux.HandleFunc("GET /api/albums/{id}/photos", middleware.RequireAuth(handler.ListPhotos))
	mux.HandleFunc("PUT /api/albums/{id}/photos/reorder", middleware.RequireAuth(handler.ReorderPhotos))

	// Convites de clientes
	mux.HandleFunc("POST /api/albums/{id}/invite", middleware.RequireAuth(handler.InviteClient))
	mux.HandleFunc("GET /api/albums/{id}/invites", middleware.RequireAuth(handler.ListInvites))
	mux.HandleFunc("DELETE /api/albums/{id}/invites/{inviteId}", middleware.RequireAuth(handler.RevokeInvite))
	mux.HandleFunc("POST /api/albums/{id}/invites/{inviteId}/resend", middleware.RequireAuth(handler.ResendInvite))

	// Fotógrafos colaboradores
	mux.HandleFunc("POST /api/albums/{id}/invite-photographer", middleware.RequireAuth(handler.InvitePhotographer))
	mux.HandleFunc("DELETE /api/albums/{id}/photographers/{faId}", middleware.RequireAuth(handler.RemovePhotographer))

	// Clientes convidados
	mux.HandleFunc("DELETE /api/albums/{id}/clients/{clientId}", middleware.RequireAuth(handler.RemoveClient))

	// Pedidos
	mux.HandleFunc("POST /api/albums/{id}/orders", middleware.RequireAuth(handler.CreateOrder))
	mux.HandleFunc("GET /api/orders", middleware.RequireAuth(handler.ListMyOrders))
	mux.HandleFunc("GET /api/orders/{id}", middleware.RequireAuth(handler.GetOrder))
	mux.HandleFunc("POST /api/orders/{id}/checkout", middleware.RequireAuth(handler.CreateCheckout))
	mux.HandleFunc("POST /api/orders/{id}/pix", middleware.RequireAuth(handler.CreatePix))
	mux.HandleFunc("GET /api/orders/{id}/downloads", middleware.RequireAuth(handler.GetDownloadLinks))
	mux.HandleFunc("GET /api/orders/{id}/download", middleware.RequireAuth(handler.DownloadOrderZip))
	mux.HandleFunc("GET /api/orders/{id}/download/{fotoId}", middleware.RequireAuth(handler.DownloadOrderPhoto))

	// Webhook do Mercado Pago — público (sem autenticação)
	mux.HandleFunc("POST /api/webhooks/mercadopago", handler.MercadoPagoWebhook)

	// WebSocket
	mux.HandleFunc("GET /ws/albums/{id}", handler.ServeAlbumWS)

	// Frontend SPA — serve arquivos buildados; fallback para index.html
	mux.Handle("/", spaHandler{root: "public"})

	return middleware.CORS(mux)
}
