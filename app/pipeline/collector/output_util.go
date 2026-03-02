package collector

import (
	"github.com/andeya/pholcus/logs"
)

// namespace returns the main namespace (relative to DB name); optional, does not depend on data content.
func (c *Collector) namespace() string {
	if c.Spider.Namespace == nil {
		if c.Spider.GetSubName() == "" {
			return c.Spider.GetName()
		}
		return c.Spider.GetName() + "__" + c.Spider.GetSubName()
	}
	return c.Spider.Namespace(c.Spider)
}

// subNamespace returns the sub-namespace (relative to table name); optional, may depend on data content.
func (c *Collector) subNamespace(dataCell map[string]interface{}) string {
	if c.Spider.SubNamespace == nil {
		return dataCell["RuleName"].(string)
	}
	defer func() {
		if p := recover(); p != nil {
			logs.Log().Error("subNamespace: %v", p)
		}
	}()
	return c.Spider.SubNamespace(c.Spider, dataCell)
}

// joinNamespaces concatenates main and sub-namespace with double underscore.
func joinNamespaces(namespace, subNamespace string) string {
	if namespace == "" {
		return subNamespace
	} else if subNamespace != "" {
		return namespace + "__" + subNamespace
	}
	return namespace
}
