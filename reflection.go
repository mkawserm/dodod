package dodod

import (
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/analysis/analyzer/simple"
	"github.com/blevesearch/bleve/mapping"
	"github.com/go-openapi/inflect"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func ExtractFields(document interface{}) map[string]string {
	data := make(map[string]string)
	extractFields(reflect.TypeOf(document), data)
	return data
}

func extractFields(t reflect.Type, data map[string]string) {
	if t.Kind() == reflect.Ptr {
		extractFields(t.Elem(), data)
	} else if t.Kind() == reflect.Struct {
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			tags := strings.Split(f.Tag.Get("json"), ",")
			if len(tags) > 0 {
				name := strings.TrimSpace(tags[0])
				if len(name) > 0 {
					data[name] = f.Type.Name()
				}
			}
		}
	}
}

func GetId(document interface{}) string {
	return getId(reflect.TypeOf(document), reflect.ValueOf(document))
}

func GetType(document interface{}) string {
	if d, ok := document.(Document); ok {
		return d.Type()
	} else {
		return ""
	}
}

func getId(t reflect.Type, v reflect.Value) string {
	if t.Kind() == reflect.Ptr {
		return getId(t.Elem(), v.Elem())
	} else if t.Kind() == reflect.Struct {
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			tags := strings.Split(f.Tag.Get("json"), ",")
			if len(tags) > 0 {
				name := strings.TrimSpace(tags[0])
				if name == "id" {
					v2 := v.FieldByName(f.Name)
					if v2.Kind() == reflect.String {
						return v2.String()
					}
				}
			}
		}
	}

	return ""
}

//func getDododType(t reflect.Type, v reflect.Value) string {
//	if t.Kind() == reflect.Ptr {
//		return getId(t.Elem(), v.Elem())
//	} else if t.Kind() == reflect.Struct {
//		for i := 0; i < t.NumField(); i++ {
//			f := t.Field(i)
//			tags := strings.Split(f.Tag.Get("json"),",")
//			if len(tags) > 0 {
//				name := strings.TrimSpace(tags[0])
//				if name == "dododType" {
//					v2 := v.FieldByName(f.Name)
//					if v2.Kind() == reflect.String {
//						return v2.String()
//					}
//				}
//			}
//		}
//	}
//
//	return ""
//}

func registerDocumentMapping(base interface{}, doc mapping.Classifier, docName ...string) (err error) {
	baseValue := reflect.ValueOf(base)
	//if !baseValue.CanInterface() {
	//	return ErrInvalidBase
	//}

	docValue := reflect.ValueOf(doc).Elem()
	//if !docValue.IsValid() {
	//	return ErrInvalidDoc
	//}

	//if docValue.Kind() != reflect.Struct {
	//	return ErrInvalidDocNotStruct
	//}

	docType := docValue.Type()
	docMapping := bleve.NewDocumentMapping()
	docMapping.DefaultAnalyzer = simple.Name

	for i := 0; i < docType.NumField(); i++ {
		field := docType.Field(i)
		bleveTag := field.Tag.Get(`bleve`)
		jsonTag := field.Tag.Get(`json`)
		bleveTags := strings.Split(bleveTag, `,`)
		jsonTags := strings.Split(jsonTag, `,`)

		name := field.Name
		if jsonTag != `-` && jsonTags[0] != `` {
			name = jsonTags[0]
		} else if bleveTag != `-` && bleveTags[0] != `` {
			name = bleveTags[0]
		}

		if bleveTag == `-` {
			disable := bleve.NewDocumentMapping()
			disable.Enabled = false
			docMapping.AddSubDocumentMapping(name, disable)
			continue
		}

		fieldMap := bleve.NewTextFieldMapping()
		k := parseFieldKind(field.Type)
		switch k {
		case reflect.Int:
			fieldMap = bleve.NewNumericFieldMapping()
		case reflect.Bool:
			fieldMap = bleve.NewBooleanFieldMapping()
		case reflect.Struct:
			f := docValue.FieldByName(field.Name)
			if f.CanInterface() {
				fi := f.Interface()
				switch fi.(type) {
				case time.Time:
					fieldMap = bleve.NewDateTimeFieldMapping()
				default:
					if _, ok := fi.(mapping.Classifier); ok {
						var d interface{}
						if f.Type().Kind() == reflect.Ptr {
							d = reflect.New(f.Type().Elem()).Interface()
						} else {
							d = reflect.New(f.Type()).Interface()
						}
						if err = registerDocumentMapping(docMapping, d.(mapping.Classifier), name); err != nil {
							return
						}
						continue
					}
				}
			}
		case reflect.Array, reflect.Map, reflect.Slice:
			continue
		}

		fieldMap.Name = name
		mapValue := reflect.ValueOf(fieldMap).Elem()

		if len(bleveTags) > 1 {
			for _, v := range bleveTags[1:] {
				kv := strings.Split(v, `:`)
				if len(kv) == 2 {
					key := inflect.Camelize(kv[0])
					f := mapValue.FieldByName(key)
					if f.IsValid() && f.CanSet() {
						switch f.Kind() {
						case reflect.Bool:
							b, err := strconv.ParseBool(kv[1])
							if err != nil {
								return ErrNonBooleanValueForBooleanField
							}
							f.SetBool(b)
						case reflect.String:
							f.SetString(kv[1])
							//default:
							//	return ErrUnknownMappingField
						}
					}
				}
			}
		}

		docMapping.AddFieldMappingsAt(field.Name, fieldMap)
	}

	switch baseValue.Interface().(type) {
	case *mapping.IndexMappingImpl:
		b := base.(*mapping.IndexMappingImpl)
		b.AddDocumentMapping(doc.Type(), docMapping)
	case *mapping.DocumentMapping:
		b := base.(*mapping.DocumentMapping)
		if len(docName) > 0 {
			b.AddSubDocumentMapping(docName[0], docMapping)
		}
		//else {
		//	b.AddSubDocumentMapping(doc.Type(), docMapping)
		//}
	default:
		return ErrUnknownBaseType
	}

	return
}

func parseFieldKind(t reflect.Type) reflect.Kind {
	switch t.Kind() {
	case reflect.Ptr:
		return parseFieldKind(t.Elem())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16,
		reflect.Uint32, reflect.Uint64, reflect.Float32,
		reflect.Float64:
		return reflect.Int
	default:
		return t.Kind()
	}
}
