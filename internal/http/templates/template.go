/*
SPDX-License-Identifier: GPL-3.0-or-later

Copyright (C) 2025 Aaron Mathis aaron.mathis@gmail.com

This file is part of GoSight.

GoSight is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

GoSight is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with GoSight. If not, see https://www.gnu.org/licenses/.
*/

// server/internal/http/template.go
// Handle loading of template files

package templates

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/aaronlmathis/gosight/server/internal/config"
	"github.com/aaronlmathis/gosight/server/internal/store"
	"github.com/aaronlmathis/gosight/server/internal/store/userstore"
	"github.com/aaronlmathis/gosight/shared/utils"
)

var Tmpl *template.Template

type TemplateData struct {
	Title       string              // Title of the page
	Body        string              // Body of the page
	MetricStore store.MetricStore   // Metric store for the page
	MetricIndex *store.MetricIndex  // Metric index for the page
	UserStore   userstore.UserStore // User store for the page

}

func InitTemplates(cfg *config.Config, funcMap template.FuncMap) error {
	Tmpl = template.New("").Funcs(funcMap) // Keep the initial name
	counter := 0

	err := filepath.Walk(cfg.Web.TemplateDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".html") {
			return nil
		}

		relativePath, err := filepath.Rel(cfg.Web.TemplateDir, path)
		if err != nil {
			utils.Debug("error getting relative path %v: %v", path, err)
			return err
		}

		templateName := strings.TrimSuffix(filepath.ToSlash(relativePath), ".html")

		_, err = Tmpl.New(templateName).Funcs(funcMap).ParseFiles(path)
		if err != nil {
			utils.Debug("error parsing template %v: %v", path, err)
		}

		counter++
		utils.Debug("📦 Template loaded: %v - %v - %d", path, templateName, counter)

		return nil
	})

	if err != nil {
		utils.Debug("error walking the path %v: %v", cfg.Web.TemplateDir, err)
		return err
	}

	utils.Debug("📦 Total Templates loaded: %d", counter)
	return nil
}

func RenderTemplate(w http.ResponseWriter, layout string, data any) error {

	utils.Debug("Rendering template: %s", layout)

	err := Tmpl.ExecuteTemplate(w, layout, data)
	if err != nil {
		utils.Debug("🔴 ExecuteTemplate failed: %v", err)
	}
	return err
}
