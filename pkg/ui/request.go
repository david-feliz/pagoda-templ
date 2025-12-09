package ui

import (
	"net/url"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"github.com/mikestefanello/pagoda/config"
	"github.com/mikestefanello/pagoda/ent"
	ctx "github.com/mikestefanello/pagoda/pkg/context"
)

type (
	// Request contains the context for the current request
	Request struct {
		Context     echo.Context
		Config      *config.Config
		URL         *url.URL
		CurrentPath string
		IsAuth      bool
		IsAdmin     bool
		AuthUser    *ent.User
		Msg         string
		Title       string
		Metatags    Metatags
		CSRF        string
		Htmx        Htmx
	}

	// Metatags contains the metatags for the current request
	Metatags struct {
		Description string
		Keywords    []string
	}

	// Htmx contains the HTMX context for the current request
	Htmx struct {
		Enabled bool
		Boosted bool
		Target  string
		Trigger string
	}
)

// NewRequest creates a new Request
func NewRequest(c echo.Context) *Request {
	u, ok := c.Get(ctx.AuthenticatedUserKey).(*ent.User)

	cfg, _ := c.Get(ctx.ConfigKey).(*config.Config)
	csrf, _ := c.Get(ctx.CSRFKey).(string)

	r := &Request{
		Context:     c,
		Config:      cfg,
		URL:         c.Request().URL,
		CurrentPath: c.Request().URL.Path,
		IsAuth:      ok && u != nil,
		AuthUser:    u,
		IsAdmin:     false,
		CSRF:        csrf,
	}

	if r.IsAuth {
		r.IsAdmin = r.AuthUser.Admin
	}

	// Check for HTMX request
	if c.Request().Header.Get("HX-Request") == "true" {
		r.Htmx.Enabled = true
		r.Htmx.Boosted = c.Request().Header.Get("HX-Boosted") == "true"
		r.Htmx.Target = c.Request().Header.Get("HX-Target")
		r.Htmx.Trigger = c.Request().Header.Get("HX-Trigger")
	}

	return r
}

// Path returns the URL path for the given route name and parameters
func (r *Request) Path(name string, params ...any) string {
	return r.Context.Echo().Reverse(name, params...)
}

// IsActive returns true if the given path is the current path
func (r *Request) IsActive(path string) bool {
	return r.CurrentPath == path
}

// RenderTempl renders a Templ component
func (r *Request) RenderTempl(layout func(*Request, templ.Component) templ.Component, component templ.Component) error {
	// If HTMX is enabled and not boosted, render only the component (partial)
	if r.Htmx.Enabled && !r.Htmx.Boosted {
		// If a specific target is requested, we might want to handle that logic here or in the handler.
		// For now, we assume the component passed is what needs to be rendered.
		// However, for OOB swaps, we might need more complex logic.
		// For simple partials:
		return component.Render(r.Context.Request().Context(), r.Context.Response().Writer)
	}

	// Otherwise, render the full layout
	if layout != nil {
		return layout(r, component).Render(r.Context.Request().Context(), r.Context.Response().Writer)
	}

	return component.Render(r.Context.Request().Context(), r.Context.Response().Writer)
}
