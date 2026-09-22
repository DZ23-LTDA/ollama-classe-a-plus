package agent

import "errors"

var ErrPluginOrganizationScope = errors.New("plugin is not owned by the requested organization")

func pluginOwnedByOrganization(owner, organizationID string) bool {
	return owner != "" && organizationID != "" && owner == organizationID
}

func pluginAccessibleByOrganization(owner, organizationID string) bool {
	if owner == "" {
		return true
	}
	return pluginOwnedByOrganization(owner, organizationID)
}
