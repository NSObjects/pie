/*
 *
 * mapper.go
 * names
 *
 * Created by lintao on 2020/8/8 4:20 下午
 * Copyright © 2020-2020 LINTAO. All rights reserved.
 *
 */

package names

import (
	"strings"
	"sync"
)

// Mapper represents a name convertation between struct's fields name and table's column name
type Mapper interface {
	Obj2Collection(string) string
	Collection2Obj(string) string
}

type CacheMapper struct {
	oriMapper           Mapper
	obj2CollectionCache map[string]string
	obj2CollectionMutex sync.RWMutex
	collection2objCache map[string]string
	collection2objMutex sync.RWMutex
}

func NewCacheMapper(mapper Mapper) *CacheMapper {
	return &CacheMapper{oriMapper: mapper, obj2CollectionCache: make(map[string]string),
		collection2objCache: make(map[string]string),
	}
}

func (m *CacheMapper) Obj2Collection(o string) string {
	m.obj2CollectionMutex.RLock()
	t, ok := m.obj2CollectionCache[o]
	m.obj2CollectionMutex.RUnlock()
	if ok {
		return t
	}

	t = m.oriMapper.Obj2Collection(o)
	m.obj2CollectionMutex.Lock()
	m.obj2CollectionCache[o] = t
	m.obj2CollectionMutex.Unlock()
	return t
}

func (m *CacheMapper) Collection2Obj(t string) string {
	m.collection2objMutex.RLock()
	o, ok := m.collection2objCache[t]
	m.collection2objMutex.RUnlock()
	if ok {
		return o
	}

	o = m.oriMapper.Collection2Obj(t)
	m.collection2objMutex.Lock()
	m.collection2objCache[t] = o
	m.collection2objMutex.Unlock()
	return o
}

// SameMapper implements IMapper and provides same name between struct and
// database table
type SameMapper struct {
}

func (m SameMapper) Obj2Collection(o string) string {
	return o
}

func (m SameMapper) Collection2Obj(t string) string {
	return t
}

// SnakeMapper implements IMapper and provides name transaltion between
// struct and database table
type SnakeMapper struct {
}

func snakeCasedName(name string) string {
	newstr := make([]rune, 0)
	for idx, chr := range name {
		if isUpper := 'A' <= chr && chr <= 'Z'; isUpper {
			if idx > 0 {
				newstr = append(newstr, '_')
			}
			chr -= ('A' - 'a')
		}
		newstr = append(newstr, chr)
	}

	return string(newstr)
}

func (mapper SnakeMapper) Obj2Collection(name string) string {
	return snakeCasedName(name)
}

func titleCasedName(name string) string {
	newstr := make([]rune, 0)
	upNextChar := true

	name = strings.ToLower(name)

	for _, chr := range name {
		switch {
		case upNextChar:
			upNextChar = false
			if 'a' <= chr && chr <= 'z' {
				chr -= ('a' - 'A')
			}
		case chr == '_':
			upNextChar = true
			continue
		}

		newstr = append(newstr, chr)
	}

	return string(newstr)
}

func (mapper SnakeMapper) Collection2Obj(name string) string {
	return titleCasedName(name)
}

// GonicMapper implements IMapper. It will consider initialisms when mapping names.
// E.g. id -> ID, user -> User and to table names: UserID -> user_id, MyUID -> my_uid
type GonicMapper map[string]bool

func isASCIIUpper(r rune) bool {
	return 'A' <= r && r <= 'Z'
}

func toASCIIUpper(r rune) rune {
	if 'a' <= r && r <= 'z' {
		r -= ('a' - 'A')
	}
	return r
}

func gonicCasedName(name string) string {
	newstr := make([]rune, 0, len(name)+3)
	for idx, chr := range name {
		if isASCIIUpper(chr) && idx > 0 {
			if !isASCIIUpper(newstr[len(newstr)-1]) {
				newstr = append(newstr, '_')
			}
		}

		if !isASCIIUpper(chr) && idx > 1 {
			l := len(newstr)
			if isASCIIUpper(newstr[l-1]) && isASCIIUpper(newstr[l-2]) {
				newstr = append(newstr, newstr[l-1])
				newstr[l-1] = '_'
			}
		}

		newstr = append(newstr, chr)
	}
	return strings.ToLower(string(newstr))
}

func (mapper GonicMapper) Obj2Collection(name string) string {
	return gonicCasedName(name)
}

func (mapper GonicMapper) Collection2Obj(name string) string {
	newstr := make([]rune, 0)

	name = strings.ToLower(name)
	parts := strings.Split(name, "_")

	for _, p := range parts {
		_, isInitialism := mapper[strings.ToUpper(p)]
		for i, r := range p {
			if i == 0 || isInitialism {
				r = toASCIIUpper(r)
			}
			newstr = append(newstr, r)
		}
	}

	return string(newstr)
}

// LintGonicMapper is A GonicMapper that contains a list of common initialisms taken from golang/lint
var LintGonicMapper = GonicMapper{
	"API":   true,
	"ASCII": true,
	"CPU":   true,
	"CSS":   true,
	"DNS":   true,
	"EOF":   true,
	"GUID":  true,
	"HTML":  true,
	"HTTP":  true,
	"HTTPS": true,
	"ID":    true,
	"IP":    true,
	"JSON":  true,
	"LHS":   true,
	"QPS":   true,
	"RAM":   true,
	"RHS":   true,
	"RPC":   true,
	"SLA":   true,
	"SMTP":  true,
	"SSH":   true,
	"TLS":   true,
	"TTL":   true,
	"UI":    true,
	"UID":   true,
	"UUID":  true,
	"URI":   true,
	"URL":   true,
	"UTF8":  true,
	"VM":    true,
	"XML":   true,
	"XSRF":  true,
	"XSS":   true,
}

// PrefixMapper provides prefix table name support
type PrefixMapper struct {
	Mapper Mapper
	Prefix string
}

func (mapper PrefixMapper) Obj2Collection(name string) string {
	return mapper.Prefix + mapper.Mapper.Obj2Collection(name)
}

func (mapper PrefixMapper) Collection2Obj(name string) string {
	return mapper.Mapper.Collection2Obj(name[len(mapper.Prefix):])
}

func NewPrefixMapper(mapper Mapper, prefix string) PrefixMapper {
	return PrefixMapper{mapper, prefix}
}

// SuffixMapper provides suffix table name support
type SuffixMapper struct {
	Mapper Mapper
	Suffix string
}

func (mapper SuffixMapper) Obj2Collection(name string) string {
	return mapper.Mapper.Obj2Collection(name) + mapper.Suffix
}

func (mapper SuffixMapper) Collection2Obj(name string) string {
	return mapper.Mapper.Collection2Obj(name[:len(name)-len(mapper.Suffix)])
}

func NewSuffixMapper(mapper Mapper, suffix string) SuffixMapper {
	return SuffixMapper{mapper, suffix}
}
