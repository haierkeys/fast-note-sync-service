package api_router

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/haierkeys/fast-note-sync-service/internal/app"
	"github.com/haierkeys/fast-note-sync-service/internal/config"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	"github.com/haierkeys/fast-note-sync-service/internal/oauth"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"go.uber.org/zap"
)

type StytchOAuthHandler struct {
	*Handler
	client stytchAuthorizer
}

type stytchAuthorizer interface {
	AuthorizeStart(ctx *gin.Context, params oauth.StytchAuthorizeParams) (*oauth.StytchAuthorizeStartResponse, error)
	AuthorizeSubmit(ctx *gin.Context, params oauth.StytchAuthorizeParams) (*oauth.StytchAuthorizeSubmitResponse, error)
}

type StytchOAuthAuthorizeRequest struct {
	ClientID            string `json:"client_id" binding:"required"`
	RedirectURI         string `json:"redirect_uri" binding:"required"`
	ResponseType        string `json:"response_type"`
	Scope               string `json:"scope"`
	State               string `json:"state"`
	Nonce               string `json:"nonce"`
	CodeChallenge       string `json:"code_challenge"`
	CodeChallengeMethod string `json:"code_challenge_method"`
}

type StytchOAuthAuthorizeSubmitRequest struct {
	StytchOAuthAuthorizeRequest
	ConsentGranted bool `json:"consent_granted"`
}

type stytchOAuthClient struct {
	client *oauth.StytchClient
}

func (c *stytchOAuthClient) AuthorizeStart(ctx *gin.Context, params oauth.StytchAuthorizeParams) (*oauth.StytchAuthorizeStartResponse, error) {
	return c.client.AuthorizeStart(ctx.Request.Context(), params)
}

func (c *stytchOAuthClient) AuthorizeSubmit(ctx *gin.Context, params oauth.StytchAuthorizeParams) (*oauth.StytchAuthorizeSubmitResponse, error) {
	return c.client.AuthorizeSubmit(ctx.Request.Context(), params)
}

func NewStytchOAuthHandler(a *app.App) *StytchOAuthHandler {
	cfg := a.Config().OAuth.Stytch
	return &StytchOAuthHandler{
		Handler: NewHandler(a),
		client: &stytchOAuthClient{client: oauth.NewStytchClient(oauth.StytchClientConfig{
			Domain:    cfg.Domain,
			ProjectID: cfg.ProjectID,
			Secret:    cfg.Secret,
			Kind:      cfg.Kind,
		})},
	}
}

func (h *StytchOAuthHandler) AuthorizeStart(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	params := &StytchOAuthAuthorizeRequest{}
	valid, errs := pkgapp.BindAndValid(c, params)
	if !valid {
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	user, ok := h.currentUser(c)
	if !ok {
		response.ToResponse(code.ErrorNotUserAuthToken)
		return
	}

	cfg := h.App.Config().OAuth
	if !cfg.Stytch.Enabled {
		response.ToResponse(code.ErrorInvalidParams.WithDetails("stytch oauth is disabled"))
		return
	}

	stytchParams := buildStytchAuthorizeParams(cfg, params.toBase(), user, false)
	resp, err := h.client.AuthorizeStart(c, stytchParams)
	if err != nil {
		h.App.Logger().Error("StytchOAuthHandler.AuthorizeStart",
			zap.Error(err),
			zap.String("client_id", params.ClientID),
			zap.String("redirect_uri", params.RedirectURI),
			zap.Strings("scopes", stytchParams.Scopes),
		)
		response.ToResponse(code.ErrorInvalidParams.WithDetails("stytch authorize start failed"))
		return
	}

	response.ToResponse(code.Success.WithData(resp))
}

func (h *StytchOAuthHandler) AuthorizeSubmit(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	params := &StytchOAuthAuthorizeSubmitRequest{}
	valid, errs := pkgapp.BindAndValid(c, params)
	if !valid {
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	user, ok := h.currentUser(c)
	if !ok {
		response.ToResponse(code.ErrorNotUserAuthToken)
		return
	}

	cfg := h.App.Config().OAuth
	if !cfg.Stytch.Enabled {
		response.ToResponse(code.ErrorInvalidParams.WithDetails("stytch oauth is disabled"))
		return
	}

	stytchParams := buildStytchAuthorizeParams(cfg, params.toBase(), user, params.ConsentGranted)
	resp, err := h.client.AuthorizeSubmit(c, stytchParams)
	if err != nil {
		h.App.Logger().Error("StytchOAuthHandler.AuthorizeSubmit",
			zap.Error(err),
			zap.String("client_id", params.ClientID),
			zap.String("redirect_uri", params.RedirectURI),
			zap.Strings("scopes", stytchParams.Scopes),
		)
		response.ToResponse(code.ErrorInvalidParams.WithDetails("stytch authorize submit failed"))
		return
	}

	response.ToResponse(code.Success.WithData(resp))
}

func (h *StytchOAuthHandler) currentUser(c *gin.Context) (*dto.UserDTO, bool) {
	uid := pkgapp.GetUID(c)
	if uid == 0 {
		return nil, false
	}
	user, err := h.App.UserService.GetInfo(c.Request.Context(), uid)
	if err != nil || user == nil {
		return nil, false
	}
	return user, true
}

func (r *StytchOAuthAuthorizeRequest) toBase() StytchOAuthAuthorizeRequest {
	return *r
}

func (r *StytchOAuthAuthorizeSubmitRequest) toBase() StytchOAuthAuthorizeRequest {
	return r.StytchOAuthAuthorizeRequest
}

func buildStytchAuthorizeParams(cfg config.OAuthConfig, req StytchOAuthAuthorizeRequest, user *dto.UserDTO, consentGranted bool) oauth.StytchAuthorizeParams {
	responseType := req.ResponseType
	if responseType == "" {
		responseType = "code"
	}

	params := oauth.StytchAuthorizeParams{
		ClientID:            req.ClientID,
		RedirectURI:         req.RedirectURI,
		ResponseType:        responseType,
		Scopes:              strings.Fields(req.Scope),
		State:               req.State,
		Nonce:               req.Nonce,
		CodeChallenge:       req.CodeChallenge,
		CodeChallengeMethod: req.CodeChallengeMethod,
		Resource:            cfg.Resource,
		ConsentGranted:      consentGranted,
	}
	if cfg.Stytch.Kind == oauth.StytchKindB2B {
		params.OrganizationID = cfg.Stytch.OrganizationID
		params.MemberID = cfg.Stytch.MemberID
	} else if cfg.Stytch.UserID != "" {
		params.UserID = cfg.Stytch.UserID
	} else {
		params.UserID = fmt.Sprintf("%s%d", cfg.Stytch.UserIDPrefix, user.UID)
	}
	return params
}
