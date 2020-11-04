package role

import (
	"errors"
	"go-rest-skeleton/application"
	"go-rest-skeleton/domain/entity"
	"go-rest-skeleton/domain/repository"
	"go-rest-skeleton/infrastructure/authorization"
	"go-rest-skeleton/infrastructure/message/exception"
	"go-rest-skeleton/infrastructure/message/success"
	"go-rest-skeleton/pkg/response"
	"net/http"

	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
)

// Roles is a struct defines the dependencies that will be used.
type Roles struct {
	ur application.RoleAppInterface
	rd authorization.AuthInterface
	tk authorization.TokenInterface
}

// NewRoles is constructor will initialize role handler.
func NewRoles(
	ur application.RoleAppInterface,
	rd authorization.AuthInterface,
	tk authorization.TokenInterface) *Roles {
	return &Roles{
		ur: ur,
		rd: rd,
		tk: tk,
	}
}

// SaveRole is a function uses to handle create a new role.
func (s *Roles) SaveRole(c *gin.Context) {
	var roleEntity entity.Role
	if err := c.ShouldBindJSON(&roleEntity); err != nil {
		_ = c.AbortWithError(http.StatusUnprocessableEntity, exception.ErrorTextUnprocessableEntity)
		return
	}
	validateErr := roleEntity.ValidateSaveRole()
	if len(validateErr) > 0 {
		exceptionData := response.TranslateErrorForm(c, validateErr)
		c.Set("data", exceptionData)
		_ = c.AbortWithError(http.StatusUnprocessableEntity, exception.ErrorTextUnprocessableEntity)
		return
	}
	newRole, errDesc, errException := s.ur.SaveRole(&roleEntity)
	if errException != nil {
		c.Set("data", errDesc)
		if errors.Is(errException, exception.ErrorTextUnprocessableEntity) {
			_ = c.AbortWithError(http.StatusUnprocessableEntity, errException)
			return
		}
		_ = c.AbortWithError(http.StatusInternalServerError, exception.ErrorTextInternalServerError)
		return
	}
	c.Status(http.StatusCreated)
	response.NewSuccess(c, newRole.DetailRole(), success.RoleSuccessfullyCreateRole).JSON()
}

// UpdateUser is a function uses to handle create a new user.
func (s *Roles) UpdateRole(c *gin.Context) {
	var roleEntity entity.Role
	if err := c.ShouldBindUri(&roleEntity.UUID); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, exception.ErrorTextBadRequest)
		return
	}

	if err := c.ShouldBindJSON(&roleEntity); err != nil {
		_ = c.AbortWithError(http.StatusUnprocessableEntity, exception.ErrorTextUnprocessableEntity)
		return
	}

	validateErr := roleEntity.ValidateUpdateRole()
	if len(validateErr) > 0 {
		exceptionData := response.TranslateErrorForm(c, validateErr)
		c.Set("data", exceptionData)
		_ = c.AbortWithError(http.StatusUnprocessableEntity, exception.ErrorTextUnprocessableEntity)
		return
	}
	UUID := c.Param("uuid")
	updatedRole, errDesc, errException := s.ur.UpdateRole(UUID, &roleEntity)
	if errException != nil {
		c.Set("data", errDesc)
		if errors.Is(errException, exception.ErrorTextRoleNotFound) {
			_ = c.AbortWithError(http.StatusNotFound, errException)
			return
		}
		if errors.Is(errException, exception.ErrorTextUnprocessableEntity) {
			_ = c.AbortWithError(http.StatusUnprocessableEntity, errException)
			return
		}
		_ = c.AbortWithError(http.StatusInternalServerError, exception.ErrorTextInternalServerError)
		return
	}
	c.Status(http.StatusOK)
	response.NewSuccess(c, updatedRole.DetailRole(), success.RoleSuccessfullyUpdateRole).JSON()
}

// DeleteRole is a function uses to handle delete role by UUID.
func (s *Roles) DeleteRole(c *gin.Context) {
	var roleEntity entity.Role
	if err := c.ShouldBindUri(&roleEntity.UUID); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, exception.ErrorTextBadRequest)
		return
	}

	UUID := c.Param("uuid")
	err := s.ur.DeleteRole(UUID)
	if err != nil {
		if errors.Is(err, exception.ErrorTextUserNotFound) {
			_ = c.AbortWithError(http.StatusNotFound, exception.ErrorTextUserNotFound)
			return
		}
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	response.NewSuccess(c, nil, success.RoleSuccessfullyDeleteRole).JSON()
}

// GetRoles is a function uses to handle get role list.
func (s *Roles) GetRoles(c *gin.Context) {
	var role entity.Role
	var roles entity.Roles
	var err error
	parameters := repository.NewGinParameters(c)
	validateErr := parameters.ValidateParameter(role.FilterableFields()...)
	if len(validateErr) > 0 {
		exceptionData := response.TranslateErrorForm(c, validateErr)
		c.Set("data", exceptionData)
		_ = c.AbortWithError(http.StatusUnprocessableEntity, exception.ErrorTextUnprocessableEntity)
		return
	}

	roles, meta, err := s.ur.GetRoles(parameters)
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	response.NewSuccess(c, roles.DetailRoles(), success.RoleSuccessfullyGetRoleList).WithMeta(meta).JSON()
}

// GetRole is a function uses to handle get role detail by UUID.
func (s *Roles) GetRole(c *gin.Context) {
	var roleEntity entity.Role
	if err := c.ShouldBindUri(&roleEntity.UUID); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, exception.ErrorTextBadRequest)
		return
	}

	UUID := c.Param("uuid")
	role, err := s.ur.GetRoleWithPermissions(UUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			_ = c.AbortWithError(http.StatusNotFound, exception.ErrorTextUserNotFound)
			return
		}
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	response.NewSuccess(c, role.DetailRole(), success.RoleSuccessfullyGetRoleDetail).JSON()
}
