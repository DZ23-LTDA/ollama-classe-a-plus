package server

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ollama/ollama/internal/agent"
)

func companyOrganizationID(c *gin.Context) string {
	if organizationID := agentOrganizationID(c); organizationID != "" {
		return organizationID
	}
	return "local"
}

func (a *agentAPI) companyForRequest(c *gin.Context) (agent.Company, error) {
	company, err := a.runtime.CompanyStore().Get(c.Param("id"))
	if err != nil {
		return agent.Company{}, err
	}
	if organizationID := companyOrganizationID(c); company.OrganizationID != organizationID {
		return agent.Company{}, errAgentForbidden
	}
	return company, nil
}

func (a *agentAPI) companies(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"companies": a.runtime.CompanyStore().List(companyOrganizationID(c))})
}

func (a *agentAPI) createCompany(c *gin.Context) {
	var input agent.Company
	if err := decodeJSON(c, &input); err != nil {
		writeAgentError(c, http.StatusBadRequest, err)
		return
	}
	input.ID = ""
	input.OrganizationID = companyOrganizationID(c)
	company, err := a.runtime.CompanyStore().Create(input)
	if err != nil {
		writeAgentError(c, statusForAgentError(err), err)
		return
	}
	c.JSON(http.StatusCreated, company)
}

func (a *agentAPI) getCompany(c *gin.Context) {
	company, err := a.companyForRequest(c)
	if err != nil {
		writeAgentError(c, statusForAgentError(err), err)
		return
	}
	c.JSON(http.StatusOK, company)
}

func (a *agentAPI) updateCompany(c *gin.Context) {
	if _, err := a.companyForRequest(c); err != nil {
		writeAgentError(c, statusForAgentError(err), err)
		return
	}
	var input agent.CompanyUpdate
	if err := decodeJSON(c, &input); err != nil {
		writeAgentError(c, http.StatusBadRequest, err)
		return
	}
	company, err := a.runtime.CompanyStore().Update(c.Param("id"), input)
	if err != nil {
		writeAgentError(c, statusForAgentError(err), err)
		return
	}
	c.JSON(http.StatusOK, company)
}

func (a *agentAPI) companyReport(c *gin.Context) {
	if _, err := a.companyForRequest(c); err != nil {
		writeAgentError(c, statusForAgentError(err), err)
		return
	}
	report, err := a.runtime.CompanyStore().Report(c.Param("id"))
	if err != nil {
		writeAgentError(c, statusForAgentError(err), err)
		return
	}
	c.JSON(http.StatusOK, report)
}

func (a *agentAPI) addCompanyRoadmap(c *gin.Context) {
	if _, err := a.companyForRequest(c); err != nil {
		writeAgentError(c, statusForAgentError(err), err)
		return
	}
	var input agent.CompanyRoadmapItem
	if err := decodeJSON(c, &input); err != nil {
		writeAgentError(c, http.StatusBadRequest, err)
		return
	}
	company, err := a.runtime.CompanyStore().AddRoadmap(c.Param("id"), input)
	if err != nil {
		writeAgentError(c, statusForAgentError(err), err)
		return
	}
	c.JSON(http.StatusCreated, company)
}

func (a *agentAPI) addCompanyGoal(c *gin.Context) {
	if _, err := a.companyForRequest(c); err != nil {
		writeAgentError(c, statusForAgentError(err), err)
		return
	}
	var input agent.CompanyGoal
	if err := decodeJSON(c, &input); err != nil {
		writeAgentError(c, http.StatusBadRequest, err)
		return
	}
	company, err := a.runtime.CompanyStore().AddGoal(c.Param("id"), input)
	if err != nil {
		writeAgentError(c, statusForAgentError(err), err)
		return
	}
	c.JSON(http.StatusCreated, company)
}

func (a *agentAPI) addCompanyBacklog(c *gin.Context) {
	if _, err := a.companyForRequest(c); err != nil {
		writeAgentError(c, statusForAgentError(err), err)
		return
	}
	var input agent.CompanyBacklogItem
	if err := decodeJSON(c, &input); err != nil {
		writeAgentError(c, http.StatusBadRequest, err)
		return
	}
	company, err := a.runtime.CompanyStore().AddBacklog(c.Param("id"), input)
	if err != nil {
		writeAgentError(c, statusForAgentError(err), err)
		return
	}
	c.JSON(http.StatusCreated, company)
}

func (a *agentAPI) addCompanyCycle(c *gin.Context) {
	company, err := a.companyForRequest(c)
	if err != nil {
		writeAgentError(c, statusForAgentError(err), err)
		return
	}
	if company.Status == agent.CompanyPaused || company.Risk.Paused {
		writeAgentError(c, http.StatusLocked, agent.ErrCompanyPaused)
		return
	}
	var input agent.CompanyCycle
	if err := decodeJSON(c, &input); err != nil {
		writeAgentError(c, http.StatusBadRequest, err)
		return
	}
	created, err := a.runtime.CompanyStore().AddCycle(c.Param("id"), input)
	if err != nil {
		writeAgentError(c, statusForAgentError(err), err)
		return
	}
	if len(created.Cycles) == 0 {
		writeAgentError(c, http.StatusInternalServerError, errors.New("company cycle was not persisted"))
		return
	}
	cycle := created.Cycles[len(created.Cycles)-1]
	schedule, err := a.context.CreateSchedule(agent.Schedule{
		Objective:       "Company OS / " + company.Name + ": " + cycle.Objective,
		Model:           "",
		Workspace:       "company://" + company.ID,
		OrganizationID:  company.OrganizationID,
		IntervalSeconds: cycle.IntervalSeconds,
		Enabled:         true,
		NextRunAt:       cycle.NextRunAt,
	})
	if err != nil {
		writeAgentError(c, statusForAgentError(err), err)
		return
	}
	updated, err := a.runtime.CompanyStore().SetCycleSchedule(c.Param("id"), cycle.ID, schedule.ID)
	if err != nil {
		writeAgentError(c, statusForAgentError(err), err)
		return
	}
	c.JSON(http.StatusCreated, updated)
}

func (a *agentAPI) pauseCompany(c *gin.Context) {
	var input struct {
		Reason string `json:"reason"`
	}
	if err := decodeJSON(c, &input); err != nil {
		writeAgentError(c, http.StatusBadRequest, err)
		return
	}
	if _, err := a.companyForRequest(c); err != nil {
		writeAgentError(c, statusForAgentError(err), err)
		return
	}
	company, err := a.runtime.CompanyStore().Pause(c.Param("id"), input.Reason)
	if err != nil {
		writeAgentError(c, statusForAgentError(err), err)
		return
	}
	c.JSON(http.StatusOK, company)
}

func (a *agentAPI) resumeCompany(c *gin.Context) {
	if _, err := a.companyForRequest(c); err != nil {
		writeAgentError(c, statusForAgentError(err), err)
		return
	}
	company, err := a.runtime.CompanyStore().Resume(c.Param("id"))
	if err != nil {
		writeAgentError(c, statusForAgentError(err), err)
		return
	}
	c.JSON(http.StatusOK, company)
}

func (a *agentAPI) recordCompanyAnomaly(c *gin.Context) {
	if _, err := a.companyForRequest(c); err != nil {
		writeAgentError(c, statusForAgentError(err), err)
		return
	}
	var input struct {
		Severity string `json:"severity"`
		Reason   string `json:"reason"`
	}
	if err := decodeJSON(c, &input); err != nil {
		writeAgentError(c, http.StatusBadRequest, err)
		return
	}
	company, err := a.runtime.CompanyStore().RecordAnomaly(c.Param("id"), input.Severity, input.Reason)
	if err != nil {
		writeAgentError(c, statusForAgentError(err), err)
		return
	}
	c.JSON(http.StatusOK, company)
}

func (a *agentAPI) recordCompanySpend(c *gin.Context) {
	if _, err := a.companyForRequest(c); err != nil {
		writeAgentError(c, statusForAgentError(err), err)
		return
	}
	var input struct {
		Category    string `json:"category"`
		AmountCents int64  `json:"amount_cents"`
		Approved    bool   `json:"approved"`
	}
	if err := decodeJSON(c, &input); err != nil {
		writeAgentError(c, http.StatusBadRequest, err)
		return
	}
	company, err := a.runtime.CompanyStore().RecordSpend(c.Param("id"), input.Category, input.AmountCents, input.Approved)
	if err != nil {
		writeAgentError(c, statusForAgentError(err), err)
		return
	}
	c.JSON(http.StatusOK, company)
}

func companyErrorStatus(err error) int {
	switch {
	case errors.Is(err, agent.ErrCompanyNotFound):
		return http.StatusNotFound
	case errors.Is(err, agent.ErrCompanyBudgetExceeded), errors.Is(err, agent.ErrCompanyApprovalRequired):
		return http.StatusConflict
	case errors.Is(err, agent.ErrCompanyCampaignNotFound), errors.Is(err, agent.ErrCompanyAffiliateNotFound), errors.Is(err, agent.ErrCompanyProductNotFound), errors.Is(err, agent.ErrCompanyOrderNotFound):
		return http.StatusNotFound
	case errors.Is(err, agent.ErrCompanyApprovalRequiredForExternal):
		return http.StatusConflict
	case errors.Is(err, agent.ErrCompanyInvalidExternalURL):
		return http.StatusBadRequest
	case errors.Is(err, agent.ErrCompanyPaused):
		return http.StatusLocked
	case strings.Contains(strings.ToLower(err.Error()), "company"):
		return http.StatusBadRequest
	default:
		return 0
	}
}
