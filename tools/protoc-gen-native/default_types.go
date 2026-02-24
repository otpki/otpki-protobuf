package main

import (
	"fmt"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func getGoTypeForField(g *protogen.GeneratedFile, field *protogen.Field) string {
	if field.Desc.IsMap() {
		key, val := getGoTypeForMap(g, field)
		return fmt.Sprintf("map[%s]%s", key, val)
	}
	baseType := getBaseGoType(g, field)

	if field.Desc.IsList() {
		return "[]" + baseType
	}
	// "bytes" is not a list, but it is equivalent to a slice of bytes (which is a list), meaning we do not
	// need to make a pointer to it.
	if isNilable(field) && isScalar(field.Desc.Kind()) && field.Desc.Kind() != protoreflect.BytesKind {
		return "*" + baseType
	}
	return baseType
}

func getBaseGoType(g *protogen.GeneratedFile, field *protogen.Field) string {
	return getGoTypeForKind(g, field, field.Desc.Kind())
}

func getGoTypeForMap(g *protogen.GeneratedFile, field *protogen.Field) (key string, value string) {
	key = getGoTypeForKind(g, field, field.Desc.MapKey().Kind())
	value = getGoTypeForKind(g, field, field.Desc.MapValue().Kind())
	return
}

func getGoTypeForKind(g *protogen.GeneratedFile, field *protogen.Field, kind protoreflect.Kind) string {
	switch kind {
	case protoreflect.BoolKind:
		return "bool"
	case protoreflect.StringKind:
		return "string"
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		return "int32"
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		return "uint32"
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return "int64"
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return "uint64"
	case protoreflect.FloatKind:
		return "float32"
	case protoreflect.DoubleKind:
		return "float64"
	case protoreflect.BytesKind:
		return "[]byte"
	case protoreflect.EnumKind:
		return g.QualifiedGoIdent(field.Enum.GoIdent)
	case protoreflect.MessageKind:
		return "*" + g.QualifiedGoIdent(field.Message.GoIdent)
	default:
		return "any"
	}
}

func isNilable(field *protogen.Field) bool {
	return field.Desc.HasPresence()
}

func isScalar(kind protoreflect.Kind) bool {
	switch kind {
	case protoreflect.BoolKind, protoreflect.StringKind, protoreflect.Int32Kind, protoreflect.Uint32Kind,
		protoreflect.Int64Kind, protoreflect.Uint64Kind, protoreflect.FloatKind, protoreflect.DoubleKind,
		protoreflect.BytesKind, protoreflect.EnumKind:
		return true
	}
	return false
}
