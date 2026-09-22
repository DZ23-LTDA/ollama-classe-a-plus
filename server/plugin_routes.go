package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ollama/ollama/internal/agent"
)

func (a *agentAPI) requirePluginAdmin(c *gin.Context) bool {
	if !a.authRequired {
		return true
	}
	value, ok := c.Get("agent.membership")
	membership, ok := value.(agent.Membership)
	if !ok || (membership.Role != agent.RoleOwner && membership.Role != agent.RoleAdmin) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "plugin lifecycle requires organization admin"})
		return false
	}
	return true
}

func pluginCatalog(c *gin.Context, a *agentAPI) {
	c.JSON(http.StatusOK, gin.H{
		"connectors": a.runtime.Connectors(),
		"mcp":        a.runtime.MCPServers(),
		"remote_mcp": a.runtime.RemoteMCPServers(),
		"skills":     a.context.Skills(),
	})
}

func (a *agentAPI) enableConnector(c *gin.Context)  { a.connectorLifecycle(c, true, false) }
func (a *agentAPI) disableConnector(c *gin.Context) { a.connectorLifecycle(c, false, false) }
func (a *agentAPI) removeConnector(c *gin.Context)  { a.connectorLifecycle(c, false, true) }
func (a *agentAPI) connectorLifecycle(c *gin.Context, enabled, remove bool) {
	if !a.requirePluginAdmin(c) {
		return
	}
	var err error
	if remove {
		err = a.runtime.RemoveConnector(c.Param("id"))
	} else {
		err = a.runtime.SetConnectorEnabled(c.Param("id"), enabled)
	}
	if err != nil {
		writeAgentError(c, http.StatusNotFound, err)
		return
	}
	pluginCatalog(c, a)
}

func (a *agentAPI) enableMCP(c *gin.Context)  { a.mcpLifecycle(c, true, false) }
func (a *agentAPI) disableMCP(c *gin.Context) { a.mcpLifecycle(c, false, false) }
func (a *agentAPI) removeMCP(c *gin.Context)  { a.mcpLifecycle(c, false, true) }
func (a *agentAPI) mcpLifecycle(c *gin.Context, enabled, remove bool) {
	if !a.requirePluginAdmin(c) {
		return
	}
	var err error
	if remove {
		err = a.runtime.RemoveMCP(c.Param("id"))
	} else {
		err = a.runtime.SetMCPEnabled(c.Param("id"), enabled)
	}
	if err != nil {
		writeAgentError(c, http.StatusNotFound, err)
		return
	}
	pluginCatalog(c, a)
}

func (a *agentAPI) enableRemoteMCP(c *gin.Context)  { a.remoteMCPLifecycle(c, true, false) }
func (a *agentAPI) disableRemoteMCP(c *gin.Context) { a.remoteMCPLifecycle(c, false, false) }
func (a *agentAPI) removeRemoteMCP(c *gin.Context)  { a.remoteMCPLifecycle(c, false, true) }
func (a *agentAPI) remoteMCPLifecycle(c *gin.Context, enabled, remove bool) {
	if !a.requirePluginAdmin(c) {
		return
	}
	var err error
	if remove {
		err = a.runtime.RemoveRemoteMCP(c.Param("id"))
	} else {
		err = a.runtime.SetRemoteMCPEnabled(c.Param("id"), enabled)
	}
	if err != nil {
		writeAgentError(c, http.StatusNotFound, err)
		return
	}
	pluginCatalog(c, a)
}

func (a *agentAPI) enableSkill(c *gin.Context)  { a.skillLifecycle(c, true, false) }
func (a *agentAPI) disableSkill(c *gin.Context) { a.skillLifecycle(c, false, false) }
func (a *agentAPI) removeSkill(c *gin.Context)  { a.skillLifecycle(c, false, true) }
func (a *agentAPI) skillLifecycle(c *gin.Context, enabled, remove bool) {
	if !a.requirePluginAdmin(c) {
		return
	}
	var err error
	if remove {
		err = a.runtime.RemoveSkill(c.Param("id"))
	} else {
		err = a.runtime.SetSkillEnabled(c.Param("id"), enabled)
	}
	if err != nil {
		writeAgentError(c, http.StatusNotFound, err)
		return
	}
	pluginCatalog(c, a)
}
