package ui

import (
	"bytes"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"mime"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/evanw/esbuild/pkg/api"
	"github.com/llmite-ai/mirra/internal/logger"
)

//go:embed static/*
var staticContent embed.FS

//go:embed dist/*
var dist embed.FS

type Manager struct {
	env string
}

type Option func(*Manager)

func NewManager(opts ...Option) *Manager {
	return &Manager{
		env: os.Getenv("NODE_ENV"),
	}
}

func (m *Manager) Static(root, remove string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := logger.NewDefaultLogger()
		trimmedPath := strings.TrimPrefix(r.URL.Path, remove)

		var (
			data []byte
			path string
			err  error
		)

		if trimmedPath == "" {
			trimmedPath = "index.html"
		}

		if m.env != "production" {
			// Get the current working directory
			workingDir, err := os.Getwd()
			if err != nil {
				fmt.Println("Error getting working directory:", err)
				panic(err)
			}

			path = filepath.Join(workingDir, root, trimmedPath)

			log.Debug("[assets] requested file", "path", path)

			// Read the file content from the os file system.
			data, err = os.ReadFile(path)
			if err == nil {
				log.Debug("[assets] os file", "path", path)
			}
		}

		if data == nil {
			path := filepath.Join("static", filepath.Clean("/"+trimmedPath))
			if strings.HasSuffix(r.URL.Path, "/") {
				http.Error(w, "Not Found", http.StatusNotFound)
				return
			}

			log.Debug("[assets] embedded file", "path", path)

			// Read the file content from the embedded file system.
			data, err = fs.ReadFile(staticContent, path)
			if err != nil {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
		}

		// Determine the content type of the file.
		contentType := mime.TypeByExtension(filepath.Ext(path))
		if contentType == "" {
			contentType = "application/octet-stream"
		}

		// Set the Content-Type header.
		w.Header().Set("Content-Type", contentType)
		w.Write(data)
	}
}

func (m *Manager) BuildAssets() error {
	buildOptions, err := m.buildOptions()
	if err != nil {
		slog.Error("[build] Failed to create build options", "error", err)
		return err
	}

	err = m.build(buildOptions)
	if err != nil {
		slog.Error("[build] Failed to build package", "error", err)
		return err
	}

	return nil
}

func (m *Manager) SrcHandler(root string) http.HandlerFunc {
	buildOptions, err := m.buildOptions()
	if err != nil {
		slog.Error("[build] Failed to build package", "error", err)
		panic(err)
	}

	responseWithEmbedded := func(w http.ResponseWriter, r *http.Request) {
		log := logger.NewDefaultLogger()
		urlPath := r.URL.Path
		requestPath := strings.TrimPrefix(urlPath, root)
		filePath := filepath.Join("dist", requestPath)

		log.Info("Serving embedded file", "path", r.URL.Path, "filename", filePath)

		file, err := dist.Open(filePath)
		if err != nil {
			log.Error("Failed to open embedded file", "path", filePath, "error", err)
			http.NotFound(w, r)
			return
		}
		defer file.Close()

		contentType := mime.TypeByExtension(filepath.Ext(requestPath))
		if contentType == "" {
			contentType = "application/octet-stream"
		}

		w.Header().Set("Content-Type", contentType)

		_, copyErr := io.Copy(w, file)
		if copyErr != nil {
			log.Error("Failed to serve embedded file", "path", filePath, "error", copyErr)
		} else {
			log.Info("Served embedded file", "path", filePath)
		}
	}

	return func(w http.ResponseWriter, r *http.Request) {
		log := logger.NewDefaultLogger()

		log.Info("Serving file", "path", r.URL.Path)

		if m.env == "production" {
			responseWithEmbedded(w, r)
			return
		}

		urlPath := r.URL.Path
		requestPath := strings.TrimPrefix(urlPath, root)

		if requestPath == "" || requestPath == "/" {
			m.index(dist, w, r)
			return
		}

		now := time.Now()
		err := m.buildAndServerFromESBuild(buildOptions, requestPath, w, r)
		log.Info("Built package", "filename", requestPath, "duration", time.Since(now))
		if err != nil {
			log.Error("Error building package", "filename", requestPath, "error", err)
			err = fmt.Errorf("failed to build %s: %v", requestPath, err)

			w.Header().Set("Content-Type", "application/javascript")
			w.Write([]byte(buildErrorScript(err)))
		}

		return
	}
}

func (m *Manager) buildOptions() (api.BuildOptions, error) {
	// Get the current working directory
	workingDir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting working directory:", err)
		return api.BuildOptions{}, err
	}

	postcssPath := filepath.Join(workingDir, "internal/ui/src/node_modules/.bin/postcss")

	buildOptions := api.BuildOptions{
		Outdir: filepath.Join(workingDir, "internal/ui/dist"),
		EntryPoints: []string{
			filepath.Join(workingDir, "internal/ui/src/style/index.css"),
			filepath.Join(workingDir, "internal/ui/src/index.tsx"),
		},
		Platform:     api.PlatformBrowser,
		Bundle:       true,
		MinifySyntax: true,
		Sourcemap:    api.SourceMapLinked,
		Loader: map[string]api.Loader{
			".tsx":   api.LoaderTSX,
			".ts":    api.LoaderTS,
			".css":   api.LoaderCSS,
			".ttf":   api.LoaderText,
			".woff2": api.LoaderText,
			".svg":   api.LoaderText,
		},
		Define: map[string]string{
			"process.env.APP_HOST": `"` + os.Getenv("APP_HOST") + `"`,
			"process.env.NODE_ENV": `"` + os.Getenv("NODE_ENV") + `"`,
		},
		Plugins: []api.Plugin{
			{
				Name: "postcss",
				Setup: func(build api.PluginBuild) {
					build.OnLoad(api.OnLoadOptions{Filter: `\.css$`}, func(args api.OnLoadArgs) (api.OnLoadResult, error) {
						content, err := os.ReadFile(args.Path)
						if err != nil {
							return api.OnLoadResult{}, err
						}

						cmd := exec.Command(postcssPath, args.Path)
						cmd.Dir = "internal/ui/src"
						cmd.Stdin = bytes.NewReader(content)
						cmd.Stderr = os.Stderr
						out, err := cmd.Output()
						if err != nil {
							return api.OnLoadResult{}, err
						}

						outString := string(out)

						return api.OnLoadResult{
							Contents: &outString,
							Loader:   api.LoaderCSS,
						}, nil
					})
				},
			},
		},
		Write: true,
	}

	return buildOptions, nil
}

func (m *Manager) build(buildOptions api.BuildOptions) error {
	result := api.Build(buildOptions)
	if len(result.Errors) != 0 {
		return fmt.Errorf("failed to build package: %v", result.Errors)
	}
	return nil
}

func (m *Manager) buildAndServerFromESBuild(
	buildOptions api.BuildOptions,
	requestPath string,
	w http.ResponseWriter,
	_ *http.Request,
) error {
	result := api.Build(buildOptions)
	if len(result.Errors) != 0 {
		// TODO: make the error better.
		// - there much more information in the error
		// - could make a custom error type that includes that information?
		return fmt.Errorf("failed to build package: %v", result.Errors)
	}

	// Determine the content type of the file.
	contentType := mime.TypeByExtension(filepath.Ext(requestPath))
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Set the Content-Type header.
	w.Header().Set("Content-Type", contentType)

	existingFiles := []string{}
	for _, outputFile := range result.OutputFiles {
		relativePath := strings.TrimPrefix(outputFile.Path, buildOptions.Outdir)
		if strings.HasSuffix(relativePath, requestPath) {
			w.Write(outputFile.Contents)
			return nil
		}
		existingFiles = append(existingFiles, outputFile.Path)
	}

	return fmt.Errorf("file not found: %s. Existing files: %v", requestPath, existingFiles)
}

func buildErrorScript(err error) string {
	return fmt.Sprintf(`window.addEventListener('DOMContentLoaded', () => {
  document.body.innerHTML += '<div style="color: red;">%s</div>';
});
		`, err.Error())
}

func (m *Manager) index(efs fs.FS, w http.ResponseWriter, _ *http.Request) {
}

func (m *Manager) list(efs fs.FS, w http.ResponseWriter, _ *http.Request) {
	log := logger.NewDefaultLogger()
	files, err := listEmbeddedFiles(efs)
	if err != nil {
		log.Error("[assets] Failed to list embedded files", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte("<html><body><ul>"))
	for _, file := range files {
		w.Write([]byte("<li><a href=\"" + file + "\">" + file + "</a></li>"))
	}
	w.Write([]byte("</ul></body></html>"))
}

func listEmbeddedFiles(efs fs.FS) ([]string, error) {
	var files []string
	err := fs.WalkDir(efs, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		files = append(files, path)
		return nil
	})

	return files, err
}

func getFilename(root, rawurl string) *string {
	// remove root from rawurl
	rawurl = strings.TrimPrefix(rawurl, root)

	parsedURL, err := url.Parse(rawurl)
	if err != nil {
		return nil // or handle the error as you prefer
	}

	filename := path.Base(parsedURL.Path)
	if filename == "/" || filename == "." {
		return nil
	}

	return &filename
}
