package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/billykore/project-one/internal/api/dto"
	"github.com/billykore/project-one/internal/core/domain"
	"github.com/billykore/project-one/internal/core/ports"
	vo "github.com/billykore/project-one/internal/core/valueobject"
	"github.com/labstack/echo/v4"
)

// FeatureFlagHandler handles feature-flag evaluation and administration.
type FeatureFlagHandler struct {
	useCase     ports.FeatureFlagUseCase
	validator   ports.Validator
	environment string
}

// NewFeatureFlagHandler creates a feature-flag handler.
func NewFeatureFlagHandler(useCase ports.FeatureFlagUseCase, validator ports.Validator, environment string) *FeatureFlagHandler {
	return &FeatureFlagHandler{useCase: useCase, validator: validator, environment: environment}
}

// ListFlags handles GET /admin/feature-flags.
//
//	@Summary	List feature flags
//	@Tags		feature-flags
//	@Produce	json
//	@Success	200	{object}	dto.FeatureFlagListResponse
//	@Security	BearerAuth
//	@Router		/admin/feature-flags [get]
func (h *FeatureFlagHandler) ListFlags(c echo.Context) error {
	flags, err := h.useCase.ListFlags(c.Request().Context())
	if err != nil {
		return err
	}
	response := dto.FeatureFlagListResponse{Flags: make([]dto.FeatureFlagResponse, 0, len(flags))}
	for _, flag := range flags {
		response.Flags = append(response.Flags, flagResponse(flag))
	}
	return c.JSON(http.StatusOK, response)
}

// CreateFlag handles POST /admin/feature-flags.
//
//	@Summary	Create feature flag
//	@Tags		feature-flags
//	@Accept		json
//	@Produce	json
//	@Param		request	body		dto.CreateFeatureFlagRequest	true	"Flag details"
//	@Success	201		{object}	dto.FeatureFlagResponse
//	@Security	BearerAuth
//	@Router		/admin/feature-flags [post]
func (h *FeatureFlagHandler) CreateFlag(c echo.Context) error {
	user, ok := currentUser(c)
	if !ok {
		return echo.ErrUnauthorized
	}
	var request dto.CreateFeatureFlagRequest
	if err := c.Bind(&request); err != nil {
		return echo.ErrBadRequest
	}
	if err := h.validator.Validate(request); err != nil {
		return err
	}
	flag, err := h.useCase.CreateFlag(c.Request().Context(), user.Username, request.Key, request.Name, request.Purpose, request.Owner, request.SafeDefault)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, flagResponse(*flag))
}

// GetFlag handles GET /admin/feature-flags/:key.
//
//	@Summary	Get feature flag
//	@Tags		feature-flags
//	@Produce	json
//	@Param		key		path		string	true	"Flag key"
//	@Param		cursor	query		string	false	"Pagination cursor from the previous response"
//	@Param		limit	query		int		false	"Items per page (1-100, default 20)"
//	@Success	200		{object}	dto.FeatureFlagDetailResponse
//	@Security	BearerAuth
//	@Router		/admin/feature-flags/{key} [get]
func (h *FeatureFlagHandler) GetFlag(c echo.Context) error {
	detail, err := h.useCase.GetFlagDetail(c.Request().Context(), c.Param("key"))
	if err != nil {
		return err
	}
	response := dto.FeatureFlagDetailResponse{FeatureFlagResponse: flagResponse(detail.Flag)}
	response.Settings = make([]dto.FeatureFlagSettingResponse, 0, len(detail.Settings))
	for _, setting := range detail.Settings {
		response.Settings = append(response.Settings, dto.FeatureFlagSettingResponse{
			Environment: string(setting.Environment), Mode: string(setting.Mode),
			RolloutPercentage: setting.RolloutPercentage, Revision: setting.Revision,
		})
	}
	response.Overrides = make([]dto.FeatureFlagOverrideResponse, 0, len(detail.Overrides))
	for _, override := range detail.Overrides {
		response.Overrides = append(response.Overrides, dto.FeatureFlagOverrideResponse{
			Environment: string(override.Environment), Username: override.Username, Type: string(override.Type),
		})
	}
	return c.JSON(http.StatusOK, response)
}

// UpdateFlag handles PUT /admin/feature-flags/:key.
//
//	@Summary	Update feature flag metadata
//	@Tags		feature-flags
//	@Accept		json
//	@Produce	json
//	@Param		key		path		string							true	"Flag key"
//	@Param		request	body		dto.UpdateFeatureFlagRequest	true	"Flag metadata"
//	@Success	200		{object}	dto.FeatureFlagResponse
//	@Security	BearerAuth
//	@Router		/admin/feature-flags/{key} [put]
func (h *FeatureFlagHandler) UpdateFlag(c echo.Context) error {
	user, ok := currentUser(c)
	if !ok {
		return echo.ErrUnauthorized
	}
	var request dto.UpdateFeatureFlagRequest
	if err := c.Bind(&request); err != nil {
		return echo.ErrBadRequest
	}
	if err := h.validator.Validate(request); err != nil {
		return err
	}
	flag, err := h.useCase.UpdateFlag(c.Request().Context(), user.Username, c.Param("key"), request.Name, request.Purpose, request.Owner, request.SafeDefault, request.Reason)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, flagResponse(*flag))
}

// SetEnvironment handles PATCH /admin/feature-flags/:key/environment/:environment.
//
//	@Summary	Set environment availability
//	@Tags		feature-flags
//	@Accept		json
//	@Produce	json
//	@Param		key			path		string									true	"Flag key"
//	@Param		environment	path		string									true	"Environment"
//	@Param		request		body		dto.SetFeatureFlagEnvironmentRequest	true	"Availability"
//	@Success	200			{object}	dto.FeatureFlagSettingResponse
//	@Security	BearerAuth
//	@Router		/admin/feature-flags/{key}/environment/{environment} [patch]
func (h *FeatureFlagHandler) SetEnvironment(c echo.Context) error {
	user, ok := currentUser(c)
	if !ok {
		return echo.ErrUnauthorized
	}
	var request dto.SetFeatureFlagEnvironmentRequest
	if err := c.Bind(&request); err != nil {
		return echo.ErrBadRequest
	}
	if err := h.validator.Validate(request); err != nil {
		return err
	}
	setting, err := h.useCase.SetEnvironment(c.Request().Context(), user.Username, c.Param("key"), domain.Environment(c.Param("environment")), domain.AvailabilityMode(request.Mode), request.RolloutPercentage, request.Revision, request.Reason)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, dto.FeatureFlagSettingResponse{
		Environment: string(setting.Environment), Mode: string(setting.Mode), RolloutPercentage: setting.RolloutPercentage, Revision: setting.Revision,
	})
}

// SetOverrides handles PUT /admin/feature-flags/:key/overrides.
//
//	@Summary	Set user overrides
//	@Tags		feature-flags
//	@Accept		json
//	@Param		key		path	string								true	"Flag key"
//	@Param		request	body	dto.SetFeatureFlagOverridesRequest	true	"Overrides"
//	@Success	204
//	@Security	BearerAuth
//	@Router		/admin/feature-flags/{key}/overrides [put]
func (h *FeatureFlagHandler) SetOverrides(c echo.Context) error {
	user, ok := currentUser(c)
	if !ok {
		return echo.ErrUnauthorized
	}
	var request dto.SetFeatureFlagOverridesRequest
	if err := c.Bind(&request); err != nil {
		return echo.ErrBadRequest
	}
	if err := h.validator.Validate(request); err != nil {
		return err
	}
	if err := h.useCase.SetOverrides(c.Request().Context(), user.Username, c.Param("key"), domain.Environment(request.Environment), request.Include, request.Exclude, request.Reason); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

// Archive handles POST /admin/feature-flags/:key/archive.
//
//	@Summary	Archive feature flag
//	@Tags		feature-flags
//	@Accept		json
//	@Param		key		path	string							true	"Flag key"
//	@Param		request	body	dto.FeatureFlagArchiveRequest	true	"Archive reason"
//	@Success	204
//	@Security	BearerAuth
//	@Router		/admin/feature-flags/{key}/archive [post]
func (h *FeatureFlagHandler) Archive(c echo.Context) error {
	user, ok := currentUser(c)
	if !ok {
		return echo.ErrUnauthorized
	}
	var request dto.FeatureFlagArchiveRequest
	if err := c.Bind(&request); err != nil {
		return echo.ErrBadRequest
	}
	if err := h.validator.Validate(request); err != nil {
		return err
	}
	if err := h.useCase.Archive(c.Request().Context(), user.Username, c.Param("key"), request.Reason); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

// ListAudit handles GET /admin/feature-flags/:key/audit.
//
//	@Summary	List feature flag audit history
//	@Tags		feature-flags
//	@Produce	json
//	@Param		key	path		string	true	"Flag key"
//	@Success	200	{object}	dto.FeatureFlagAuditListResponse
//	@Security	BearerAuth
//	@Router		/admin/feature-flags/{key}/audit [get]
func (h *FeatureFlagHandler) ListAudit(c echo.Context) error {
	var cursor *vo.Cursor
	if cursorStr := c.QueryParam("cursor"); cursorStr != "" {
		decoded, err := vo.DecodeCursor(cursorStr)
		if err != nil {
			return domain.ErrInvalidCursor
		}
		cursor = &decoded
	}
	limit := 20
	if limitStr := c.QueryParam("limit"); limitStr != "" {
		parsed, err := strconv.Atoi(limitStr)
		if err != nil || parsed < 1 || parsed > 100 {
			return echo.ErrBadRequest
		}
		limit = parsed
	}
	records, hasMore, err := h.useCase.ListAudit(c.Request().Context(), c.Param("key"), cursor, limit)
	if err != nil {
		return err
	}
	response := dto.FeatureFlagAuditListResponse{Data: make([]dto.FeatureFlagAuditResponse, 0, len(records)), HasMore: hasMore}
	if len(records) > 0 && hasMore {
		response.NextCursor = (&vo.Cursor{ID: records[len(records)-1].ID}).Encode()
	}
	for _, record := range records {
		environment := ""
		if record.Environment != nil {
			environment = string(*record.Environment)
		}
		response.Data = append(response.Data, dto.FeatureFlagAuditResponse{
			Field: record.Field, Environment: environment, PreviousValue: record.PreviousValue,
			NewValue: record.NewValue, Actor: record.Actor, Reason: record.Reason, CreatedAt: record.CreatedAt,
		})
	}
	return c.JSON(http.StatusOK, response)
}

// Evaluate handles GET /feature-flags/evaluate?keys=a,b.
//
//	@Summary	Evaluate feature flags
//	@Tags		feature-flags
//	@Produce	json
//	@Param		keys	query		string	true	"Comma-separated flag keys"
//	@Success	200		{object}	dto.FeatureFlagEvaluateResponse
//	@Router		/feature-flags/evaluate [get]
func (h *FeatureFlagHandler) Evaluate(c echo.Context) error {
	keys := strings.Split(c.QueryParam("keys"), ",")
	if len(keys) == 1 && strings.TrimSpace(keys[0]) == "" || len(keys) > 50 {
		return echo.ErrBadRequest
	}
	username, _ := c.Get("username").(string)
	response := dto.FeatureFlagEvaluateResponse{Environment: h.environment, Decisions: make([]dto.FeatureFlagDecisionResponse, 0, len(keys))}
	for _, key := range keys {
		key = strings.TrimSpace(key)
		if key == "" {
			return echo.ErrBadRequest
		}
		decision := h.useCase.Evaluate(c.Request().Context(), key, username)
		response.Decisions = append(response.Decisions, dto.FeatureFlagDecisionResponse{Key: key, Enabled: decision.Enabled, Source: decision.Source})
	}
	return c.JSON(http.StatusOK, response)
}

func flagResponse(flag domain.FeatureFlag) dto.FeatureFlagResponse {
	return dto.FeatureFlagResponse{
		Key: flag.Key, Name: flag.Name, Purpose: flag.Purpose, Owner: flag.Owner,
		Lifecycle: string(flag.Lifecycle), SafeDefault: flag.SafeDefault,
		CreatedAt: flag.CreatedAt, UpdatedAt: flag.UpdatedAt,
	}
}
