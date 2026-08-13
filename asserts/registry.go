// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * This file is part of snapd.
 * Copyright (C) 2026 Canonical Ltd.
 *
 * This program is free software: you can redistribute it and/or modify it
 * under the terms of the GNU General Public License version 3, as published
 * by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful, but WITHOUT
 * ANY WARRANTY; without even the implied warranties of MERCHANTABILITY,
 * SATISFACTORY QUALITY, or FITNESS FOR A PARTICULAR PURPOSE.
 * See the GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License along
 * with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package asserts

import (
	"fmt"
	"sort"
)

// Catalog is an immutable collection of assertion types.
type Catalog struct {
	assertionTypes []*AssertionType
}

// NewCatalog creates a catalog containing assertionTypes.
func NewCatalog(assertionTypes ...*AssertionType) (*Catalog, error) {
	registry, err := NewRegistry(assertionTypes...)
	if err != nil {
		return nil, err
	}
	return &Catalog{
		assertionTypes: registry.assertionTypes(),
	}, nil
}

func mustNewCatalog(assertionTypes ...*AssertionType) *Catalog {
	catalog, err := NewCatalog(assertionTypes...)
	if err != nil {
		panic(err)
	}
	return catalog
}

// Registry is an immutable collection of assertion types.
//
// Assertion types are shared between registries.
// Assertions can be passed between registries containing the same type
// pointers.
type Registry struct {
	types map[string]*AssertionType
}

// NewRegistry creates a registry containing assertionTypes.
func NewRegistry(assertionTypes ...*AssertionType) (*Registry, error) {
	types := make(map[string]*AssertionType, len(assertionTypes))
	for _, assertionType := range assertionTypes {
		if assertionType == nil {
			return nil, fmt.Errorf("cannot register a nil assertion type")
		}
		if assertionType.Name == "" {
			return nil, fmt.Errorf(
				"cannot register an assertion type without a name",
			)
		}
		if assertionType.assembler == nil {
			return nil, fmt.Errorf(
				"cannot register assertion type %q without an assembler",
				assertionType.Name,
			)
		}
		err := assertionType.validate()
		if err != nil {
			return nil, err
		}
		if types[assertionType.Name] != nil {
			return nil, fmt.Errorf(
				"cannot register duplicate assertion type %q",
				assertionType.Name,
			)
		}
		types[assertionType.Name] = assertionType
	}
	return &Registry{types: types}, nil
}

func mustNewRegistry(assertionTypes ...*AssertionType) *Registry {
	registry, err := NewRegistry(assertionTypes...)
	if err != nil {
		panic(err)
	}
	return registry
}

// NewRegistryFromCatalogs creates a registry containing the assertion types
// from catalogs.
func NewRegistryFromCatalogs(catalogs ...*Catalog) (*Registry, error) {
	var assertionTypes []*AssertionType
	for _, catalog := range catalogs {
		if catalog == nil {
			return nil, fmt.Errorf("cannot use a nil assertion catalog")
		}
		assertionTypes = append(
			assertionTypes,
			catalog.assertionTypes...,
		)
	}
	return NewRegistry(assertionTypes...)
}

func mustNewRegistryFromCatalogs(catalogs ...*Catalog) *Registry {
	registry, err := NewRegistryFromCatalogs(catalogs...)
	if err != nil {
		panic(err)
	}
	return registry
}

func (r *Registry) assertionTypes() []*AssertionType {
	assertionTypes := make([]*AssertionType, 0, len(r.types))
	for _, assertionType := range r.types {
		assertionTypes = append(assertionTypes, assertionType)
	}
	return assertionTypes
}

// Type returns the assertion type with name or nil.
func (r *Registry) Type(name string) *AssertionType {
	return r.types[name]
}

// TypeNames returns a sorted list of known assertion type names.
func (r *Registry) TypeNames() []string {
	names := make([]string, 0, len(r.types))
	for name := range r.types {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (r *Registry) checkAssertType(assertionType *AssertionType) error {
	if assertionType == nil {
		return fmt.Errorf("internal error: assertion type cannot be nil")
	}
	validity := r.types[assertionType.Name]
	switch validity {
	case assertionType:
		return nil
	case nil:
		return fmt.Errorf(
			"internal error: unknown assertion type: %q",
			assertionType.Name,
		)
	default:
		return fmt.Errorf(
			"internal error: unpredefined assertion type for name %q used "+
				"(unexpected address %p)",
			assertionType.Name,
			assertionType,
		)
	}
}
