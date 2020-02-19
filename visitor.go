package main

import (
	"go/ast"
	"go/token"
	"reflect"
	"strings"
)

type visitor struct {
	envs *[]string
}

func (v visitor) Visit(node ast.Node) (w ast.Visitor) {
	if node == nil {
		return v
	}

	if d := getDecl(node); d != nil {
		if t := getType(d); t != nil {
			if s := getStruct(t); s != nil {
				for _, field := range s.Fields.List {
					if field != nil && field.Tag != nil {
						if tag := parseTag(field.Tag); len(tag) > 0 {
							if env := getEnv(tag); len(env) > 0 {
								*v.envs = append(*v.envs, env)
							}
						}
					}
				}
			}
		}
	}

	return v
}

func getDecl(node ast.Node) *ast.GenDecl {
	if n, ok := node.(*ast.GenDecl); ok && n.Tok == token.TYPE {
		return n
	}
	return nil
}

func getType(decl *ast.GenDecl) *ast.TypeSpec {
	if t, ok := decl.Specs[0].(*ast.TypeSpec); ok {
		return t
	}
	return nil
}

func getStruct(t *ast.TypeSpec) *ast.StructType {
	if s, ok := t.Type.(*ast.StructType); ok {
		return s
	}
	return nil
}

func parseTag(raw *ast.BasicLit) string {
	return strings.TrimPrefix(
		strings.TrimSuffix(
			raw.Value,
			"`"),
		"`")
}

func getEnv(tag string) string {
	if env, ok := reflect.StructTag(tag).Lookup("env"); ok {
		return strings.Split(env, ",")[0]
	}
	return ""
}
