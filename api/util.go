package api

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/Sirupsen/logrus"
)

const (
	// Resource types
	RESOURCE_GROUP  = "group"
	RESOURCE_USER   = "user"
	RESOURCE_POLICY = "policy"

	// Constraints
	MAX_EXTERNAL_ID_LENGTH = 128
	MAX_NAME_LENGTH        = 128
	MAX_ACTION_LENGTH      = 128
	MAX_PATH_LENGTH        = 512
	MAX_RESOURCE_NUMBER    = 50
	MAX_LIMIT_SIZE         = 1000
	DEFAULT_LIMIT_SIZE     = 20

	// Actions

	// User actions
	USER_ACTION_CREATE_USER          = "iam:CreateUser"
	USER_ACTION_DELETE_USER          = "iam:DeleteUser"
	USER_ACTION_GET_USER             = "iam:GetUser"
	USER_ACTION_LIST_USERS           = "iam:ListUsers"
	USER_ACTION_UPDATE_USER          = "iam:UpdateUser"
	USER_ACTION_LIST_GROUPS_FOR_USER = "iam:ListGroupsForUser"

	// Group actions
	GROUP_ACTION_CREATE_GROUP                 = "iam:CreateGroup"
	GROUP_ACTION_DELETE_GROUP                 = "iam:DeleteGroup"
	GROUP_ACTION_GET_GROUP                    = "iam:GetGroup"
	GROUP_ACTION_LIST_GROUPS                  = "iam:ListGroups"
	GROUP_ACTION_UPDATE_GROUP                 = "iam:UpdateGroup"
	GROUP_ACTION_LIST_MEMBERS                 = "iam:ListMembers"
	GROUP_ACTION_ADD_MEMBER                   = "iam:AddMember"
	GROUP_ACTION_REMOVE_MEMBER                = "iam:RemoveMember"
	GROUP_ACTION_ATTACH_GROUP_POLICY          = "iam:AttachGroupPolicy"
	GROUP_ACTION_DETACH_GROUP_POLICY          = "iam:DetachGroupPolicy"
	GROUP_ACTION_LIST_ATTACHED_GROUP_POLICIES = "iam:ListAttachedGroupPolicies"

	// Policy actions
	POLICY_ACTION_CREATE_POLICY        = "iam:CreatePolicy"
	POLICY_ACTION_DELETE_POLICY        = "iam:DeletePolicy"
	POLICY_ACTION_UPDATE_POLICY        = "iam:UpdatePolicy"
	POLICY_ACTION_GET_POLICY           = "iam:GetPolicy"
	POLICY_ACTION_LIST_ATTACHED_GROUPS = "iam:ListAttachedGroups"
	POLICY_ACTION_LIST_POLICIES        = "iam:ListPolicies"
)

var (
	rUserExtID, _          = regexp.Compile(`^[\w+.@=\-_]+$`)
	rName, _               = regexp.Compile(`^[\w\-_]+$`)
	rOrder, _              = regexp.Compile(`^\w+\-(asc|desc)$`)
	rOrg, _                = regexp.Compile(`^[\w\-_]+$`)
	rPath, _               = regexp.Compile(`^/$|^/[\w+/\-_]+\w+/$`)
	rPathExclude, _        = regexp.Compile(`[/]{2,}`)
	rAction, _             = regexp.Compile(`^[\w\-_:]+[\w\-_*]+$`)
	rActionExclude, _      = regexp.Compile(`[*]{2,}|[:]{2,}`)
	rWordResource, _       = regexp.Compile(`^[\w+\-_.@]+$`)
	rWordResourcePrefix, _ = regexp.Compile(`^[\w+\-_.@]+\*$`)
	rUrn, _                = regexp.Compile(`^\*$|^[\w+\-@.]+\*?$|^[\w+\-@.]+\*?$|^[\w+\-@.]+(/?([\w+\-@.]+/)*([\w+\-@.]|[*])+)?$`)
	rUrnExclude, _         = regexp.Compile(`[/]{2,}|[:]{2,}|[*]{2,}`)
)

func CreateUrn(org string, resource string, path string, name string) string {
	switch resource {
	case RESOURCE_USER:
		return fmt.Sprintf("urn:iws:iam::user%v%v", path, name)
	default:
		return fmt.Sprintf("urn:iws:iam:%v:%v%v%v", org, resource, path, name)
	}
}

func GetUrnPrefix(org string, resource string, path string) string {
	switch resource {
	case RESOURCE_USER:
		return fmt.Sprintf("urn:iws:iam::user%v*", path)
	default:
		return fmt.Sprintf("urn:iws:iam:%v:%v%v*", org, resource, path)
	}
}

func IsValidUserExternalID(externalID string) bool {
	return rUserExtID.MatchString(externalID) && len(externalID) < MAX_EXTERNAL_ID_LENGTH
}

func IsValidOrg(org string) bool {
	return rOrg.MatchString(org) && len(org) < MAX_NAME_LENGTH
}

// IsValidName validates group and policy names
func IsValidName(name string) bool {
	return rName.MatchString(name) && len(name) < MAX_NAME_LENGTH
}

// IsValidOrder validates the OrderBy query param
func IsValidOrder(order string) bool {
	return rOrder.MatchString(order) && len(order) < MAX_NAME_LENGTH
}

func IsValidPath(path string) bool {
	return rPath.MatchString(path) && !rPathExclude.MatchString(path) && len(path) < MAX_PATH_LENGTH
}

func IsValidEffect(effect string) error {
	if effect != "allow" && effect != "deny" {
		return &Error{
			Code:    INVALID_PARAMETER_ERROR,
			Message: fmt.Sprintf("Invalid effect: %v - Only 'allow' and 'deny' accepted", effect),
		}
	}
	return nil
}

func AreValidActions(actions []string) error {

	for _, action := range actions {
		if !rAction.MatchString(action) || rActionExclude.MatchString(action) || len(action) > MAX_ACTION_LENGTH {
			return &Error{
				Code:    REGEX_NO_MATCH,
				Message: fmt.Sprintf("No regex match in action: %v", action),
			}
		}
	}
	return nil
}

func AreValidResources(resources []string) error {
	//err generator helper
	errFunc := func(resource string) error {
		return &Error{
			Code:    REGEX_NO_MATCH,
			Message: fmt.Sprintf("No regex match in resource: %v", resource),
		}
	}

	for _, resource := range resources {
		blocks := strings.Split(resource, ":")
		for n, block := range blocks {
			switch n {
			case 0:
				if len(blocks) < 2 { // This is the last block
					if block != "*" {
						return errFunc(resource)
					}
				} else {
					if block != "urn" {
						return errFunc(resource)
					}
				}
			case 1:
				if len(blocks) < 3 { // This is the last block
					if block != "*" && !rWordResourcePrefix.MatchString(block) {
						return errFunc(resource)
					}
				} else {
					if !rWordResource.MatchString(block) {
						return errFunc(resource)
					}
				}
			case 2:
				if len(blocks) < 4 { // This is the last block
					if block != "*" && !rWordResourcePrefix.MatchString(block) {
						return errFunc(resource)
					}
				} else {
					if !rWordResource.MatchString(block) {
						return errFunc(resource)
					}
				}
			case 3:
				if len(blocks) < 5 { // This is the last block
					if block != "*" && !rWordResourcePrefix.MatchString(block) {
						return errFunc(resource)
					}
				} else {
					if block != "" && !rWordResource.MatchString(block) {
						return errFunc(resource)
					}
				}
			case 4:
				if !rUrn.MatchString(block) || rUrnExclude.MatchString(block) {
					return errFunc(resource)
				}
			default:
				return &Error{
					Code:    INVALID_PARAMETER_ERROR,
					Message: fmt.Sprintf("Invalid resource definition: %v", resource),
				}

			}
		}
	}
	return nil
}

func AreValidStatements(statements *[]Statement) error {
	for _, statement := range *statements {
		err := IsValidEffect(statement.Effect)
		if err != nil {
			return err
		}

		// check actions
		if len(statement.Actions) < 1 {
			return &Error{
				Code:    INVALID_PARAMETER_ERROR,
				Message: "Empty actions",
			}
		}
		err = AreValidActions(statement.Actions)
		if err != nil {
			return err
		}

		// check resources
		if len(statement.Resources) < 1 {
			return &Error{
				Code:    INVALID_PARAMETER_ERROR,
				Message: "Empty resources",
			}
		}
		err = AreValidResources(statement.Resources)
		if err != nil {
			return err
		}
	}
	return nil
}

func validateFilter(filter *Filter, validColumns []string) error {
	if len(filter.Org) > 0 && !IsValidOrg(filter.Org) {
		return &Error{
			Code:    INVALID_PARAMETER_ERROR,
			Message: fmt.Sprintf("Invalid parameter: org %v", filter.Org),
		}
	}

	if len(filter.PathPrefix) == 0 {
		filter.PathPrefix = "/"
	} else if !IsValidPath(filter.PathPrefix) {
		return &Error{
			Code:    INVALID_PARAMETER_ERROR,
			Message: fmt.Sprintf("Invalid parameter: pathPrefix %v", filter.PathPrefix),
		}
	}

	if len(filter.GroupName) > 0 && !IsValidName(filter.GroupName) {
		return &Error{
			Code:    INVALID_PARAMETER_ERROR,
			Message: fmt.Sprintf("Invalid parameter: group %v", filter.GroupName),
		}
	}

	if len(filter.ExternalID) > 0 && !IsValidUserExternalID(filter.ExternalID) {
		return &Error{
			Code:    INVALID_PARAMETER_ERROR,
			Message: fmt.Sprintf("Invalid parameter: externalID %v", filter.ExternalID),
		}
	}

	if len(filter.PolicyName) > 0 && !IsValidName(filter.PolicyName) {
		return &Error{
			Code:    INVALID_PARAMETER_ERROR,
			Message: fmt.Sprintf("Invalid parameter: policy %v", filter.PolicyName),
		}
	}

	if filter.Limit == 0 {
		filter.Limit = DEFAULT_LIMIT_SIZE
	} else if filter.Limit > MAX_LIMIT_SIZE {
		return &Error{
			Code:    INVALID_PARAMETER_ERROR,
			Message: fmt.Sprintf("Invalid parameter: limit %v, max limit allowed: %v", filter.Limit, MAX_LIMIT_SIZE),
		}
	}

	if len(filter.OrderBy) > 0 {
		if !IsValidOrder(filter.OrderBy) {
			return &Error{
				Code:    INVALID_PARAMETER_ERROR,
				Message: fmt.Sprintf("Invalid parameter: OrderBy %v", filter.OrderBy),
			}
		} else {
			column := strings.Split(filter.OrderBy, "-")[0]
			for _, validCol := range validColumns {
				if column == validCol {
					// replace "-" with space to match GoRM syntax
					filter.OrderBy = strings.Replace(filter.OrderBy, "-", " ", 1)
					// if the column matches, finish func
					return nil
				}
			}

			return &Error{
				Code:    INVALID_PARAMETER_ERROR,
				Message: fmt.Sprintf("Invalid parameter: OrderBy column %v", column),
			}
		}
	}

	return nil
}

func LogOperation(logger *logrus.Logger, requestInfo RequestInfo, message string) {
	logger.WithFields(logrus.Fields{
		"requestID": requestInfo.RequestID,
		"userID":    requestInfo.Identifier,
	}).Info(message)
}
