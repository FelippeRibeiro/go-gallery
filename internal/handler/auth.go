package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/FelippeRibeiro/go-gallery/internal/db"
	"github.com/FelippeRibeiro/go-gallery/internal/middleware"
	"golang.org/x/crypto/bcrypt"
)

const (
	minSenhaLen   = 6
	msgSenhaCurta = "a senha deve ter pelo menos 6 caracteres"
)

type userDTO struct {
	ID    int64  `json:"id"`
	Nome  string `json:"nome"`
	Email string `json:"email"`
	Tipo  string `json:"tipo"`
}

func toUserDTO(u db.Usuario) userDTO {
	return userDTO{
		ID:    u.ID,
		Nome:  u.Nome,
		Email: u.Email,
		Tipo:  fmt.Sprintf("%v", u.Tipo),
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// clientIP extrai o IP do cliente do RemoteAddr (sem confiar em headers
// encaminhados, que são forjáveis).
func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// POST /api/auth/register
func Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Nome  string `json:"nome"`
		Email string `json:"email"`
		Senha string `json:"senha"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "corpo inválido")
		return
	}
	if req.Nome == "" || req.Email == "" || req.Senha == "" {
		writeError(w, http.StatusBadRequest, "nome, email e senha são obrigatórios")
		return
	}
	if utf8.RuneCountInString(req.Senha) < minSenhaLen {
		writeError(w, http.StatusBadRequest, msgSenhaCurta)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Senha), bcrypt.DefaultCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao processar senha")
		return
	}

	queries := db.GetQueries()
	user, err := queries.CriarFotografo(r.Context(), db.CriarFotografoParams{
		Nome:      req.Nome,
		Email:     req.Email,
		SenhaHash: sql.NullString{String: string(hash), Valid: true},
	})
	if err != nil {
		writeError(w, http.StatusConflict, "email já cadastrado")
		return
	}

	token, err := middleware.GenerateToken(user.ID, "fotografo")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao gerar token")
		return
	}

	middleware.SetTokenCookie(w, token)
	writeJSON(w, http.StatusCreated, map[string]any{
		"user": toUserDTO(user),
	})
}

// POST /api/auth/login
func Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
		Senha string `json:"senha"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "corpo inválido")
		return
	}

	// Proteção contra força bruta: limita tentativas por IP+email; um login
	// correto zera o contador.
	limiterKey := clientIP(r) + "|" + strings.ToLower(req.Email)
	if !middleware.LoginAllowed(limiterKey) {
		writeError(w, http.StatusTooManyRequests, "muitas tentativas de login — tente novamente em alguns minutos")
		return
	}

	queries := db.GetQueries()
	user, err := queries.ObterUsuarioPorEmail(r.Context(), req.Email)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "credenciais inválidas")
		return
	}

	if !user.SenhaHash.Valid {
		writeError(w, http.StatusUnauthorized, "credenciais inválidas")
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.SenhaHash.String), []byte(req.Senha)); err != nil {
		writeError(w, http.StatusUnauthorized, "credenciais inválidas")
		return
	}
	middleware.LoginSucceeded(limiterKey)

	tipo := fmt.Sprintf("%v", user.Tipo)
	token, err := middleware.GenerateToken(user.ID, tipo)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao gerar token")
		return
	}

	middleware.SetTokenCookie(w, token)
	writeJSON(w, http.StatusOK, map[string]any{
		"user": toUserDTO(user),
	})
}

// POST /api/auth/accept-invite
func AcceptInvite(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token string `json:"token"`
		Nome  string `json:"nome"`
		Senha string `json:"senha"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "corpo inválido")
		return
	}
	if req.Token == "" || req.Nome == "" || req.Senha == "" {
		writeError(w, http.StatusBadRequest, "token, nome e senha são obrigatórios")
		return
	}
	if utf8.RuneCountInString(req.Senha) < minSenhaLen {
		writeError(w, http.StatusBadRequest, msgSenhaCurta)
		return
	}

	queries := db.GetQueries()
	convite, err := queries.ObterConvitePorToken(r.Context(), req.Token)
	if err != nil {
		writeError(w, http.StatusNotFound, "convite não encontrado")
		return
	}
	if convite.UsedAt.Valid {
		writeError(w, http.StatusGone, "convite já utilizado")
		return
	}
	if time.Now().After(convite.ExpiresAt) {
		writeError(w, http.StatusGone, "convite expirado")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Senha), bcrypt.DefaultCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao processar senha")
		return
	}

	// Check if user already exists with this email
	existingUser, err := queries.ObterUsuarioPorEmail(r.Context(), convite.Email)
	var user db.Usuario
	if err == nil {
		// User already exists — just associate with album
		user = existingUser
	} else if errors.Is(err, sql.ErrNoRows) {
		user, err = queries.CriarCliente(r.Context(), db.CriarClienteParams{
			Nome:      req.Nome,
			Email:     convite.Email,
			SenhaHash: sql.NullString{String: string(hash), Valid: true},
		})
		if err != nil {
			writeError(w, http.StatusInternalServerError, "erro ao criar conta")
			return
		}
	} else {
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}

	queries.AssociarClienteAlbum(r.Context(), user.ID, convite.IDAlbum) //nolint:errcheck
	queries.MarcarConviteUsado(r.Context(), req.Token)                  //nolint:errcheck

	// Usa o tipo real do usuário: se o e-mail convidado já pertencia a uma
	// conta existente, o token não pode rebaixá-la para 'cliente'.
	token, err := middleware.GenerateToken(user.ID, fmt.Sprintf("%v", user.Tipo))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao gerar token")
		return
	}

	middleware.SetTokenCookie(w, token)
	writeJSON(w, http.StatusOK, map[string]any{
		"user":     toUserDTO(user),
		"album_id": convite.IDAlbum,
	})
}

// GET /api/me — retorna o usuário autenticado (decodificado do cookie JWT).
// Fonte da verdade do frontend para restaurar a sessão.
func Me(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int64)
	user, err := db.GetQueries().ObterUsuarioPorID(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "usuário não encontrado")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"user": toUserDTO(user)})
}

// POST /api/auth/logout — limpa o cookie de autenticação.
func Logout(w http.ResponseWriter, r *http.Request) {
	middleware.ClearTokenCookie(w)
	w.WriteHeader(http.StatusNoContent)
}
