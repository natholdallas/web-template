package swg

import (
	"encoding/json"
	"reflect"
	"strings"

	"webtplmst/docs"
	"webtplmst/internal/conf"

	"github.com/gofiber/fiber/v3"
	"github.com/yokeTH/gofiber-scalar/scalar/v3"
)

// tagPrefixes are the per-app tag prefixes (lowercase), stripped in the UI.
var tagPrefixes = []string{"usr", "adm"}

// schemaPrefixes are the per-app schema definition prefixes (capitalized),
// stripped in the UI and rewritten in every $ref.
var schemaPrefixes = []string{"Usr", "Adm"}

func stripAny(t string, prefixes []string) string {
	for _, p := range prefixes {
		if strings.HasPrefix(t, p) && len(t) > len(p) {
			return t[len(p):]
		}
	}
	return t
}

// displayDoc returns the swagger spec with per-app tag AND schema prefixes
// stripped, so the docs UI shows clean group/model names (`User`, `Admin`, ...)
// while the on-disk docs/swagger.json keeps the prefixed names for codegen.
func displayDoc() string {
	doc := docs.SwaggerInfo.ReadDoc()
	var v map[string]any
	if err := json.Unmarshal([]byte(doc), &v); err != nil {
		return doc
	}

	// Rewrite every $ref string: #/definitions/AdmUserIn -> #/definitions/UserIn
	var rewriteRefs func(node any)
	rewriteRefs = func(node any) {
		switch n := node.(type) {
		case map[string]any:
			if ref, ok := n["$ref"].(string); ok && strings.HasPrefix(ref, "#/definitions/") {
				name := ref[strings.LastIndex(ref, "/")+1:]
				if stripped := stripAny(name, schemaPrefixes); stripped != name {
					n["$ref"] = "#/definitions/" + stripped
				}
			}
			for _, val := range n {
				rewriteRefs(val)
			}
		case []any:
			for _, val := range n {
				rewriteRefs(val)
			}
		}
	}

	// Rename definition keys (strip schema prefix) before refs are rewritten.
	if defs, ok := v["definitions"].(map[string]any); ok {
		renamed := map[string]any{}
		for name, def := range defs {
			stripped := stripAny(name, schemaPrefixes)
			if existing, dup := renamed[stripped]; dup && !reflect.DeepEqual(existing, def) {
				// two different prefixed schemas collapse onto the same clean
				// name (e.g. usr & adm both have an `Auth`). Keep the prefixed
				// name to avoid losing one in the merged UI doc.
				renamed[name] = def
				continue
			}
			renamed[stripped] = def
		}
		v["definitions"] = renamed
		rewriteRefs(v)
	}

	// Strip tag prefixes on each operation.
	if paths, ok := v["paths"].(map[string]any); ok {
		for _, item := range paths {
			ops, ok := item.(map[string]any)
			if !ok {
				continue
			}
			for _, op := range ops {
				om, ok := op.(map[string]any)
				if !ok {
					continue
				}
				if tags, ok := om["tags"].([]any); ok {
					stripped := make([]any, 0, len(tags))
					for _, t := range tags {
						if s, ok := t.(string); ok {
							stripped = append(stripped, stripAny(s, tagPrefixes))
						}
					}
					om["tags"] = stripped
				}
			}
		}
	}

	out, err := json.Marshal(v)
	if err != nil {
		return doc
	}
	return string(out)
}

func Setup(app fiber.Router) {
	if !conf.App.Swagger {
		return
	}
	app.Get("/*", conf.App.SwaggerMiddleware, scalar.New(scalar.Config{
		Theme:             scalar.ThemeSaturn,
		FileContentString: displayDoc(),
		Title:             "API Documentation",
	}))
}
