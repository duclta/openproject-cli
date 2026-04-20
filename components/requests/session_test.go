package requests

import (
	"net/http"
	"testing"

	operrors "github.com/opf/openproject-cli/components/errors"
)

func TestLoginErrorFromResponse_ExtractsBannerMessage(t *testing.T) {
	body := []byte(`<html><body><p class="Banner-title" data-target="x-banner.titleText">Invalid user or password or the account is blocked due to multiple failed login attempts. If so, it will be unblocked automatically in a short time.</p></body></html>`)

	err := loginErrorFromResponse(http.StatusUnprocessableEntity, body)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	expected := "Invalid user or password or the account is blocked due to multiple failed login attempts. If so, it will be unblocked automatically in a short time."
	if err.Error() != expected {
		t.Fatalf("expected %q, got %q", expected, err.Error())
	}
}

func TestLoginErrorFromResponse_FallsBackToResponseError(t *testing.T) {
	body := []byte(`{"error":"unprocessable"}`)

	err := loginErrorFromResponse(http.StatusUnprocessableEntity, body)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	responseErr, ok := err.(*operrors.ResponseError)
	if !ok {
		t.Fatalf("expected ResponseError, got %T", err)
	}

	if responseErr.Status() != http.StatusUnprocessableEntity {
		t.Fatalf("expected status %d, got %d", http.StatusUnprocessableEntity, responseErr.Status())
	}

	if string(responseErr.Response()) != string(body) {
		t.Fatalf("expected body %q, got %q", string(body), string(responseErr.Response()))
	}
}
