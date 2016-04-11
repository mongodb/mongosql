package evaluator

import (
	"github.com/mongodb/mongo-tools/common/log"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// NewAuthProvider constructs a new lazy auth provider.
func NewAuthProvider(session *mgo.Session) AuthProvider {
	return &lazyAuthProvider{session: session}
}

// AuthProvider provides authorization information.
type AuthProvider interface {
	// IsDatabaseAllowed indicates whether a database is allowed to be used.
	IsDatabaseAllowed(string) bool
	// IsCollectionAllowed indicates whether a collection is allowed to be used.
	IsCollectionAllowed(string, string) bool
}

// FixedAuthProvider returns a fixed value for restrictions.
type FixedAuthProvider struct {
	allow bool
}

// IsDatabaseAllowed indicates whether a database is allowed to be used.
func (p *FixedAuthProvider) IsDatabaseAllowed(_ string) bool {
	return p.allow
}

// IsCollectionAllowed indicates whether a collection is allowed to be used.
func (p *FixedAuthProvider) IsCollectionAllowed(_, _ string) bool {
	return p.allow
}

type lazyAuthProvider struct {
	session *mgo.Session
	actual  AuthProvider
}

func (p *lazyAuthProvider) IsDatabaseAllowed(dbName string) bool {
	p.ensureInitialized()
	return p.actual.IsDatabaseAllowed(dbName)
}

func (p *lazyAuthProvider) IsCollectionAllowed(dbName, colName string) bool {
	p.ensureInitialized()
	return p.actual.IsCollectionAllowed(dbName, colName)
}

func (p *lazyAuthProvider) ensureInitialized() {
	if p.actual == nil {
		a, err := loadAuthProvider(p.session)
		if err != nil {
			log.Logf(log.Always, "failed to initialize auth provider: %v", err)
			p.actual = &FixedAuthProvider{false}
		}
		p.actual = a
	}
}

type mongoAuthProvider struct {
	databases map[string]map[string]bool
}

func loadAuthProvider(session *mgo.Session) (AuthProvider, error) {
	cmd := bson.D{
		{"connectionStatus", 1},
		{"showPrivileges", 1},
	}
	result := bson.M{}
	if err := session.Run(cmd, &result); err != nil {
		// no auth required. if this fails, it's a real problem...
		return nil, err
	}

	return loadAuthProviderFromConnectionStatus(&result), nil
}

func loadAuthProviderFromConnectionStatus(result *bson.M) AuthProvider {
	authUsers, found := findArrayInDoc("authInfo.authenticatedUsers", result)
	if !found || len(authUsers) == 0 {
		return &FixedAuthProvider{true}
	}

	provider := &mongoAuthProvider{}

	authUserPrivileges, found := findArrayInDoc("authInfo.authenticatedUserPrivileges", result)
	if found {
		for _, p := range authUserPrivileges {
			if resource, found := findDocInDoc("resource", p); found {
				if actions, found := findArrayInDoc("actions", p); found {
					hasFindAction := false
					for _, action := range actions {
						if a, ok := action.(string); ok {
							if a == "find" {
								hasFindAction = true
							}
						}
					}

					if dbName, ok := findStringInDoc("db", resource); ok {
						if colName, ok := findStringInDoc("collection", resource); ok {
							provider.addCollection(dbName, colName, hasFindAction)
						}
					}
				}
			}
		}
	}

	return provider
}

func (p *mongoAuthProvider) addCollection(dbName, colName string, allowed bool) {
	if p.databases == nil {
		p.databases = make(map[string]map[string]bool)
	}
	cols, ok := p.databases[dbName]
	if !ok {
		cols = make(map[string]bool)
		p.databases[dbName] = cols
	}

	cols[colName] = allowed
}

func (p *mongoAuthProvider) IsDatabaseAllowed(dbName string) bool {
	if p.databases == nil {
		return false
	}

	if cols, found := p.databases[dbName]; found {
		// if any of the collections are allowed, return true, otherwise, false
		for _, v := range cols {
			if v {
				return true
			}
		}
	}

	return false
}

func (p *mongoAuthProvider) IsCollectionAllowed(dbName string, colName string) bool {
	if p.databases == nil {
		return false
	}

	db, found := p.databases[dbName]
	if !found {
		db, found = p.databases[""]
		if !found {
			return false
		}
	}

	val, ok := db[colName]
	if !ok {
		val, ok = db[""]
	}

	return ok && val
}
