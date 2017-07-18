// Package conf parses start up args and config file
package conf

import (
	//"flag"
	"errors"
	"fmt"
	"github.com/MG-RAST/golib/goconfig/config"
	"os"
	"strconv"
	"strings"
)

type idxOpts struct {
	unique   bool
	dropDups bool
	sparse   bool
}

var (
	// Admin
	ADMIN_EMAIL string
	ADMIN_USERS string
	AdminUsers  []string

	// Permissions for anonymous user
	ANON_READ   bool
	ANON_WRITE  bool
	ANON_DELETE bool

	// Address
	API_IP   string
	API_PORT int
	API_URL  string // for external address only

	// Auth
	AUTH_BASIC              bool
	AUTH_GLOBUS_TOKEN_URL   string
	AUTH_GLOBUS_PROFILE_URL string
	AUTH_OAUTH_URL_STR      string
	AUTH_OAUTH_BEARER_STR   string
	AUTH_CACHE_TIMEOUT      int
	AUTH_OAUTH              = make(map[string]string)
	OAUTH_DEFAULT           string // first value in AUTH_OAUTH_URL_STR

	// Default Chunksize for size virtual index
	CHUNK_SIZE int64 = 1048576

	// Config File
	CONFIG_FILE string
	LOG_OUTPUT  string

	// Runtime
	EXPIRE_WAIT   int // wait time for reaper in minutes
	GOMAXPROCS    string
	MAX_REVISIONS int // max number of node revisions to keep; values < 0 mean keep all

	// Logs
	LOG_PERF   bool // Indicates whether performance logs should be stored
	LOG_ROTATE bool // Indicates whether logs should be rotated daily

	// Mongo information
	MONGODB_HOSTS             string
	MONGODB_DATABASE          string
	MONGODB_USER              string
	MONGODB_PASSWORD          string
	MONGODB_ATTRIBUTE_INDEXES string

	// Node Indices
	NODE_IDXS map[string]idxOpts = nil

	// Paths
	PATH_SITE    string
	PATH_DATA    string
	PATH_LOGS    string
	PATH_LOCAL   string
	PATH_PIDFILE string

	// Reload
	RELOAD string

	// SSL
	SSL      bool
	SSL_KEY  string
	SSL_CERT string

	// Versions
	VERSIONS = make(map[string]int)

	PRINT_HELP   bool // full usage
	SHOW_HELP    bool // simple usage
	SHOW_VERSION bool

	// internal config control
	FAKE_VAR = false
)

// Initialize is an explicit init. Enables outside use
// of shock-server packages. Parses config and populates
// the conf variables.
func Initialize() (err error) {

	for i, elem := range os.Args {
		if strings.HasPrefix(elem, "-conf") || strings.HasPrefix(elem, "--conf") {
			parts := strings.SplitN(elem, "=", 2)
			if len(parts) == 2 {
				CONFIG_FILE = parts[1]
			} else if i+1 < len(os.Args) {
				CONFIG_FILE = os.Args[i+1]
			} else {
				err = errors.New("ERROR: parsing command options, missing conf file")
				return
			}
		}
	}

	var c *config.Config = nil
	if CONFIG_FILE != "" {
		c, err = config.ReadDefault(CONFIG_FILE)
		if err != nil {
			err = errors.New("ERROR: error reading conf file: " + err.Error())
			return
		}
		fmt.Printf("read %s\n", CONFIG_FILE)
	} else {
		fmt.Printf("No config file specified.\n")
		c = config.NewDefault()
	}

	c_store, err := getConfiguration(c) // from config file and command line arguments
	if err != nil {
		err = errors.New("ERROR: error reading conf file: " + err.Error())
		return
	}

	// ####### at this point configuration variables are set ########

	if FAKE_VAR == false {
		err = errors.New("ERROR: config was not parsed")
		return
	}
	if PRINT_HELP || SHOW_HELP {
		c_store.PrintHelp()
		os.Exit(0)
	}

	return
}

// Bool is a convenience wrapper around strconv.ParseBool
func Bool(s string) bool {
	b, _ := strconv.ParseBool(s)
	return b
}

// Print prints the configuration loads to stdout
func Print() {
	fmt.Printf("####### Anonymous ######\nread:\t%v\nwrite:\t%v\ndelete:\t%v\n\n", ANON_READ, ANON_WRITE, ANON_DELETE)
	if (AUTH_GLOBUS_TOKEN_URL != "" && AUTH_GLOBUS_PROFILE_URL != "") || len(AUTH_OAUTH) > 0 {
		fmt.Printf("##### Auth #####\n")
		if AUTH_GLOBUS_TOKEN_URL != "" && AUTH_GLOBUS_PROFILE_URL != "" {
			fmt.Printf("type:\tglobus\ntoken_url:\t%s\nprofile_url:\t%s\n\n", AUTH_GLOBUS_TOKEN_URL, AUTH_GLOBUS_PROFILE_URL)
		}
		if len(AUTH_OAUTH) > 0 {
			for b, u := range AUTH_OAUTH {
				fmt.Printf("bearer: %s\turl: %s\n", b, u)
			}
			fmt.Printf("\n")
		}
	}
	fmt.Printf("##### Admin #####\nusers:\t%s\n\n", ADMIN_USERS)
	fmt.Printf("##### Paths #####\nsite:\t%s\ndata:\t%s\nlogs:\t%s\nlocal_paths:\t%s\n\n", PATH_SITE, PATH_DATA, PATH_LOGS, PATH_LOCAL)
	if SSL {
		fmt.Printf("##### SSL enabled #####\n")
		fmt.Printf("##### SSL key:\t%s\n##### SSL cert:\t%s\n\n", SSL_KEY, SSL_CERT)
	} else {
		fmt.Printf("##### SSL disabled #####\n\n")
	}
	fmt.Printf("##### Mongodb #####\nhost(s):\t%s\ndatabase:\t%s\n\n", MONGODB_HOSTS, MONGODB_DATABASE)
	fmt.Printf("##### Address #####\nip:\t%s\nport:\t%d\n\n", API_IP, API_PORT)
	if LOG_PERF {
		fmt.Printf("##### PerfLog enabled #####\n\n")
	}
	if LOG_ROTATE {
		fmt.Printf("##### Log rotation enabled #####\n\n")
	} else {
		fmt.Printf("##### Log rotation disabled #####\n\n")
	}
	fmt.Printf("##### Expiration #####\nexpire_wait:\t%d minutes\n\n", EXPIRE_WAIT)
	fmt.Printf("##### Max Revisions #####\nmax_revisions:\t%d\n\n", MAX_REVISIONS)
	fmt.Printf("API_PORT: %d\n", API_PORT)
}

func getConfiguration(c *config.Config) (c_store *Config_store, err error) {
	c_store = NewCS(c)

	// Admin
	//ADMIN_EMAIL, _ = c.String("Admin", "email")
	c_store.AddString(&ADMIN_EMAIL, "", "Admin", "email", "", "")
	//ADMIN_USERS, _ = c.String("Admin", "users")
	c_store.AddString(&ADMIN_USERS, "", "Admin", "users", "", "")
	if ADMIN_USERS != "" {
		for _, name := range strings.Split(ADMIN_USERS, ",") {
			AdminUsers = append(AdminUsers, strings.TrimSpace(name))
		}
	}

	// Access-Control
	c_store.AddBool(&ANON_READ, true, "Anonymous", "read", "", "")
	c_store.AddBool(&ANON_WRITE, true, "Anonymous", "write", "", "")
	c_store.AddBool(&ANON_DELETE, true, "Anonymous", "delete", "", "")

	// Address
	c_store.AddString(&API_IP, "0.0.0.0", "Address", "api-ip", "", "")
	c_store.AddInt(&API_PORT, 7445, "Address", "api-port", "", "")

	// URLs
	c_store.AddString(&API_URL, "http://localhost", "External", "api-url", "", "")

	// Auth
	c_store.AddBool(&AUTH_BASIC, false, "Auth", "basic", "", "")
	c_store.AddString(&AUTH_GLOBUS_TOKEN_URL, "", "Auth", "globus_token_url", "", "")
	c_store.AddString(&AUTH_GLOBUS_PROFILE_URL, "", "Auth", "globus_profile_url", "", "")
	c_store.AddString(&AUTH_OAUTH_URL_STR, "", "Auth", "oauth_urls", "", "")
	c_store.AddString(&AUTH_OAUTH_BEARER_STR, "", "Auth", "oauth_bearers", "", "")
	c_store.AddInt(&AUTH_CACHE_TIMEOUT, 60, "Auth", "cache_timeout", "", "")

	// parse OAuth settings if used
	if AUTH_OAUTH_URL_STR != "" && AUTH_OAUTH_BEARER_STR != "" {
		ou := strings.Split(AUTH_OAUTH_URL_STR, ",")
		ob := strings.Split(AUTH_OAUTH_BEARER_STR, ",")
		if len(ou) != len(ob) {
			err = errors.New("ERROR: number of items in oauth_urls and oauth_bearers are not the same")
			return
		}
		for i := range ob {
			AUTH_OAUTH[ob[i]] = ou[i]
		}
		OAUTH_DEFAULT = ou[0] // first url is default for "oauth" bearer token
	}

	// Runtime
	c_store.AddInt(&EXPIRE_WAIT, 60, "Runtime", "expire_wait", "", "")
	c_store.AddString(&GOMAXPROCS, "", "Runtime", "GOMAXPROCS", "", "")
	c_store.AddInt(&MAX_REVISIONS, 3, "Runtime", "max_revisions", "", "")

	c_store.AddBool(&LOG_PERF, false, "Log", "perf_log", "", "")
	c_store.AddBool(&LOG_ROTATE, true, "Log", "rotate", "", "")

	// Mongodb
	c_store.AddString(&MONGODB_ATTRIBUTE_INDEXES, "", "Mongodb", "attribute_indexes", "", "")
	c_store.AddString(&MONGODB_DATABASE, "ShockDB", "Mongodb", "database", "", "")
	c_store.AddString(&MONGODB_HOSTS, "mongo", "Mongodb", "hosts", "", "")
	c_store.AddString(&MONGODB_PASSWORD, "", "Mongodb", "password", "", "")
	c_store.AddString(&MONGODB_USER, "", "Mongodb", "user", "", "")

	// parse Node-Indices
	NODE_IDXS = map[string]idxOpts{}
	nodeIdx, _ := c.Options("Node-Indices")
	for _, opt := range nodeIdx {
		val, _ := c.String("Node-Indices", opt)
		opts := idxOpts{}
		for _, parts := range strings.Split(val, ",") {
			p := strings.Split(parts, ":")
			if p[0] == "unique" {
				if p[1] == "true" {
					opts.unique = true
				} else {
					opts.unique = false
				}
			} else if p[0] == "dropDups" {
				if p[1] == "true" {
					opts.dropDups = true
				} else {
					opts.dropDups = false
				}
			} else if p[0] == "sparse" {
				if p[1] == "true" {
					opts.sparse = true
				} else {
					opts.sparse = false
				}
			}
		}
		NODE_IDXS[opt] = opts
	}

	// Paths
	c_store.AddString(&PATH_SITE, "/usr/local/shock/site", "Paths", "site", "", "")
	c_store.AddString(&PATH_DATA, "/usr/local/shock", "Paths", "data", "", "")
	c_store.AddString(&PATH_LOGS, "/var/log/shock", "Paths", "logs", "", "")
	c_store.AddString(&PATH_LOCAL, "", "Paths", "local_paths", "", "")
	c_store.AddString(&PATH_PIDFILE, "", "Paths", "pidfile", "", "")

	// SSL
	c_store.AddBool(&SSL, false, "SSL", "enable", "", "")
	if SSL {
		c_store.AddString(&SSL_KEY, "", "SSL", "key", "", "")
		c_store.AddString(&SSL_CERT, "", "SSL", "cert", "", "")
	}

	// Log
	c_store.AddString(&LOG_OUTPUT, "console", "Log", "logoutput", "console, file or both", "")

	//Other
	c_store.AddString(&RELOAD, "", "Other", "reload", "path or url to shock data. WARNING this will drop all current data.", "")
	gopath := os.Getenv("GOPATH")
	c_store.AddString(&CONFIG_FILE, gopath+"/src/github.com/MG-RAST/Shock/shock-server.conf.template", "Other", "conf", "path to config file", "")
	c_store.AddBool(&SHOW_VERSION, false, "Other", "version", "show version", "")
	c_store.AddBool(&PRINT_HELP, false, "Other", "fullhelp", "show detailed usage without \"--\"-prefixes", "")
	c_store.AddBool(&SHOW_HELP, false, "Other", "help", "show usage", "")

	VERSIONS["ACL"] = 2
	VERSIONS["Auth"] = 1
	VERSIONS["Node"] = 4

	c_store.Parse()

	return

}
