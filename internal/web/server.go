package web

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// Server represents the web server.
type Server struct {
	engine         *gin.Engine
	sessionManager *SessionManager
	templatesDir   string
	staticDir      string
	port           int
	apiKey         string // Anthropic API key for campaign plan generation
}

// Config holds server configuration.
type Config struct {
	Port         int
	APIKey       string
	TemplatesDir string
	StaticDir    string
	Debug        bool
}

// NewServer creates a new web server with the given configuration.
func NewServer(cfg Config) *Server {
	if !cfg.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	engine.Use(gin.Recovery())

	if cfg.Debug {
		engine.Use(gin.Logger())
	}

	server := &Server{
		engine:         engine,
		sessionManager: NewSessionManager(cfg.APIKey),
		templatesDir:   cfg.TemplatesDir,
		staticDir:      cfg.StaticDir,
		port:           cfg.Port,
		apiKey:         cfg.APIKey,
	}

	server.setupTemplates()
	server.setupRoutes()

	return server
}

// setupTemplates loads HTML templates with custom functions.
func (s *Server) setupTemplates() {
	funcMap := template.FuncMap{
		"safe": func(str string) template.HTML {
			return template.HTML(str)
		},
		"formatDuration": func(ms int64) string {
			if ms < 1000 {
				return fmt.Sprintf("%dms", ms)
			}
			return fmt.Sprintf("%.1fs", float64(ms)/1000)
		},
		"lower": strings.ToLower,
		// iterate creates a slice of integers from 0 to n-1 for range loops
		"iterate": func(n int) []int {
			result := make([]int, n)
			for i := range result {
				result[i] = i
			}
			return result
		},
	}

	// Collect all template files
	var allFiles []string

	patterns := []string{
		filepath.Join(s.templatesDir, "*.html"),
		filepath.Join(s.templatesDir, "partials", "*.html"),
	}

	for _, pattern := range patterns {
		files, err := filepath.Glob(pattern)
		if err != nil {
			fmt.Printf("Warning: failed to glob pattern %s: %v\n", pattern, err)
			continue
		}
		allFiles = append(allFiles, files...)
	}

	if len(allFiles) == 0 {
		fmt.Printf("Warning: no template files found in %s\n", s.templatesDir)
		return
	}

	// Parse all templates together
	tmpl, err := template.New("").Funcs(funcMap).ParseFiles(allFiles...)
	if err != nil {
		fmt.Printf("Warning: failed to parse templates: %v\n", err)
		return
	}

	s.engine.SetHTMLTemplate(tmpl)
}

// setupRoutes configures all HTTP routes.
func (s *Server) setupRoutes() {
	// Static files
	s.engine.Static("/static", s.staticDir)

	// Adventure images (served from data directory)
	s.engine.GET("/play/:slug/images/*filepath", s.handleAdventureImages)

	// Global maps (served from data/maps)
	s.engine.GET("/maps/*filepath", s.handleMaps)

	// Main routes
	s.engine.GET("/", s.handleIndex)
	s.engine.GET("/adventures", s.handleAdventuresList)
	s.engine.POST("/adventures", s.handleCreateAdventure)
	s.engine.GET("/play/:slug", s.handleGame)
	s.engine.POST("/play/:slug/message", s.handleMessage)
	s.engine.GET("/play/:slug/stream", s.handleStream)
	s.engine.GET("/play/:slug/characters", s.handleCharacters)
	s.engine.GET("/play/:slug/character/:name", s.handleCharacterSheet)
	s.engine.GET("/play/:slug/info", s.handleAdventureInfo)
	s.engine.GET("/play/:slug/minimap", s.handleMinimap)

	// Character images (served from data/characters/)
	s.engine.GET("/characters/images/:filename", s.handleCharacterImages)

	// Gallery routes
	s.engine.GET("/play/:slug/gallery", s.handleGallery)
}

// handleAdventureImages serves images from adventure directories.
func (s *Server) handleAdventureImages(c *gin.Context) {
	slug := c.Param("slug")
	filePath := c.Param("filepath")

	// Construct path to adventure images
	imagePath := filepath.Join("data", "adventures", slug, "images", filePath)

	// Security check: ensure path doesn't escape
	cleanPath := filepath.Clean(imagePath)
	if !strings.HasPrefix(cleanPath, filepath.Join("data", "adventures", slug)) {
		c.Status(http.StatusForbidden)
		return
	}

	c.File(imagePath)
}

// handleCharacterImages serves character portrait images from data/characters/.
func (s *Server) handleCharacterImages(c *gin.Context) {
	filename := c.Param("filename")

	// Construct path to character image
	imagePath := filepath.Join("data", "characters", filename)

	// Security check: ensure path doesn't escape
	cleanPath := filepath.Clean(imagePath)
	if !strings.HasPrefix(cleanPath, filepath.Join("data", "characters")) {
		c.Status(http.StatusForbidden)
		return
	}

	c.File(imagePath)
}

// Run starts the server.
func (s *Server) Run() error {
	addr := fmt.Sprintf(":%d", s.port)
	fmt.Printf("Starting SkillsWeaver Web on http://localhost%s\n", addr)
	return s.engine.Run(addr)
}

// Stop gracefully stops the server.
func (s *Server) Stop() {
	s.sessionManager.Stop()
}

// renderError renders an error page.
func (s *Server) renderError(c *gin.Context, status int, message string) {
	c.HTML(status, "error.html", gin.H{
		"Title":   "Error",
		"Message": message,
	})
}

// SetupSSE sets up an SSE connection.
func SetupSSE(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Transfer-Encoding", "chunked")
	c.Header("X-Accel-Buffering", "no") // Disable nginx buffering
}

// WriteSSE writes an SSE event to the response.
func WriteSSE(w io.Writer, event SSEEvent) {
	fmt.Fprintf(w, "event: %s\n", event.Event)
	fmt.Fprintf(w, "data: %s\n\n", event.Data)
}
