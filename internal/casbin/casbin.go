package casbin

import (
	"bufio"
	_ "embed"
	"strings"

	"github.com/casbin/casbin/v3"
	"github.com/casbin/casbin/v3/model"
	"github.com/casbin/casbin/v3/persist"
	"github.com/leleo886/lopic/internal/log"
)

//go:embed rbac_model.conf
var modelConf string

//go:embed rbac_policy.csv
var policyCSV string

var Enforcer *casbin.Enforcer

type memoryAdapter struct {
	policy string
}

func (a *memoryAdapter) LoadPolicy(model model.Model) error {
	scanner := bufio.NewScanner(strings.NewReader(a.policy))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		persist.LoadPolicyLine(line, model)
	}
	return scanner.Err()
}

func (a *memoryAdapter) SavePolicy(model model.Model) error {
	return nil
}

func (a *memoryAdapter) AddPolicy(sec string, ptype string, rule []string) error {
	return nil
}

func (a *memoryAdapter) RemovePolicy(sec string, ptype string, rule []string) error {
	return nil
}

func (a *memoryAdapter) RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) error {
	return nil
}

func InitCasbin() (*casbin.Enforcer, error) {
	m, err := model.NewModelFromString(modelConf)
	if err != nil {
		return nil, err
	}

	a := &memoryAdapter{policy: policyCSV}

	enforcer, err := casbin.NewEnforcer(m, a)
	if err != nil {
		return nil, err
	}

	Enforcer = enforcer

	log.Info("Casbin initialized successfully")
	return enforcer, nil
}

// GetEnforcer 获取Casbin执行器实例
func GetEnforcer() *casbin.Enforcer {
	return Enforcer
}

// // AddPolicy
// func AddPolicy(params ...interface{}) bool {
// 	added, _ := Enforcer.AddPolicy(params...)
// 	return added
// }

// // RemovePolicy
// func RemovePolicy(params ...interface{}) bool {
// 	removed, _ := Enforcer.RemovePolicy(params...)
// 	return removed
// }

// // UpdatePolicy
// func UpdatePolicy(oldParams, newParams []interface{}) bool {
// 	// 将[]interface{}转换为[]string
// 	oldStrParams := make([]string, len(oldParams))
// 	for i, param := range oldParams {
// 		if str, ok := param.(string); ok {
// 			oldStrParams[i] = str
// 		}
// 	}

// 	newStrParams := make([]string, len(newParams))
// 	for i, param := range newParams {
// 		if str, ok := param.(string); ok {
// 			newStrParams[i] = str
// 		}
// 	}

// 	updated, _ := Enforcer.UpdatePolicy(oldStrParams, newStrParams)
// 	return updated
// }
