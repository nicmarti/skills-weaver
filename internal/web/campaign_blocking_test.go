package web

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"dungeons/internal/adventure"
	"dungeons/internal/llm"
	"dungeons/internal/tarot"
)

// failingStreamClient simulates an OpenRouter failure for handler tests: it
// implements llm.Client and returns the scripted normalized error from both
// Complete and Stream. No network access.
type failingStreamClient struct{ err error }

func (f *failingStreamClient) Complete(ctx context.Context, req llm.Request) (llm.Response, error) {
	return llm.Response{}, f.err
}

func (f *failingStreamClient) Stream(ctx context.Context, req llm.Request, obs llm.StreamObserver) (llm.Response, error) {
	return llm.Response{}, f.err
}

func timeoutLLMError() *llm.Error {
	return &llm.Error{
		Kind:      llm.ErrKindTimeout,
		Message:   "request deadline exceeded",
		Retryable: true,
	}
}

// newBlockingTestServer builds a Server wired to a test engine with the
// creation routes registered. Templates and data files are read from the
// repository root (tests must chdir there first).
func newBlockingTestServer(t *testing.T, client llm.Client) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	s := &Server{
		engine:       engine,
		templatesDir: "web/templates",
		staticDir:    "web/static",
		llmCfg:       llm.DefaultModels(),
		llmClient:    client,
	}
	s.setupTemplates()
	engine.POST("/adventures", s.handleCreateAdventure)
	engine.POST("/adventure", s.handleCreateAdventureWizard)
	return engine
}

// adventureDirAbsent asserts the adventure directory was never created:
// a failed campaign-plan generation must not leave a playable adventure
// behind (docs/issues-2026-09-24.md).
func adventureDirAbsent(t *testing.T, slug string) {
	t.Helper()
	if _, err := os.Stat(filepath.Join(adventuresDir, slug)); !os.IsNotExist(err) {
		t.Errorf("adventure directory %q must not exist after a failed generation (stat err: %v)", slug, err)
	}
}

func cleanupAdventure(t *testing.T, slug string) {
	t.Helper()
	if _, err := os.Stat(filepath.Join(adventuresDir, slug)); err == nil {
		if err := adventure.Delete(adventuresDir, slug); err != nil {
			t.Logf("cleanup of %s failed: %v", slug, err)
		}
	}
}

// TestCreateAdventure_TimeoutIsBlocking proves the production incident fix:
// when the campaign-plan generation times out, the quick-create form answers
// with a blocking error page and no adventure is persisted.
func TestCreateAdventure_TimeoutIsBlocking(t *testing.T) {
	t.Chdir("../..")

	engine := newBlockingTestServer(t, &failingStreamClient{err: timeoutLLMError()})

	form := url.Values{}
	form.Set("name", "ZZ Test Blocking Timeout")
	form.Set("theme", "Un culte s'infiltre à la cour")
	form.Set("duration", "short")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/adventures", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want 502 (timeout is a blocking error)", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "temps imparti") {
		t.Errorf("error page must explain the timeout, got: %s", body)
	}
	slug := adventure.New("ZZ Test Blocking Timeout", "").Slug
	if strings.Contains(body, "/play/"+slug) {
		t.Error("the failed adventure must not appear as a playable card in the rendered list")
	}
	adventureDirAbsent(t, slug)
}

// TestCreateAdventure_MissingKeyIsBlocking: a theme requires a plan, so a
// missing OpenRouter key must block creation instead of silently skipping.
func TestCreateAdventure_MissingKeyIsBlocking(t *testing.T) {
	t.Chdir("../..")

	engine := newBlockingTestServer(t, nil)

	form := url.Values{}
	form.Set("name", "ZZ Test Blocking NoKey")
	form.Set("theme", "un thème")
	form.Set("duration", "short")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/adventures", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", w.Code)
	}
	if !strings.Contains(w.Body.String(), "OPENROUTER_API_KEY") {
		t.Errorf("error page must mention the missing key, got: %s", w.Body.String())
	}
	adventureDirAbsent(t, adventure.New("ZZ Test Blocking NoKey", "").Slug)
}

// TestCreateAdventure_NoThemeStillCreates guards the non-regression: the
// quick-create form without a theme never generates a plan and must keep
// creating adventures directly.
func TestCreateAdventure_NoThemeStillCreates(t *testing.T) {
	t.Chdir("../..")

	engine := newBlockingTestServer(t, nil)

	slug := adventure.New("ZZ Test Blocking NoTheme", "").Slug
	cleanupAdventure(t, slug)
	defer cleanupAdventure(t, slug)

	form := url.Values{}
	form.Set("name", "ZZ Test Blocking NoTheme")
	form.Set("duration", "oneshot")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/adventures", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want 303 redirect to the play page", w.Code)
	}
	if !strings.Contains(w.Header().Get("Location"), "/play/"+slug) {
		t.Errorf("redirect = %q, want /play/%s", w.Header().Get("Location"), slug)
	}
	if _, err := os.Stat(filepath.Join(adventuresDir, slug)); err != nil {
		t.Errorf("adventure should exist without a plan when no theme is given: %v", err)
	}
}

// wizardPayload builds the JSON body the fortune-teller wizard posts, with a
// valid card per question.
func wizardPayload(t *testing.T, name string) map[string]interface{} {
	t.Helper()
	deck, err := tarot.LoadDeck("")
	if err != nil {
		t.Fatalf("load tarot deck: %v", err)
	}
	cards := make([]string, 0, len(deck.Questions))
	for _, q := range deck.Questions {
		deckCards := deck.CardsForQuestion(q.ID)
		if len(deckCards) == 0 {
			t.Fatalf("no cards for question %s", q.ID)
		}
		cards = append(cards, deckCards[0].ID)
	}
	return map[string]interface{}{
		"name":     name,
		"theme":    "Une ombre grandit sur les Quatre Royaumes",
		"duration": "campaign",
		"card_ids": cards,
		"debug":    false,
	}
}

// TestWizardCreate_TimeoutIsBlocking proves the wizard's contract: a
// campaign-plan generation failure (here: the production timeout) answers
// with a JSON error the front-end displays, and no adventure is persisted.
func TestWizardCreate_TimeoutIsBlocking(t *testing.T) {
	t.Chdir("../..")

	engine := newBlockingTestServer(t, &failingStreamClient{err: timeoutLLMError()})

	w := httptest.NewRecorder()
	body, err := json.Marshal(wizardPayload(t, "ZZ Wizard Blocking Timeout"))
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/adventure", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want 502 (timeout is a blocking error)", w.Code)
	}
	respBody := w.Body.String()
	if !strings.Contains(respBody, "Aucune aventure") {
		t.Errorf("JSON error must state no adventure was created, got: %s", respBody)
	}
	if !strings.Contains(respBody, "temps imparti") {
		t.Errorf("JSON error must explain the timeout, got: %s", respBody)
	}
	adventureDirAbsent(t, adventure.New("ZZ Wizard Blocking Timeout", "").Slug)
}
