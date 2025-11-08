package echokit

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

// ErrorHandler returns a custom HTTP error handler that wraps non-HTTPError errors
// in HTTPError with status 500 and includes request ID in the response message,
// then delegates to Echo's default error handler to send the response.
func ErrorHandler(e *echo.Echo) echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		if c.Response().Committed {
			return
		}

		// Check if error is already an HTTPError
		he, ok := err.(*echo.HTTPError)
		if !ok {
			// Not an HTTPError - wrap in HTTPError with request ID
			requestID := c.Request().Header.Get(echo.HeaderXRequestID)
			message := fmt.Sprintf("%s (request_id: %s)", http.StatusText(http.StatusInternalServerError), requestID)
			he = echo.NewHTTPError(http.StatusInternalServerError, message)
			he.Internal = err
		}

		// Delegate to Echo's default error handler to send the response
		e.DefaultHTTPErrorHandler(he, c)
	}
}
