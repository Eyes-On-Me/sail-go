package service

import (
	"net/http"
	"path"
	"github.com/sail-services/sail-go/com/base"
	"github.com/sail-services/sail-go/com/data/convert"
	estr "github.com/sail-services/sail-go/com/data/strings"
	"regexp"
	"strings"
	"sync"
)

type (
	routerPro struct {
		ser          *Service
		absolutePath string
		routers      map[string]*proTree
		groups       []proGroup
		notFound     http.HandlerFunc
		*proMap
	}
	proGroup struct {
		pattern  string
		handlers []Handler
	}
	proCombo struct {
		router   *routerPro
		pattern  string
		handlers []Handler
		methods  map[string]bool
	}
	proTree struct {
		parent    *proTree
		ptype     patternType
		pattern   string
		wildcards []string
		reg       *regexp.Regexp
		subtrees  []*proTree
		leaves    []*proLeaf
	}
	proLeaf struct {
		parent    *proTree
		ptype     patternType
		pattern   string
		wildcards []string
		reg       *regexp.Regexp
		optional  bool
		name      string
		handle    handle
	}
	proMap struct {
		lock   sync.RWMutex
		routes map[string]map[string]bool
	}
	handle      func(http.ResponseWriter, *http.Request, reqParams)
	patternType int8
)

const (
	_PATTERN_STATIC patternType = iota
	_PATTERN_REGEXP
	_PATTERN_PATH_EXT
	_PATTERN_MATCH_ALL
)

var (
	_wildcard_pattern = regexp.MustCompile(`:[a-zA-Z0-9]+`)
	_string_pattern   = regexp.MustCompile(`(.+)`)
	_HTTP_METHODS     = map[string]bool{
		"GET":     true,
		"POST":    true,
		"PUT":     true,
		"DELETE":  true,
		"PATCH":   true,
		"OPTIONS": true,
		"HEAD":    true,
	}
)

// ========================================================
// routerPro
// ========================================================
func (rou *routerPro) init(ser *Service) {
	rou.ser = ser
	rou.absolutePath = _PATH_ROOT
	rou.routers = make(map[string]*proTree)
	rou.proMap = proMapNew()
	var not_found_func []Handler
	not_found_func = append(not_found_func, func(con *Context) {
		con.Ren.S(404, _E404)
	})
	rou.notFound = func(resp http.ResponseWriter, req *http.Request) {
		hds := rou.ser.modsCombine(not_found_func)
		con := rou.ser.contextNew(resp, req, hds)
		con.Resp.WriteHeader(404)
		con.Next()
		con.Resp.writeHeader()
		rou.ser.pool.Put(con)
	}
}

func (rou *routerPro) Group(rpath string, function func(), hds ...Handler) {
	rou.groups = append(rou.groups, proGroup{rpath, hds})
	function()
	rou.groups = rou.groups[:len(rou.groups)-1]
}

func (rou *routerPro) Get(rpath string, hds ...Handler) {
	rou.Handle("GET", rpath, hds)
}

func (rou *routerPro) Patch(rpath string, hds ...Handler) {
	rou.Handle("PATCH", rpath, hds)
}

func (rou *routerPro) Post(rpath string, hds ...Handler) {
	rou.Handle("POST", rpath, hds)
}

func (rou *routerPro) Put(rpath string, hds ...Handler) {
	rou.Handle("PUT", rpath, hds)
}

func (rou *routerPro) Delete(rpath string, hds ...Handler) {
	rou.Handle("DELETE", rpath, hds)
}

func (rou *routerPro) Options(rpath string, hds ...Handler) {
	rou.Handle("OPTIONS", rpath, hds)
}

func (rou *routerPro) Head(rpath string, hds ...Handler) {
	rou.Handle("HEAD", rpath, hds)
}

func (rou *routerPro) Link(rpath string, hds ...Handler) {
	rou.Handle("LINK", rpath, hds)
}

func (rou *routerPro) Unlink(rpath string, hds ...Handler) {
	rou.Handle("UNLINK", rpath, hds)
}

func (rou *routerPro) Any(rpath string, hds ...Handler) {
	rou.Handle("*", rpath, hds)
}

func (rou *routerPro) Route(rpath, methods string, hds ...Handler) {
	for _, m := range strings.Split(methods, ",") {
		rou.Handle(strings.TrimSpace(m), rpath, hds)
	}
}

func (rou *routerPro) Combo(rpath string, hds ...Handler) *proCombo {
	return &proCombo{rou, rpath, hds, map[string]bool{}}
}

func (rou *routerPro) File(rpath, fpath string) {
	full_pattern := rou.calculateAbsolutePath(rpath)
	if len(rou.groups) > 0 {
		group_pattern := ""
		for _, g := range rou.groups {
			group_pattern += g.pattern
		}
		full_pattern = group_pattern + rpath
	}
	if ModeIsDev() {
		rou.ser.Log.Infof("%v %v -> %v\n", "GET", full_pattern, fpath)
	}
	rou.handle("GET", full_pattern, func(resp http.ResponseWriter, req *http.Request, params reqParams) {
		http.ServeFile(resp, req, fpath)
	})
}

func (rou *routerPro) NotFound(hds ...Handler) {
	hds = rou.ser.modsCombine(hds)
	rou.notFound = func(resp http.ResponseWriter, req *http.Request) {
		con := rou.ser.contextNew(resp, req, hds)
		con.Resp.WriteHeader(404)
		con.Next()
		con.Resp.writeHeader()
		rou.ser.pool.Put(con)
	}
}

func (rou *routerPro) Handle(method string, rpath string, hds []Handler) {
	full_pattern := rpath
	if len(rou.groups) > 0 {
		group_pattern := ""
		h := make([]Handler, 0)
		for _, g := range rou.groups {
			group_pattern += g.pattern
			h = append(h, g.handlers...)
		}
		full_pattern = group_pattern + rpath
		h = append(h, hds...)
		hds = h
	}
	hds = rou.ser.modsCombine(hds)
	if ModeIsDev() {
		rou.ser.Log.Infof("%v %v -> %v (%v)\n", method, full_pattern, base.FuncNameGet(hds[len(hds)-1]), len(hds))
	}
	rou.handle(method, full_pattern, func(resp http.ResponseWriter, req *http.Request, params reqParams) {
		con := rou.ser.contextNew(resp, req, hds)
		con.Req.params = params
		con.Next()
		con.Resp.writeHeader()
		rou.ser.pool.Put(con)
	})
}

func (rou *routerPro) handle(method, rpath string, handle handle) {
	method = strings.ToUpper(method)
	if rou.isExist(method, rpath) {
		return
	}
	if !_HTTP_METHODS[method] && method != "*" {
		panic("unknown HTTP method: " + method)
	}
	methods := make(map[string]bool)
	if method == "*" {
		for m := range _HTTP_METHODS {
			methods[m] = true
		}
	} else {
		methods[method] = true
	}
	for m := range methods {
		if t, ok := rou.routers[m]; ok {
			t.Add(rpath, "", handle)
		} else {
			t := treeNew()
			t.Add(rpath, "", handle)
			rou.routers[m] = t
		}
		rou.add(m, rpath)
	}
}

func (rou *routerPro) calculateAbsolutePath(rpath string) string {
	if rpath == "" {
		return rou.absolutePath
	}
	apath := path.Join(rou.absolutePath, rpath)
	append_slash := estr.LastChar(rpath) == '/' && estr.LastChar(apath) != '/'
	if append_slash {
		return apath + "/"
	}
	return apath
}

// --------------------------------------------------------
// routerPro - GO
// --------------------------------------------------------
func (r *routerPro) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if t, ok := r.routers[req.Method]; ok {
		h, p, ok := t.Match(req.URL.Path)
		if ok {
			if splat, ok := p["*0"]; ok {
				p["*"] = splat
			}
			h(rw, req, p)
			return
		}
	}
	r.notFound(rw, req)
}

// ========================================================
// proCombo
// ========================================================
func (cr *proCombo) checkMethod(name string) {
	if cr.methods[name] {
		panic("method '" + name + "' has already been registered")
	}
	cr.methods[name] = true
}

func (cr *proCombo) route(fn func(string, ...Handler), method string, h ...Handler) *proCombo {
	cr.checkMethod(method)
	fn(cr.pattern, append(cr.handlers, h...)...)
	return cr
}

func (cr *proCombo) Get(h ...Handler) *proCombo {
	return cr.route(cr.router.Get, "GET", h...)
}

func (cr *proCombo) Patch(h ...Handler) *proCombo {
	return cr.route(cr.router.Patch, "PATCH", h...)
}

func (cr *proCombo) Post(h ...Handler) *proCombo {
	return cr.route(cr.router.Post, "POST", h...)
}

func (cr *proCombo) Put(h ...Handler) *proCombo {
	return cr.route(cr.router.Put, "PUT", h...)
}

func (cr *proCombo) Delete(h ...Handler) *proCombo {
	return cr.route(cr.router.Delete, "DELETE", h...)
}

func (cr *proCombo) Options(h ...Handler) *proCombo {
	return cr.route(cr.router.Options, "OPTIONS", h...)
}

func (cr *proCombo) Head(h ...Handler) *proCombo {
	return cr.route(cr.router.Head, "HEAD", h...)
}

// ========================================================
// proLeaf
// ========================================================
func leafNew(parent *proTree, pattern, name string, handle handle) *proLeaf {
	typ, wildcards, reg := checkPattern(pattern)
	optional := false
	if len(pattern) > 0 && pattern[0] == '?' {
		optional = true
	}
	return &proLeaf{parent, typ, pattern, wildcards, reg, optional, name, handle}
}

// ========================================================
// proTree
// ========================================================
func treeNew() *proTree {
	return subTreeNew(nil, "")
}

func subTreeNew(parent *proTree, pattern string) *proTree {
	typ, wildcards, reg := checkPattern(pattern)
	return &proTree{parent, typ, pattern, wildcards, reg, make([]*proTree, 0, 5), make([]*proLeaf, 0, 5)}
}

func (t *proTree) addLeaf(pattern, name string, handle handle) bool {
	for i := 0; i < len(t.leaves); i++ {
		if t.leaves[i].pattern == pattern {
			return true
		}
	}
	leaf := leafNew(t, pattern, name, handle)
	if leaf.optional {
		parent := leaf.parent
		if parent.parent != nil {
			parent.parent.addLeaf(parent.pattern, name, handle)
		} else {
			parent.addLeaf("", name, handle)
		}
	}
	i := 0
	for ; i < len(t.leaves); i++ {
		if leaf.ptype < t.leaves[i].ptype {
			break
		}
	}
	if i == len(t.leaves) {
		t.leaves = append(t.leaves, leaf)
	} else {
		t.leaves = append(t.leaves[:i], append([]*proLeaf{leaf}, t.leaves[i:]...)...)
	}
	return false
}

func (t *proTree) addSubTree(segment, pattern, name string, handle handle) bool {
	for i := 0; i < len(t.subtrees); i++ {
		if t.subtrees[i].pattern == segment {
			return t.subtrees[i].addNextSegment(pattern, name, handle)
		}
	}
	subtree := subTreeNew(t, segment)
	i := 0
	for ; i < len(t.subtrees); i++ {
		if subtree.ptype < t.subtrees[i].ptype {
			break
		}
	}
	if i == len(t.subtrees) {
		t.subtrees = append(t.subtrees, subtree)
	} else {
		t.subtrees = append(t.subtrees[:i], append([]*proTree{subtree}, t.subtrees[i:]...)...)
	}
	return subtree.addNextSegment(pattern, name, handle)
}

func (t *proTree) addNextSegment(pattern, name string, handle handle) bool {
	pattern = strings.TrimPrefix(pattern, "/")
	i := strings.Index(pattern, "/")
	if i == -1 {
		return t.addLeaf(pattern, name, handle)
	}
	return t.addSubTree(pattern[:i], pattern[i+1:], name, handle)
}

func (t *proTree) Add(pattern, name string, handle handle) bool {
	pattern = strings.TrimSuffix(pattern, "/")
	return t.addNextSegment(pattern, name, handle)
}

func (t *proTree) matchLeaf(globLevel int, url string, params reqParams) (handle, bool) {
	for i := 0; i < len(t.leaves); i++ {
		switch t.leaves[i].ptype {
		case _PATTERN_STATIC:
			if t.leaves[i].pattern == url {
				return t.leaves[i].handle, true
			}
		case _PATTERN_REGEXP:
			results := t.leaves[i].reg.FindStringSubmatch(url)
			if len(results)-1 != len(t.leaves[i].wildcards) {
				break
			}
			for j := 0; j < len(t.leaves[i].wildcards); j++ {
				params[t.leaves[i].wildcards[j]] = results[j+1]
			}
			return t.leaves[i].handle, true
		case _PATTERN_PATH_EXT:
			j := strings.LastIndex(url, ".")
			if j > -1 {
				params[":path"] = url[:j]
				params[":ext"] = url[j+1:]
			} else {
				params[":path"] = url
			}
			return t.leaves[i].handle, true
		case _PATTERN_MATCH_ALL:
			params["*"+convert.IToS(globLevel)] = url
			return t.leaves[i].handle, true
		}
	}
	return nil, false
}

func (t *proTree) matchSubTree(globLevel int, segment, url string, params reqParams) (handle, bool) {
	for i := 0; i < len(t.subtrees); i++ {
		switch t.subtrees[i].ptype {
		case _PATTERN_STATIC:
			if t.subtrees[i].pattern == segment {
				if handle, ok := t.subtrees[i].matchNextSegment(globLevel, url, params); ok {
					return handle, true
				}
			}
		case _PATTERN_REGEXP:
			results := t.subtrees[i].reg.FindStringSubmatch(segment)
			if len(results)-1 != len(t.subtrees[i].wildcards) {
				break
			}
			for j := 0; j < len(t.subtrees[i].wildcards); j++ {
				params[t.subtrees[i].wildcards[j]] = results[j+1]
			}
			if handle, ok := t.subtrees[i].matchNextSegment(globLevel, url, params); ok {
				return handle, true
			}
		case _PATTERN_MATCH_ALL:
			if handle, ok := t.subtrees[i].matchNextSegment(globLevel+1, url, params); ok {
				params["*"+convert.IToS(globLevel)] = segment
				return handle, true
			}
		}
	}
	if len(t.leaves) > 0 {
		leaf := t.leaves[len(t.leaves)-1]
		if leaf.ptype == _PATTERN_PATH_EXT {
			url = segment + "/" + url
			j := strings.LastIndex(url, ".")
			if j > -1 {
				params[":path"] = url[:j]
				params[":ext"] = url[j+1:]
			} else {
				params[":path"] = url
			}
			return leaf.handle, true
		} else if leaf.ptype == _PATTERN_MATCH_ALL {
			params["*"+convert.IToS(globLevel)] = segment + "/" + url
			return leaf.handle, true
		}
	}
	return nil, false
}

func (t *proTree) Match(url string) (handle, reqParams, bool) {
	url = strings.TrimSuffix(url, "/")
	params := make(reqParams)
	handle, ok := t.matchNextSegment(0, url, params)
	return handle, params, ok
}

func (t *proTree) matchNextSegment(globLevel int, url string, params reqParams) (handle, bool) {
	url = strings.TrimPrefix(url, "/")
	i := strings.Index(url, "/")
	if i == -1 {
		return t.matchLeaf(globLevel, url, params)
	}
	return t.matchSubTree(globLevel, url[:i], url[i+1:], params)
}

// ========================================================
// proMap
// ========================================================
func proMapNew() *proMap {
	rm := &proMap{
		routes: make(map[string]map[string]bool),
	}
	for m := range _HTTP_METHODS {
		rm.routes[m] = make(map[string]bool)
	}
	return rm
}

func (rm *proMap) isExist(method, pattern string) bool {
	rm.lock.RLock()
	defer rm.lock.RUnlock()

	return rm.routes[method][pattern]
}

func (rm *proMap) add(method, pattern string) {
	rm.lock.Lock()
	defer rm.lock.Unlock()

	rm.routes[method][pattern] = true
}

// --------------------------------------------------------
// FUNC
// --------------------------------------------------------
func getNextWildcard(pattern string) (wildcard string, _ string) {
	pos := _wildcard_pattern.FindStringIndex(pattern)
	if pos == nil {
		return "", pattern
	}
	wildcard = pattern[pos[0]:pos[1]]
	if len(pattern) == pos[1] {
		return wildcard, strings.Replace(pattern, wildcard, `(.+)`, 1)
	} else if pattern[pos[1]] != '(' {
		if len(pattern) >= pos[1]+4 && pattern[pos[1]:pos[1]+4] == ":int" {
			pattern = strings.Replace(pattern, ":int", "([0-9]+)", -1)
		} else {
			return wildcard, strings.Replace(pattern, wildcard, `(.+)`, 1)
		}
	}
	return wildcard, pattern[:pos[0]] + pattern[pos[1]:]
}

func getWildcards(pattern string) (string, []string) {
	wildcards := make([]string, 0, 2)
	var wildcard string
	for {
		wildcard, pattern = getNextWildcard(pattern)
		if len(wildcard) > 0 {
			wildcards = append(wildcards, wildcard)
		} else {
			break
		}
	}
	return pattern, wildcards
}

func checkPattern(pattern string) (typ patternType, wildcards []string, reg *regexp.Regexp) {
	pattern = strings.TrimLeft(pattern, "?")
	if pattern == "*" {
		typ = _PATTERN_MATCH_ALL
	} else if pattern == "*.*" {
		typ = _PATTERN_PATH_EXT
	} else if strings.Contains(pattern, ":") {
		typ = _PATTERN_REGEXP
		pattern, wildcards = getWildcards(pattern)
		if pattern == "(.+)" {
			reg = _string_pattern
		} else {
			reg = regexp.MustCompile(pattern)
		}
	}
	return typ, wildcards, reg
}
