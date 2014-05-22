package main

import (
	"errors"
	"fmt"
	"sort"

	"github.com/garyburd/redigo/redis"
	"github.com/twinj/uuid"
)

const (
	GetPermission    = "get"
	UploadPermission = "upload"
)

var (
	permissionsByKey map[string]map[string]bool
)

func init() {
	// Change the UUID format to remove surrounding braces and dashes
	uuid.SwitchFormat(uuid.Clean, true)
}

func authInit() error {
	keys, err := listKeys()
	if err != nil {
		return err
	}

	permissionsByKey = make(map[string]map[string]bool)

	// Set up permissions for when there's no API key
	config := getConfig()
	permissionsByKey[""] = make(map[string]bool)
	permissionsByKey[""][GetPermission] = !config.authorisedGet
	permissionsByKey[""][UploadPermission] = !config.authorisedUpload

	// Set up permissions for API keys
	for _, key := range keys {
		permissions, err := infoAboutKey(key)
		if err != nil {
			return err
		}
		permissionsByKey[key] = make(map[string]bool)
		for _, permission := range permissions {
			permissionsByKey[key][permission] = true
		}
	}

	return nil
}

func hasPermission(key, permission string) bool {
	val, ok := permissionsByKey[key][permission]
	if ok {
		return val
	}
	return false
}

func generateKey() (string, error) {
	key := uuid.NewV4().String()
	_, err := conn.Do("SADD", "api-keys", key)
	if err != nil {
		return "", err
	}
	_, err = conn.Do("SADD", "key:"+key, GetPermission, UploadPermission)
	return key, err
}

func infoAboutKey(key string) ([]string, error) {
	err := checkKeyExists(key)
	if err != nil {
		return nil, err
	}
	permissions, err := redis.Strings(conn.Do("SMEMBERS", "key:"+key))
	if err != nil {
		return nil, err
	}
	sort.Strings(permissions)
	return permissions, nil
}

func listKeys() ([]string, error) {
	return redis.Strings(conn.Do("SMEMBERS", "api-keys"))
}

func modifyKey(key, op, permission string) error {
	err := checkKeyExists(key)
	if err != nil {
		return err
	}
	if op != "add" && op != "remove" {
		return errors.New("Modifier needs to be 'add' or 'remove'")
	}
	if permission != GetPermission && permission != UploadPermission {
		return fmt.Errorf("Modifier needs to end with a valid permission: %s or %s", GetPermission, UploadPermission)
	}
	if op == "add" {
		_, err = conn.Do("SADD", "key:"+key, permission)
	} else {
		_, err = conn.Do("SREM", "key:"+key, permission)
	}
	return err
}

func removeKey(key string) error {
	err := checkKeyExists(key)
	if err != nil {
		return err
	}
	_, err = conn.Do("SREM", "api-keys", key)
	if err != nil {
		return err
	}
	_, err = conn.Do("DEL", "key:"+key)
	return err
}

func authPermissionsOptions() string {
	return fmt.Sprintf("%s/%s", GetPermission, UploadPermission)
}

func checkKeyExists(key string) error {
	exists, err := redis.Bool(conn.Do("SISMEMBER", "api-keys", key))
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("Key does not exist")
	}
	return nil
}
