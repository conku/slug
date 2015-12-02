package slug

import (
	"database/sql/driver"
	"os"
	"path"
	"strings"

	"github.com/qor/qor"
	"github.com/qor/qor/admin"
	"github.com/qor/qor/resource"
	"github.com/qor/qor/utils"
	"github.com/qor/qor/validations"
)

type Slug struct {
	Slug string
}

func (slug *Slug) Scan(value interface{}) error {
	if bytes, ok := value.([]byte); ok {
		slug.Slug = string(bytes)
	} else if str, ok := value.(string); ok {
		slug.Slug = str
	} else if strs, ok := value.([]string); ok {
		slug.Slug = strs[0]
	}
	return nil
}

func (slug Slug) Value() (driver.Value, error) {
	return slug.Slug, nil
}

var injected bool

func (Slug) ConfigureQorMeta(meta resource.Metaor) {
	if meta, ok := meta.(*admin.Meta); ok {
		res := meta.GetBaseResource().(*admin.Resource)

		for _, gopath := range strings.Split(os.Getenv("GOPATH"), ":") {
			admin.RegisterViewPath(path.Join(gopath, "src/github.com/qor/slug/views"))
		}
		res.UseTheme("slug")

		name := strings.TrimSuffix(meta.Name, "WithSlug")
		if meta := res.GetMeta(name); meta != nil {
			meta.Type = "slug"
		} else {
			res.Meta(&admin.Meta{Name: name, Type: "slug"})
		}

		var fieldName = meta.Name
		res.AddValidator(func(record interface{}, metaValues *resource.MetaValues, context *qor.Context) error {
			if meta := metaValues.Get(fieldName); meta != nil {
				slug := utils.ToString(metaValues.Get(fieldName).Value)
				if slug == "" {
					return validations.NewError(record, fieldName, name+"'s slug can't be blank")
				} else if strings.Contains(slug, " ") {
					return validations.NewError(record, fieldName, name+"'s slug can't contains blank string")
				}
			} else {
				if field, ok := context.GetDB().NewScope(record).FieldByName(meta.Name); ok && field.IsBlank {
					return validations.NewError(record, fieldName, name+"'s slug can't be blank")
				}
			}
			return nil
		})

		res.IndexAttrs(append(res.IndexAttrs(), "-"+meta.Name)...)
		res.ShowAttrs(append(res.ShowAttrs(), "-"+meta.Name)...)
		res.EditAttrs(append(res.EditAttrs(), "-"+meta.Name)...)
		res.NewAttrs(append(res.NewAttrs(), "-"+meta.Name)...)
	}
}
