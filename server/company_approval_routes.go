package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (a *agentAPI) decideCompanyApproval(c *gin.Context, resourceType, resourceParam string) {
	company, err := a.companyForRequest(c)
	if err != nil {
		writeAgentError(c, statusForAgentError(err), err)
		return
	}
	if !a.requireApprovalApprover(c) {
		return
	}
	var request struct {
		Approved bool   `json:"approved"`
		Nonce    string `json:"nonce"`
		Reason   string `json:"reason"`
	}
	if err := decodeJSON(c, &request); err != nil {
		writeAgentError(c, http.StatusBadRequest, err)
		return
	}
	approval, err := a.runtime.CompanyStore().PendingApproval(c.Param("id"), resourceType, c.Param(resourceParam))
	if err != nil {
		writeAgentError(c, statusForAgentError(err), err)
		return
	}
	organizationID := agentOrganizationID(c)
	if organizationID == "" {
		organizationID = company.OrganizationID
	}
	updated, err := a.runtime.CompanyStore().DecideApproval(c.Param("id"), approval.ID, request.Approved, request.Reason, agentActorID(c), organizationID, company.Version, request.Nonce)
	if err != nil {
		writeAgentError(c, statusForAgentError(err), err)
		return
	}
	c.JSON(http.StatusOK, updated)
}
