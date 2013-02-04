package main

import (
	"errors"
	e "github.com/MG-RAST/Shock/errors"
	"github.com/MG-RAST/Shock/store"
	"github.com/MG-RAST/Shock/store/user"
	"github.com/jaredwilkening/goweb"
	"net/http"
	"strings"
)

var (
	validAclTypes = map[string]bool{"read": true, "write": true, "delete": true, "owner": true}
)

// GET, POST, PUT, DELETE: /node/{nid}/acl/{type}
var AclController goweb.ControllerFunc = func(cx *goweb.Context) {
	LogRequest(cx.Request)
	u, err := AuthenticateRequest(cx.Request)
	if err != nil && err.Error() != e.NoAuth {
		handleAuthError(err, cx)
		return
	}

	// acl require auth even for public data
	if u == nil {
		cx.RespondWithErrorMessage(e.NoAuth, http.StatusUnauthorized)
		return
	}

	rtype := cx.PathParams["type"]
	if !validAclTypes[rtype] {
		cx.RespondWithErrorMessage("Invalid acl type", http.StatusBadRequest)
		return
	}

	// Load node and handle user unauthorized
	id := cx.PathParams["nid"]
	node, err := store.LoadNode(id, u.Uuid)
	if err != nil {
		if err.Error() == e.UnAuth {
			cx.RespondWithError(http.StatusUnauthorized)
			return
		} else if err.Error() == e.MongoDocNotFound {
			cx.RespondWithNotFound()
			return
		} else {
			// In theory the db connection could be lost between
			// checking user and load but seems unlikely.
			log.Error("Err@node_Read:LoadNode: " + err.Error())
			cx.RespondWithError(http.StatusInternalServerError)
			return
		}
	}

	rights := node.Acl.Check(u.Uuid)
	if cx.Request.Method != "GET" {
		ids, err := parseTypeRequest(cx)
		if err != nil {
			cx.RespondWithErrorMessage(err.Error(), http.StatusBadRequest)
			return
		}
		if (cx.Request.Method == "POST" || cx.Request.Method == "PUT") && (u.Uuid == node.Acl.Owner || rights["write"]) {
			if rtype == "owner" {
				if len(ids) == 1 {
					node.Acl.SetOwner(ids[0])
				} else {
					cx.RespondWithErrorMessage("Too many users. Nodes may have only one owner.", http.StatusBadRequest)
					return
				}
			} else {
				for _, i := range ids {
					node.Acl.Set(i, map[string]bool{rtype: true})
				}
			}
			node.Save()
		} else if cx.Request.Method == "DELETE" && (u.Uuid == node.Acl.Owner || rights["delete"]) {
			for _, i := range ids {
				node.Acl.UnSet(i, map[string]bool{rtype: true})
			}
			node.Save()
		} else {
			cx.RespondWithError(http.StatusUnauthorized)
			return
		}
	}

	if u.Uuid == node.Acl.Owner || rights["read"] {
		switch rtype {
		case "read":
			cx.RespondWithData(map[string][]string{"read": node.Acl.Read})
		case "write":
			cx.RespondWithData(map[string][]string{"write": node.Acl.Write})
		case "delete":
			cx.RespondWithData(map[string][]string{"delete": node.Acl.Delete})
		case "owner":
			cx.RespondWithData(map[string]string{"owner": node.Acl.Owner})
		}
	} else {
		cx.RespondWithError(http.StatusUnauthorized)
		return
	}
	return
}

// GET, POST, PUT, DELETE: /node/{nid}/acl/
var AclBaseController goweb.ControllerFunc = func(cx *goweb.Context) {
	LogRequest(cx.Request)
	u, err := AuthenticateRequest(cx.Request)
	if err != nil && err.Error() != e.NoAuth {
		handleAuthError(err, cx)
		return
	}

	// acl require auth even for public data
	if u == nil {
		cx.RespondWithErrorMessage(e.NoAuth, http.StatusUnauthorized)
		return
	}

	// Load node and handle user unauthorized
	id := cx.PathParams["nid"]
	node, err := store.LoadNode(id, u.Uuid)
	if err != nil {
		if err.Error() == e.UnAuth {
			cx.RespondWithError(http.StatusUnauthorized)
			return
		} else if err.Error() == e.MongoDocNotFound {
			cx.RespondWithNotFound()
			return
		} else {
			// In theory the db connection could be lost between
			// checking user and load but seems unlikely.
			log.Error("Err@node_Read:LoadNode: " + err.Error())
			cx.RespondWithError(http.StatusInternalServerError)
			return
		}
	}

	rights := node.Acl.Check(u.Uuid)
	if cx.Request.Method != "GET" {
		ids, err := parseRequest(cx)
		if err != nil {
			cx.RespondWithErrorMessage(err.Error(), http.StatusBadRequest)
			return
		}
		if (cx.Request.Method == "POST" || cx.Request.Method == "PUT") && (u.Uuid == node.Acl.Owner || rights["write"]) {
			for k, v := range ids {
				for _, i := range v {
					node.Acl.Set(i, map[string]bool{k: true})
				}
			}
			node.Save()
		} else if cx.Request.Method == "DELETE" && (u.Uuid == node.Acl.Owner || rights["delete"]) {
			for k, v := range ids {
				for _, i := range v {
					node.Acl.UnSet(i, map[string]bool{k: true})
				}
			}
			node.Save()
		} else {
			cx.RespondWithError(http.StatusUnauthorized)
			return
		}
	}

	if u.Uuid == node.Acl.Owner || rights["read"] {
		cx.RespondWithData(node.Acl)
	} else {
		cx.RespondWithError(http.StatusUnauthorized)
		return
	}
	return
}

func parseRequest(cx *goweb.Context) (ids map[string][]string, err error) {
	ids = map[string][]string{}
	users := map[string][]string{}
	query := &Query{list: cx.Request.URL.Query()}
	params, _, err := ParseMultipartForm(cx.Request)
	if err != nil && err.Error() == "request Content-Type isn't multipart/form-data" && (query.Has("all") || query.Has("read") || query.Has("write") || query.Has("delete")) {
		if query.Has("all") {
			users["all"] = strings.Split(query.Value("all"), ",")
		}
		if query.Has("read") {
			users["read"] = strings.Split(query.Value("read"), ",")
		}
		if query.Has("write") {
			users["write"] = strings.Split(query.Value("write"), ",")
		}
		if query.Has("delete") {
			users["delete"] = strings.Split(query.Value("delete"), ",")
		}
	} else if params["all"] != "" || params["read"] != "" || params["write"] != "" || params["delete"] != "" {
		users["all"] = strings.Split(params["all"], ",")
		users["read"] = strings.Split(params["read"], ",")
		users["write"] = strings.Split(params["write"], ",")
		users["delete"] = strings.Split(params["delete"], ",")
	} else {
		return nil, errors.New("Action requires list of comma seperated email address in 'all', 'read', 'write', and/or 'delete' parameter")
	}
	for k, _ := range users {
		for _, v := range users[k] {
			if v != "" {
				if isEmail(v) {
					u := user.User{Username: v, Email: v}
					if err := u.SetUuid(); err != nil {
						return nil, err
					}
					ids[k] = append(ids[k], u.Uuid)
				} else if isUuid(v) {
					ids[k] = append(ids[k], v)
				} else {
					return nil, errors.New("Unknown user id. Must be uuid or email address")
				}
			}
		}
	}
	if len(ids["all"]) > 0 {
		ids["read"] = append(ids["read"], ids["all"]...)
		ids["write"] = append(ids["write"], ids["all"]...)
		ids["delete"] = append(ids["delete"], ids["all"]...)
	}
	delete(ids, "all")
	return ids, nil
}

func parseTypeRequest(cx *goweb.Context) (ids []string, err error) {
	var users []string
	query := &Query{list: cx.Request.URL.Query()}
	params, _, err := ParseMultipartForm(cx.Request)
	if err != nil && err.Error() == "request Content-Type isn't multipart/form-data" && query.Has("users") {
		users = strings.Split(query.Value("users"), ",")
	} else if params["users"] != "" {
		users = strings.Split(params["users"], ",")
	} else {
		return nil, errors.New("Action requires list of comma seperated email address in 'users' parameter")
	}
	for _, v := range users {
		if isEmail(v) {
			u := user.User{Username: v, Email: v}
			if err := u.SetUuid(); err != nil {
				return nil, err
			}
			ids = append(ids, u.Uuid)
		} else if isUuid(v) {
			ids = append(ids, v)
		} else {
			return nil, errors.New("Unknown user id. Must be uuid or email address")
		}
	}
	return ids, nil
}

func isEmail(s string) bool {
	return (strings.Contains(s, "@") && strings.Contains(s, "."))
}

func isUuid(s string) bool {
	if strings.Count(s, "-") == 4 {
		return true
	}
	return false
}
