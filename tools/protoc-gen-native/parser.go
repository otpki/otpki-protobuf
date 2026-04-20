package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/samber/oops"
	"golang.org/x/tools/go/packages"
	"google.golang.org/protobuf/compiler/protogen"
)

const converterPackageName = "converters"

// converterPackagePath will be set from the local directory path or Go import path provided as a parameter
// cachedPackage stores the loaded package for reuse
var (
	converterPackagePath string            // will be set from a protoc parameter
	cachedPkg            *packages.Package // if we've already parsed the package, it will be here
	cachedFset           *token.FileSet    // set of files that we've parsed
)

// setConverterPackage sets the path for the converter package
func setConverterPackage(packagePath string) error {
	converterPackagePath = packagePath

	// cache the package
	if err := loadPackage(); err != nil {
		return err
	}

	return nil
}

// loadPackage loads the converter package
func loadPackage() error {
	if converterPackagePath == "" {
		return oops.Errorf("converter package path not set")
	}
	// already loaded
	if cachedPkg != nil {
		return nil
	}

	cachedFset = token.NewFileSet()
	cfg := &packages.Config{
		Mode: packages.NeedSyntax | packages.NeedFiles | packages.NeedName | packages.NeedImports,
		Fset: cachedFset,
	}

	loadPattern := converterPackagePath
	if isExistingDirectory(converterPackagePath) {
		cfg.Dir = converterPackagePath
		loadPattern = "."
	}

	// load package
	pkgs, err := packages.Load(cfg, loadPattern)
	if err != nil {
		return oops.Errorf("error loading package: %w", err)
	}
	if len(pkgs) == 0 {
		return oops.Errorf("no packages found at %s", converterPackagePath)
	}
	//if packages.PrintErrors(pkgs) > 0 {
	//	return oops.Errorf("encountered errors loading package %s", converterPackagePath)
	//}

	cachedPkg = pkgs[0] // use the first package name
	return nil
}

func isExistingDirectory(path string) bool {
	info, err := os.Stat(path)
	if err == nil {
		return info.IsDir()
	}

	if filepath.IsAbs(path) {
		return false
	}

	wd, wdErr := os.Getwd()
	if wdErr != nil {
		return false
	}

	info, err = os.Stat(filepath.Join(wd, path))
	return err == nil && info.IsDir()
}

// searchPackageForFunctions searches all files in the cached package for functions matching a search term
func searchPackageForFunctions(searchTerm string) []*ast.FuncDecl {
	if err := loadPackage(); err != nil {
		fmt.Fprintf(os.Stderr, "Error loading package: %v\n", err)
		return nil
	}

	var results []*ast.FuncDecl

	// search all files in the package
	for _, file := range cachedPkg.Syntax {
		for _, decl := range file.Decls {
			if funcDecl, ok := decl.(*ast.FuncDecl); ok {
				if strings.Contains(funcDecl.Name.Name, searchTerm) {
					results = append(results, funcDecl)
				}
			}
		}
	}

	return results
}

// getFunction returns the first function that matches the given name from any file in the package
func getFunction(functionName string) (*ast.FuncDecl, *token.FileSet, error) {
	if err := loadPackage(); err != nil {
		return nil, nil, err
	}

	// search all files in the package for the function
	for _, file := range cachedPkg.Syntax {
		for _, decl := range file.Decls {
			if funcDecl, ok := decl.(*ast.FuncDecl); ok {
				if funcDecl.Name.Name == functionName {
					return funcDecl, cachedFset, nil
				}
			}
		}
	}

	return nil, nil, oops.Errorf("function %s not found", functionName)
}

// getFunctionImports returns the import paths that are actually used by a function
// if the function doesn't exist, it returns an empty slice
func getFunctionImports(functionName string) []string {
	// get the function and the file it's in
	funcDecl, fset, err := getFunction(functionName)
	if err != nil {
		return []string{}
	}

	// get the file containing the function
	filename := fset.Position(funcDecl.Pos()).Filename

	// find the file in the package
	var fileNode *ast.File
	for _, file := range cachedPkg.Syntax {
		pos := fset.Position(file.Pos())
		if pos.Filename == filename {
			fileNode = file
			break
		}
	}

	if fileNode == nil {
		return []string{}
	}

	// map to store import path by package name
	importsByName := make(map[string]string)
	for _, imp := range fileNode.Imports {
		// extract the import path from the quoted string
		importPath := strings.Trim(imp.Path.Value, `"`)

		// get the package alias/name
		pkgName := ""
		if imp.Name != nil {
			// if import uses an alias: import alias "path"
			pkgName = imp.Name.Name
		} else {
			// default case: use last part of path as package name
			parts := strings.Split(importPath, "/")
			pkgName = parts[len(parts)-1]
		}
		importsByName[pkgName] = importPath
	}

	// track which imports are actually used
	usedImports := make(map[string]bool)

	// visit all nodes in the function to find import usages
	ast.Inspect(funcDecl, func(n ast.Node) bool {
		// look for selector expressions like pkg.Function() or pkg.Type
		if sel, ok := n.(*ast.SelectorExpr); ok {
			if ident, ok := sel.X.(*ast.Ident); ok {
				// Check if this identifier corresponds to an import
				if importPath, exists := importsByName[ident.Name]; exists {
					usedImports[importPath] = true
				}
			}
		}
		return true // continue traversing
	})

	// convert the map to a slice
	result := make([]string, 0, len(usedImports))
	for importPath := range usedImports {
		result = append(result, importPath)
	}

	return result
}

// getFunctionDependencies recursively finds all functions that are called by the given function
// returns a map in place of a hash set
func getFunctionDependencies(functionName string) map[string]bool {
	dependencies := map[string]bool{}
	// map to track visited functions to avoid infinite recursion
	visited := map[string]bool{}

	// recursive function to find dependencies
	var findDeps func(string)

	findDeps = func(funcName string) {
		// stop if we've already visited this function
		if visited[funcName] {
			return
		}
		visited[funcName] = true

		funcDecl, _, err := getFunction(funcName)
		if err != nil {
			return
		}

		// visit all AST nodes and look for function calls
		ast.Inspect(funcDecl, func(n ast.Node) bool {
			// look for function calls
			if call, ok := n.(*ast.CallExpr); ok {
				// check if it's a simple function call like 'convertSomeTypeToProto'
				if ident, ok := call.Fun.(*ast.Ident); ok {
					calledFunc := ident.Name

					// check if this function exists in our package
					if functionExists(calledFunc) {
						// add the dependency
						dependencies[calledFunc] = true
						// recursively find its dependencies
						findDeps(calledFunc)
					}
				}

				// also check for selector expressions like 'pkg.Function()'
				if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
					// we need to check if this is a local package function call
					if ident, ok := sel.X.(*ast.Ident); ok {
						// if the identifier is the current package name, it's a local call
						calledFunc := sel.Sel.Name

						// check if the identifier is likely referring to our package or a similar one
						isLocalCall := ident.Name == cachedPkg.Name || ident.Name == converterPackageName

						// check if it's a local call and the function exists in our package
						if isLocalCall && functionExists(calledFunc) {
							// add the dependency
							dependencies[calledFunc] = true
							// recursively find its dependencies
							findDeps(calledFunc)
						}
					}
				}
			}
			return true // continue traversing the AST
		})
	}

	// start the recursive search from the given function
	findDeps(functionName)

	// remove the original function from dependencies since we'll already include it
	delete(dependencies, functionName)

	return dependencies
}

// getFunctionString returns the string representation of a function
// if the function doesn't exist, it returns an empty string
func getFunctionString(g *protogen.GeneratedFile, functionName string) string {
	funcDecl, fset, err := getFunction(functionName)
	if err != nil {
		return ""
	}

	var buf bytes.Buffer
	err = format.Node(&buf, fset, funcDecl)
	if err != nil {
		return ""
	}

	// get the function string representation
	content := buf.String()

	// see what imports are used in the function
	funcImports := getFunctionImports(functionName)
	for _, importPath := range funcImports {
		// attempt to import the package into the generated file
		// this will give us the identifier to use within the function
		ident := g.QualifiedGoIdent(protogen.GoIdent{
			GoImportPath: protogen.GoImportPath(importPath),
		})

		// now get the original alias used in our imported converters package
		// we need to find the import statement in one of the files to get its alias
		originalAlias := findOriginalImportAlias(importPath)

		if originalAlias != "" && originalAlias+"." != ident {
			// replace all non-matching aliases with the generated alias
			content = strings.ReplaceAll(content, originalAlias+".", ident)
		}
	}
	content = strings.ReplaceAll(content, "Convert", "convert")
	return content
}

// findOriginalImportAlias finds the original alias used for a package in the provided converters package
func findOriginalImportAlias(importPath string) string {
	if err := loadPackage(); err != nil {
		return ""
	}

	// for each file
	for _, syntax := range cachedPkg.Syntax {
		// for each import
		for _, imp := range syntax.Imports {
			// remove the surrounding ""
			path := strings.Trim(imp.Path.Value, `"`)
			// if we've found the import we're looking for
			if path == importPath {
				// if it has an alias
				if imp.Name != nil {
					return imp.Name.Name
				} else {
					// default import is last part of path
					parts := strings.Split(path, "/")
					return parts[len(parts)-1]
				}
			}
		}
	}

	return ""
}

// functionExists checks if a function exists in the converters package
func functionExists(functionName string) bool {
	_, _, err := getFunction(functionName)
	return err == nil
}

// getFunctionStringSimple returns a simpler version of getFunctionString without alias mapping
// This is useful for tests and basic functionality checks
func getFunctionStringSimple(functionName string) string {
	funcDecl, fset, err := getFunction(functionName)
	if err != nil {
		return ""
	}

	var buf bytes.Buffer
	err = format.Node(&buf, fset, funcDecl)
	if err != nil {
		return ""
	}
	return buf.String()
}

// listConverterFunctions returns a list of all converter functions in the converters package
func listConverterFunctions() []string {
	if err := loadPackage(); err != nil {
		fmt.Fprintf(os.Stderr, "Error loading package: %v\n", err)
		return nil
	}

	funcs := searchPackageForFunctions("convert")
	if len(funcs) == 0 {
		return nil
	}

	// Extract just the function names
	var result []string
	for _, f := range funcs {
		result = append(result, f.Name.Name)
	}

	return result
}
