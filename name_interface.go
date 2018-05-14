package blueprint

import (
	"fmt"
	"sort"
)

type ModuleGroup struct {
	*moduleGroup
}

func (h *ModuleGroup) String() string {
	return h.moduleGroup.name
}

type Namespace interface {
	namespace(Namespace)
}
type NamespaceMarker struct {
}

func (m *NamespaceMarker) namespace(Namespace) {
}

// NameInterface tells how to locate modules by
// name. There should only be one name interface
// per Context, but potentially many namespaces
type NameInterface interface {
	// NewModule gets called when a new module is created
	NewModule(ctx NamespaceContext, group ModuleGroup, module Module) (namespace Namespace, err []error)

	// ModuleFromName finds the module with the
	// given name
	ModuleFromName(moduleName string, namespace Namespace) (group ModuleGroup, found bool)

	// MissingDependencyError returns an error
	// indicating that the given module could not be
	// found. The error contains some diagnostic
	// information about where the dependency can be
	// found.
	MissingDependencyError(depender string, dependerNamespace Namespace, depName string) (err error)

	// Rename
	Rename(oldName string, newName string, namespace Namespace) []error

	// AllModules returns all modules in a
	// deterministic order.
	AllModules() []ModuleGroup

	// GetNamespace gets the namespace for a given
	// path
	GetNamespace(ctx NamespaceContext) (namespace Namespace)

	// UniqueName returns a deterministic, unique,
	// arbitrary string for the given name in the
	// given namespace
	UniqueName(ctx NamespaceContext, name string) (unique string)
}

// NamespaceContext stores the information given
// to a NameInterface to enable the NameInterface
// to choose the namespace for any given module
type NamespaceContext interface {
	ModulePath() string
}

type namespaceContextImpl struct {
	modulePath string
}

func newNamespaceContext(moduleInfo *moduleInfo) (ctx NamespaceContext) {
	return &namespaceContextImpl{moduleInfo.pos.Filename}
}

func (ctx *namespaceContextImpl) ModulePath() string {
	return ctx.modulePath
}

// SimpleNameInterface just stores all modules in a
// map based on name
type SimpleNameInterface struct {
	modules map[string]ModuleGroup
}

func NewSimpleNameInterface() *SimpleNameInterface {
	return &SimpleNameInterface{
		modules: make(map[string]ModuleGroup),
	}
}

func (s *SimpleNameInterface) NewModule(ctx NamespaceContext, group ModuleGroup, module Module) (namespace Namespace, err []error) {
	name := group.name
	if group, present := s.modules[name]; present {
		return nil, []error{
			// seven characters at the start of the second line to align with the string "error: "
			fmt.Errorf("module %q already defined\n"+
				"       %s <-- previous definition here", name, group.modules[0].pos),
		}
	}

	s.modules[name] = group

	return nil, []error{}
}

func (s *SimpleNameInterface) ModuleFromName(moduleName string, namespace Namespace) (group ModuleGroup, found bool) {
	group, found = s.modules[moduleName]
	return group, found
}

func (s *SimpleNameInterface) Rename(oldName string, newName string, namespace Namespace) (errs []error) {
	existingGroup, exists := s.modules[newName]
	if exists {
		errs = append(errs,
			// seven characters at the start of the second line to align with the string "error: "
			fmt.Errorf("renaming module %q to %q conflicts with existing module\n"+
				"       %s <-- existing module defined here",
				oldName, newName, existingGroup.modules[0].pos),
		)
		return errs
	}

	group := s.modules[oldName]
	s.modules[newName] = group
	delete(s.modules, group.name)
	group.name = newName
	return []error{}
}

func (s *SimpleNameInterface) AllModules() []ModuleGroup {
	groups := make([]ModuleGroup, 0, len(s.modules))
	for _, group := range s.modules {
		groups = append(groups, group)
	}

	duplicateName := ""
	less := func(i, j int) bool {
		if groups[i].name == groups[j].name {
			duplicateName = groups[i].name
		}
		return groups[i].name < groups[j].name
	}
	sort.Slice(groups, less)
	if duplicateName != "" {
		// It is permitted to have two moduleGroup's with the same name, but not within the same
		// Namespace. The SimpleNameInterface should catch this in NewModule, however, so this
		// should never happen.
		panic(fmt.Sprintf("Duplicate moduleGroup name %q", duplicateName))
	}
	return groups
}

func (s *SimpleNameInterface) MissingDependencyError(depender string, dependerNamespace Namespace, dependency string) (err error) {
	return fmt.Errorf("%q depends on undefined module %q", depender, dependency)
}

func (s *SimpleNameInterface) GetNamespace(ctx NamespaceContext) Namespace {
	return nil
}

func (s *SimpleNameInterface) UniqueName(ctx NamespaceContext, name string) (unique string) {
	return name
}
