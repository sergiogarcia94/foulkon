package postgresql

import (
	"testing"
	"time"

	"github.com/Tecsisa/foulkon/api"
	"github.com/Tecsisa/foulkon/database"
	"github.com/kylelemons/godebug/pretty"
)

func TestPostgresRepo_AddPolicy(t *testing.T) {
	now := time.Now().UTC()
	testcases := map[string]struct {
		previousPolicy *Policy
		statements     []Statement
		policy         api.Policy
		// Expected result
		expectedResponse *api.Policy
		expectedError    *database.Error
	}{
		"OkCase": {
			policy: api.Policy{
				ID:       "test1",
				Name:     "test",
				Org:      "123",
				Path:     "/path/",
				CreateAt: now,
				UpdateAt: now,
				Urn:      api.CreateUrn("123", api.RESOURCE_POLICY, "/path/", "test"),
				Statements: &[]api.Statement{
					{
						Effect: "allow",
						Actions: []string{
							api.USER_ACTION_GET_USER,
						},
						Resources: []string{
							api.GetUrnPrefix("", api.RESOURCE_USER, "/path/"),
						},
					},
				},
			},
			expectedResponse: &api.Policy{
				ID:       "test1",
				Name:     "test",
				Org:      "123",
				Path:     "/path/",
				CreateAt: now,
				UpdateAt: now,
				Urn:      api.CreateUrn("123", api.RESOURCE_POLICY, "/path/", "test"),
				Statements: &[]api.Statement{
					{
						Effect: "allow",
						Actions: []string{
							api.USER_ACTION_GET_USER,
						},
						Resources: []string{
							api.GetUrnPrefix("", api.RESOURCE_USER, "/path/"),
						},
					},
				},
			},
		},
		"ErrorCaseAlreadyExists": {
			previousPolicy: &Policy{
				ID:       "test1",
				Name:     "test",
				Org:      "123",
				Path:     "/path/",
				CreateAt: now.UnixNano(),
				UpdateAt: now.UnixNano(),
				Urn:      api.CreateUrn("123", api.RESOURCE_POLICY, "/path/", "test"),
			},
			statements: []Statement{
				{
					ID:        "test1",
					PolicyID:  "test1",
					Effect:    "allow",
					Actions:   api.USER_ACTION_GET_USER,
					Resources: api.GetUrnPrefix("", api.RESOURCE_USER, "/path/"),
				},
			},
			policy: api.Policy{
				ID:       "test1",
				Name:     "test",
				Org:      "123",
				Path:     "/path/",
				CreateAt: now,
				Urn:      api.CreateUrn("123", api.RESOURCE_POLICY, "/path/", "test"),
				Statements: &[]api.Statement{
					{
						Effect: "allow",
						Actions: []string{
							api.USER_ACTION_GET_USER,
						},
						Resources: []string{
							api.GetUrnPrefix("", api.RESOURCE_USER, "/path/"),
						},
					},
				},
			},
			expectedError: &database.Error{
				Code:    database.INTERNAL_ERROR,
				Message: "pq: duplicate key value violates unique constraint \"policies_pkey\"",
			},
		},
	}

	for n, test := range testcases {
		// Clean policy database
		cleanPolicyTable()
		cleanStatementTable()

		// Call to repository to add a policy
		if test.previousPolicy != nil {
			err := insertPolicy(*test.previousPolicy, test.statements)
			if err != nil {
				t.Errorf("Test %v failed. Unexpected error: %v", n, err)
				continue
			}
		}
		receivedPolicy, err := repoDB.AddPolicy(test.policy)
		if test.expectedError != nil {
			dbError, ok := err.(*database.Error)
			if !ok || dbError == nil {
				t.Errorf("Test %v failed. Unexpected data retrieved from error: %v", n, err)
				continue
			}
			if diff := pretty.Compare(dbError, test.expectedError); diff != "" {
				t.Errorf("Test %v failed. Received different error response (received/wanted) %v", n, diff)
				continue
			}
		} else {
			if err != nil {
				t.Errorf("Test %v failed. Unexpected error: %v", n, err)
				continue
			}
			// Check response
			if diff := pretty.Compare(receivedPolicy, test.expectedResponse); diff != "" {
				t.Errorf("Test %v failed. Received different responses (received/wanted) %v", n, diff)
				continue
			}
			// Check database
			policyNumber, err := getPoliciesCountFiltered(test.policy.ID, test.policy.Org, test.policy.Name, test.policy.Path, test.policy.CreateAt.UnixNano(), test.policy.Urn)
			if err != nil {
				t.Errorf("Test %v failed. Unexpected error counting policies: %v", n, err)
				continue
			}
			if policyNumber != 1 {
				t.Errorf("Test %v failed. Received different policies number: %v", n, policyNumber)
				continue
			}
			for _, statement := range *test.policy.Statements {
				statementNumber, err := getStatementsCountFiltered(
					"",
					"",
					statement.Effect,
					stringArrayToString(statement.Actions),
					stringArrayToString(statement.Resources))
				if err != nil {
					t.Errorf("Test %v failed. Unexpected error counting statements: %v", n, err)
					continue
				}
				if statementNumber != 1 {
					t.Errorf("Test %v failed. Received different statements number: %v", n, statementNumber)
					continue
				}
			}
		}
	}
}

func TestPostgresRepo_GetPolicyByName(t *testing.T) {
	now := time.Now().UTC()
	testcases := map[string]struct {
		policy     *Policy
		statements []Statement
		// Postgres Repo Args
		org  string
		name string
		// Expected result
		expectedResponse *api.Policy
		expectedError    *database.Error
	}{
		"OkCase": {
			org:  "org1",
			name: "test",
			policy: &Policy{
				ID:       "1234",
				Name:     "test",
				Org:      "org1",
				Path:     "/path/",
				CreateAt: now.UnixNano(),
				UpdateAt: now.UnixNano(),
				Urn:      api.CreateUrn("org1", api.RESOURCE_POLICY, "/path/", "test"),
			},
			statements: []Statement{
				{
					ID:        "0123",
					Effect:    "allow",
					PolicyID:  "1234",
					Actions:   api.USER_ACTION_GET_USER,
					Resources: api.GetUrnPrefix("", api.RESOURCE_USER, "/path/"),
				},
			},
			expectedResponse: &api.Policy{
				ID:       "1234",
				Name:     "test",
				Org:      "org1",
				Path:     "/path/",
				CreateAt: now,
				UpdateAt: now,
				Urn:      api.CreateUrn("org1", api.RESOURCE_POLICY, "/path/", "test"),
				Statements: &[]api.Statement{
					{
						Effect: "allow",
						Actions: []string{
							api.USER_ACTION_GET_USER,
						},
						Resources: []string{
							api.GetUrnPrefix("", api.RESOURCE_USER, "/path/"),
						},
					},
				},
			},
		},
		"ErrorCaseNotFound": {
			org:  "org1",
			name: "test",
			expectedError: &database.Error{
				Code:    database.POLICY_NOT_FOUND,
				Message: "Policy with organization org1 and name test not found",
			},
		},
	}

	for n, test := range testcases {
		// Clean policy database
		cleanPolicyTable()
		cleanStatementTable()

		// Insert previous data
		if test.policy != nil {
			err := insertPolicy(*test.policy, test.statements)
			if err != nil {
				t.Errorf("Test %v failed. Error inserting policy/statements: %v", n, err)
			}
		}
		// Call to repository to get a policy
		receivedPolicy, err := repoDB.GetPolicyByName(test.org, test.name)
		if test.expectedError != nil {
			dbError, ok := err.(*database.Error)
			if !ok || dbError == nil {
				t.Errorf("Test %v failed. Unexpected data retrieved from error: %v", n, err)
				continue
			}
			if diff := pretty.Compare(dbError, test.expectedError); diff != "" {
				t.Errorf("Test %v failed. Received different error response (received/wanted) %v", n, diff)
				continue
			}
		} else {
			if err != nil {
				t.Errorf("Test %v failed. Unexpected error: %v", n, err)
				continue
			}
			// Check response
			if diff := pretty.Compare(receivedPolicy, test.expectedResponse); diff != "" {
				t.Errorf("Test %v failed. Received different responses (received/wanted) %v", n, diff)
				continue
			}
		}
	}
}

func TestPostgresRepo_GetPolicyById(t *testing.T) {
	now := time.Now().UTC()
	testcases := map[string]struct {
		policy     *Policy
		statements []Statement
		// Postgres Repo Args
		id string
		// Expected result
		expectedResponse *api.Policy
		expectedError    *database.Error
	}{
		"OkCase": {
			id: "1234",
			policy: &Policy{
				ID:       "1234",
				Name:     "test",
				Org:      "org1",
				Path:     "/path/",
				CreateAt: now.UnixNano(),
				UpdateAt: now.UnixNano(),
				Urn:      api.CreateUrn("org1", api.RESOURCE_POLICY, "/path/", "test"),
			},
			statements: []Statement{
				{
					ID:        "0123",
					Effect:    "allow",
					PolicyID:  "1234",
					Actions:   api.USER_ACTION_GET_USER,
					Resources: api.GetUrnPrefix("", api.RESOURCE_USER, "/path/"),
				},
			},
			expectedResponse: &api.Policy{
				ID:       "1234",
				Name:     "test",
				Org:      "org1",
				Path:     "/path/",
				CreateAt: now,
				UpdateAt: now,
				Urn:      api.CreateUrn("org1", api.RESOURCE_POLICY, "/path/", "test"),
				Statements: &[]api.Statement{
					{
						Effect: "allow",
						Actions: []string{
							api.USER_ACTION_GET_USER,
						},
						Resources: []string{
							api.GetUrnPrefix("", api.RESOURCE_USER, "/path/"),
						},
					},
				},
			},
		},
		"ErrorCaseNotFound": {
			id: "1234",
			expectedError: &database.Error{
				Code:    database.POLICY_NOT_FOUND,
				Message: "Policy with id 1234 not found",
			},
		},
	}

	for n, test := range testcases {
		// Clean policy database
		cleanPolicyTable()
		cleanStatementTable()

		// Insert previous data
		if test.policy != nil {
			err := insertPolicy(*test.policy, test.statements)
			if err != nil {
				t.Errorf("Test %v failed. Error inserting policy/statements: %v", n, err)
			}
		}
		// Call to repository to get a policy
		receivedPolicy, err := repoDB.GetPolicyById(test.id)
		if test.expectedError != nil {
			dbError, ok := err.(*database.Error)
			if !ok || dbError == nil {
				t.Errorf("Test %v failed. Unexpected data retrieved from error: %v", n, err)
				continue
			}
			if diff := pretty.Compare(dbError, test.expectedError); diff != "" {
				t.Errorf("Test %v failed. Received different error response (received/wanted) %v", n, diff)
				continue
			}
		} else {
			if err != nil {
				t.Errorf("Test %v failed. Unexpected error: %v", n, err)
				continue
			}
			// Check response
			if diff := pretty.Compare(receivedPolicy, test.expectedResponse); diff != "" {
				t.Errorf("Test %v failed. Received different responses (received/wanted) %v", n, diff)
				continue
			}
		}

	}
}

func TestPostgresRepo_GetPoliciesFiltered(t *testing.T) {
	now := time.Now().UTC()
	testcases := map[string]struct {
		policies   []Policy
		statements []Statement
		// Postgres Repo Args
		filter *api.Filter
		// Expected result
		expectedResponse []api.Policy
	}{
		"OkCase": {
			filter: &api.Filter{
				PathPrefix: "/",
				Org:        "org1",
				Offset:     0,
				Limit:      20,
				OrderBy:    "name desc",
			},
			policies: []Policy{
				{
					ID:       "111",
					Name:     "test1",
					Org:      "org1",
					Path:     "/path1/",
					CreateAt: now.UnixNano(),
					UpdateAt: now.UnixNano(),
					Urn:      api.CreateUrn("org1", api.RESOURCE_POLICY, "/path1/", "test1"),
				}, {
					ID:       "222",
					Name:     "test2",
					Org:      "org1",
					Path:     "/path2/",
					CreateAt: now.UnixNano(),
					UpdateAt: now.UnixNano(),
					Urn:      api.CreateUrn("org1", api.RESOURCE_POLICY, "/path2/", "test2"),
				},
			},
			statements: []Statement{
				{
					ID:        "1",
					Effect:    "allow",
					PolicyID:  "111",
					Actions:   api.USER_ACTION_GET_USER,
					Resources: api.GetUrnPrefix("", api.RESOURCE_USER, "/path1/"),
				},
				{
					ID:        "2",
					Effect:    "allow",
					PolicyID:  "222",
					Actions:   api.USER_ACTION_GET_USER,
					Resources: api.GetUrnPrefix("", api.RESOURCE_USER, "/path2/"),
				},
			},
			expectedResponse: []api.Policy{
				{
					ID:       "222",
					Name:     "test2",
					Org:      "org1",
					Path:     "/path2/",
					CreateAt: now,
					UpdateAt: now,
					Urn:      api.CreateUrn("org1", api.RESOURCE_POLICY, "/path2/", "test2"),
					Statements: &[]api.Statement{
						{
							Effect: "allow",
							Actions: []string{
								api.USER_ACTION_GET_USER,
							},
							Resources: []string{
								api.GetUrnPrefix("", api.RESOURCE_USER, "/path2/"),
							},
						},
					},
				},
				{
					ID:       "111",
					Name:     "test1",
					Org:      "org1",
					Path:     "/path1/",
					CreateAt: now,
					UpdateAt: now,
					Urn:      api.CreateUrn("org1", api.RESOURCE_POLICY, "/path1/", "test1"),
					Statements: &[]api.Statement{
						{
							Effect: "allow",
							Actions: []string{
								api.USER_ACTION_GET_USER,
							},
							Resources: []string{
								api.GetUrnPrefix("", api.RESOURCE_USER, "/path1/"),
							},
						},
					},
				},
			},
		},
		"OKCaseNotFound": {
			filter: &api.Filter{
				PathPrefix: "test",
				Org:        "org1",
				Offset:     0,
				Limit:      20,
			},
			expectedResponse: []api.Policy{},
		},
	}

	for n, test := range testcases {
		// Clean policy database
		cleanPolicyTable()
		cleanStatementTable()

		// Insert previous data
		for i, policy := range test.policies {
			var statement []Statement = []Statement{test.statements[i]}
			err := insertPolicy(policy, statement)
			if err != nil {
				t.Errorf("Test %v failed. Error inserting policy/statements: %v", n, err)
			}
		}
		// Call to repository to get a policy
		receivedPolicy, total, err := repoDB.GetPoliciesFiltered(test.filter)
		if err != nil {
			t.Errorf("Test %v failed. Unexpected error: %v", n, err)
			continue
		}
		// Check total
		if total != len(test.expectedResponse) {
			t.Errorf("Test %v failed. Received different total elements: %v", n, total)
			continue
		}
		// Check response
		if diff := pretty.Compare(receivedPolicy, test.expectedResponse); diff != "" {
			t.Errorf("Test %v failed. Received different responses (received/wanted) %v", n, diff)
			continue
		}
	}
}

func TestPostgresRepo_UpdatePolicy(t *testing.T) {
	now := time.Now().UTC()
	testcases := map[string]struct {
		previousPolicy *api.Policy
		policy         api.Policy
		// Expected result
		expectedResponse *api.Policy
	}{
		"OkCase": {
			previousPolicy: &api.Policy{
				ID:       "test1",
				Name:     "test",
				Org:      "123",
				Path:     "/path/",
				CreateAt: now,
				Urn:      api.CreateUrn("123", api.RESOURCE_POLICY, "/path/", "test"),
				Statements: &[]api.Statement{
					{
						Effect: "allow",
						Actions: []string{
							api.USER_ACTION_GET_USER,
						},
						Resources: []string{
							api.GetUrnPrefix("", api.RESOURCE_USER, "/path/"),
						},
					},
				},
			},
			policy: api.Policy{
				ID:       "test1",
				Name:     "newName",
				Org:      "123",
				Path:     "/newPath/",
				CreateAt: now,
				Urn:      api.CreateUrn("123", api.RESOURCE_POLICY, "/newPath/", "newName"),
				Statements: &[]api.Statement{
					{
						Effect: "allow",
						Actions: []string{
							api.USER_ACTION_GET_USER,
						},
						Resources: []string{
							api.GetUrnPrefix("123", api.RESOURCE_USER, "/newPath/"),
						},
					},
				},
			},
			expectedResponse: &api.Policy{
				ID:       "test1",
				Name:     "newName",
				Org:      "123",
				Path:     "/newPath/",
				CreateAt: now,
				Urn:      api.CreateUrn("123", api.RESOURCE_POLICY, "/newPath/", "newName"),
				Statements: &[]api.Statement{
					{
						Effect: "allow",
						Actions: []string{
							api.USER_ACTION_GET_USER,
						},
						Resources: []string{
							api.GetUrnPrefix("123", api.RESOURCE_USER, "/newPath/"),
						},
					},
				},
			},
		},
	}

	for n, test := range testcases {
		// Clean policy database
		cleanPolicyTable()
		cleanStatementTable()

		// Call to repository to add a policy
		if test.previousPolicy != nil {
			_, err := repoDB.AddPolicy(*test.previousPolicy)
			if err != nil {
				t.Errorf("Test %v failed. Unexpected error: %v", n, err)
				continue
			}
		}
		receivedPolicy, err := repoDB.UpdatePolicy(test.policy)
		if err != nil {
			t.Errorf("Test %v failed. Unexpected error: %v", n, err)
			continue
		}
		// Check response
		if diff := pretty.Compare(receivedPolicy, test.expectedResponse); diff != "" {
			t.Errorf("Test %v failed. Received different responses (received/wanted) %v", n, diff)
			continue
		}
	}
}

func TestPostgresRepo_RemovePolicy(t *testing.T) {
	type relation struct {
		policyID      string
		groupID       string
		createAt      int64
		groupNotFound bool
	}
	type policyData struct {
		policy     Policy
		statements []Statement
	}
	now := time.Now().UTC()
	testcases := map[string]struct {
		// Previous data
		previousPolicies []policyData
		relations        []relation
		// Postgres Repo Args
		policyToDelete string
	}{
		"OkCase": {
			previousPolicies: []policyData{
				{
					policy: Policy{
						ID:       "test1",
						Name:     "test1",
						Org:      "123",
						Path:     "/path/",
						CreateAt: now.UnixNano(),
						UpdateAt: now.UnixNano(),
						Urn:      api.CreateUrn("123", api.RESOURCE_POLICY, "/path/", "test1"),
					},
					statements: []Statement{
						{
							ID:        "test1",
							PolicyID:  "test1",
							Effect:    "allow",
							Actions:   api.USER_ACTION_GET_USER,
							Resources: api.GetUrnPrefix("", api.RESOURCE_USER, "/path/"),
						},
					},
				},
				{
					policy: Policy{
						ID:       "test2",
						Name:     "test2",
						Org:      "123",
						Path:     "/path/",
						CreateAt: now.UnixNano(),
						UpdateAt: now.UnixNano(),
						Urn:      api.CreateUrn("123", api.RESOURCE_POLICY, "/path/", "test2"),
					},
					statements: []Statement{
						{
							ID:        "test2",
							PolicyID:  "test2",
							Effect:    "allow",
							Actions:   api.USER_ACTION_GET_USER,
							Resources: api.GetUrnPrefix("", api.RESOURCE_USER, "/path/"),
						},
					},
				},
			},
			relations: []relation{
				{
					policyID: "test1",
					groupID:  "GroupID",
					createAt: now.UnixNano(),
				},
				{
					policyID: "test2",
					groupID:  "GroupID2",
					createAt: now.UnixNano(),
				},
			},
			policyToDelete: "test1",
		},
	}

	for n, test := range testcases {
		// Clean policy database
		cleanPolicyTable()
		cleanStatementTable()
		cleanGroupTable()
		cleanGroupPolicyRelationTable()

		// insert previous policy
		if test.previousPolicies != nil {
			for _, p := range test.previousPolicies {
				err := insertPolicy(p.policy, p.statements)
				if err != nil {
					t.Errorf("Test %v failed. Unexpected error inserting previous policies: %v", n, err)
					continue
				}
			}
		}
		if test.relations != nil {
			for _, rel := range test.relations {
				err := insertGroupPolicyRelation(rel.groupID, rel.policyID, rel.createAt)
				if err != nil {
					t.Errorf("Test %v failed. Unexpected error inserting group relation: %v", n, err)
					continue
				}
			}
		}
		err := repoDB.RemovePolicy(test.policyToDelete)
		if err != nil {
			t.Errorf("Test %v failed. Unexpected error: %v", n, err)
			continue
		}
		// Check database
		policyNumber, err := getPoliciesCountFiltered(test.policyToDelete, "", "", "", 0, "")
		if err != nil {
			t.Errorf("Test %v failed. Unexpected error counting policies: %v", n, err)
			continue
		}
		if policyNumber != 0 {
			t.Errorf("Test %v failed. Received different policies number: %v", n, policyNumber)
			continue
		}

		statementNumber, err := getStatementsCountFiltered(
			"",
			test.policyToDelete,
			"",
			"",
			"")
		if err != nil {
			t.Errorf("Test %v failed. Unexpected error counting statements: %v", n, err)
			continue
		}
		if statementNumber != 0 {
			t.Errorf("Test %v failed. Received different statements number: %v", n, statementNumber)
			continue
		}

		// Check total policy number
		totalPolicyNumber, err := getPoliciesCountFiltered("", "", "", "", 0, "")
		if err != nil {
			t.Errorf("Test %v failed. Unexpected error counting total policies: %v", n, err)
			continue
		}
		if totalPolicyNumber != 1 {
			t.Errorf("Test %v failed. Received different total policy number: %v", n, totalPolicyNumber)
			continue
		}

		totalStatementNumber, err := getStatementsCountFiltered("", "", "", "", "")
		if err != nil {
			t.Errorf("Test %v failed. Unexpected error counting statements: %v", n, err)
			continue
		}
		if totalStatementNumber != 1 {
			t.Errorf("Test %v failed. Received different total statements number: %v", n, totalStatementNumber)
			continue
		}

		totalGroupPolicyRelationNumber, err := getGroupPolicyRelationCount("", "")
		if err != nil {
			t.Errorf("Test %v failed. Unexpected error counting total group policy relations: %v", n, err)
			continue
		}
		if totalGroupPolicyRelationNumber != 1 {
			t.Errorf("Test %v failed. Received different total relations group policy number: %v", n, totalGroupPolicyRelationNumber)
			continue
		}
	}
}

func TestPostgresRepo_GetAttachedGroups(t *testing.T) {
	now := time.Now().UTC()
	testcases := map[string]struct {
		previousPolicy   *Policy
		statements       []Statement
		filter           *api.Filter
		group            []Group
		createAt         []int64
		expectedResponse []PolicyGroup
	}{
		"OkCase": {
			previousPolicy: &Policy{
				ID:       "test1",
				Name:     "test",
				Org:      "123",
				Path:     "/path/",
				CreateAt: now.UnixNano(),
				UpdateAt: now.UnixNano(),
				Urn:      api.CreateUrn("123", api.RESOURCE_POLICY, "/path/", "test"),
			},
			statements: []Statement{
				{
					ID:        "test1",
					PolicyID:  "test1",
					Effect:    "allow",
					Actions:   api.USER_ACTION_GET_USER,
					Resources: api.GetUrnPrefix("", api.RESOURCE_USER, "/path/"),
				},
			},
			filter: &api.Filter{
				OrderBy: "create_at desc",
			},
			group: []Group{
				{
					ID:       "GroupID1",
					Name:     "Name1",
					Path:     "Path1",
					Urn:      "urn1",
					CreateAt: now.UnixNano(),
					UpdateAt: now.UnixNano(),
					Org:      "Org1",
				},
				{
					ID:       "GroupID2",
					Name:     "Name2",
					Path:     "Path2",
					Urn:      "urn2",
					CreateAt: now.UnixNano(),
					UpdateAt: now.UnixNano(),
					Org:      "Org2",
				},
			},
			createAt: []int64{now.UnixNano() - 1, now.UnixNano()},
			expectedResponse: []PolicyGroup{
				{
					Group: &api.Group{
						ID:       "GroupID2",
						Name:     "Name2",
						Path:     "Path2",
						Urn:      "urn2",
						CreateAt: now,
						UpdateAt: now,
						Org:      "Org2",
					},
					CreateAt: now,
				},
				{
					Group: &api.Group{
						ID:       "GroupID1",
						Name:     "Name1",
						Path:     "Path1",
						Urn:      "urn1",
						CreateAt: now,
						UpdateAt: now,
						Org:      "Org1",
					},
					CreateAt: now.Add(-1),
				},
			},
		},
	}

	for n, test := range testcases {
		// Clean database
		cleanPolicyTable()
		cleanStatementTable()
		cleanGroupTable()
		cleanGroupPolicyRelationTable()

		// Call to repository to add a policy
		err := insertPolicy(*test.previousPolicy, test.statements)
		if err != nil {
			t.Errorf("Test %v failed. Unexpected error: %v", n, err)
			continue
		}
		for i, group := range test.group {
			err := insertGroup(group)
			if err != nil {
				t.Errorf("Test %v failed. Unexpected error inserting group: %v", n, err)
				continue
			}
			err = insertGroupPolicyRelation(group.ID, test.previousPolicy.ID, test.createAt[i])
			if err != nil {
				t.Errorf("Test %v failed. Unexpected error inserting group relation: %v", n, err)
				continue
			}
		}

		groups, total, err := repoDB.GetAttachedGroups(test.previousPolicy.ID, test.filter)
		if err != nil {
			t.Errorf("Test %v failed. Unexpected error: %v", n, err)
			continue
		}
		// Check total
		if total != len(test.expectedResponse) {
			t.Errorf("Test %v failed. Received different total elements: %v", n, total)
			continue
		}
		// Check response
		if diff := pretty.Compare(groups, test.expectedResponse); diff != "" {
			t.Errorf("Test %v failed. Received different responses (received/wanted) %v", n, diff)
			continue
		}
	}
}

func Test_dbPolicyToAPIPolicy(t *testing.T) {
	now := time.Now().UTC()
	testcases := map[string]struct {
		dbPolicy  *Policy
		apiPolicy *api.Policy
	}{
		"OkCase": {
			dbPolicy: &Policy{
				ID:       "test1",
				Name:     "test",
				Org:      "123",
				Path:     "/path/",
				CreateAt: now.UnixNano(),
				UpdateAt: now.UnixNano(),
				Urn:      api.CreateUrn("123", api.RESOURCE_POLICY, "/path/", "test"),
			},
			apiPolicy: &api.Policy{
				ID:       "test1",
				Name:     "test",
				Org:      "123",
				Path:     "/path/",
				CreateAt: now,
				UpdateAt: now,
				Urn:      api.CreateUrn("123", api.RESOURCE_POLICY, "/path/", "test"),
			},
		},
	}

	for n, test := range testcases {
		receivedAPIPolicy := dbPolicyToAPIPolicy(test.dbPolicy)
		// Check response
		if diff := pretty.Compare(receivedAPIPolicy, test.apiPolicy); diff != "" {
			t.Errorf("Test %v failed. Received different responses (received/wanted) %v", n, diff)
			continue
		}
	}
}

func Test_dbStatementsToAPIStatements(t *testing.T) {
	testcases := map[string]struct {
		dbStatements  []Statement
		apiStatements *[]api.Statement
	}{
		"OkCase": {
			dbStatements: []Statement{
				{
					ID:        "0123",
					Effect:    "allow",
					PolicyID:  "1234",
					Actions:   api.USER_ACTION_GET_USER,
					Resources: api.GetUrnPrefix("", api.RESOURCE_USER, "/path/"),
				},
			},
			apiStatements: &[]api.Statement{
				{
					Effect: "allow",
					Actions: []string{
						api.USER_ACTION_GET_USER,
					},
					Resources: []string{
						api.GetUrnPrefix("", api.RESOURCE_USER, "/path/"),
					},
				},
			},
		},
		"OkCase2": {
			dbStatements: []Statement{
				{
					ID:        "0123",
					Effect:    "allow",
					PolicyID:  "1234",
					Actions:   api.USER_ACTION_GET_USER,
					Resources: api.GetUrnPrefix("", api.RESOURCE_USER, "/path/"),
				},
				{
					ID:        "4321",
					Effect:    "deny",
					PolicyID:  "1234",
					Actions:   api.GROUP_ACTION_GET_GROUP + ";" + api.GROUP_ACTION_CREATE_GROUP,
					Resources: api.GetUrnPrefix("", api.RESOURCE_GROUP, "/xxx/") + ";" + api.GetUrnPrefix("", api.RESOURCE_GROUP, "/xxx2/"),
				},
			},
			apiStatements: &[]api.Statement{
				{
					Effect: "allow",
					Actions: []string{
						api.USER_ACTION_GET_USER,
					},
					Resources: []string{
						api.GetUrnPrefix("", api.RESOURCE_USER, "/path/"),
					},
				},
				{
					Effect: "deny",
					Actions: []string{
						api.GROUP_ACTION_GET_GROUP,
						api.GROUP_ACTION_CREATE_GROUP,
					},
					Resources: []string{
						api.GetUrnPrefix("", api.RESOURCE_GROUP, "/xxx/"),
						api.GetUrnPrefix("", api.RESOURCE_GROUP, "/xxx2/"),
					},
				},
			},
		},
	}

	for n, test := range testcases {
		receivedAPIStatements := dbStatementsToAPIStatements(test.dbStatements)
		// Check response
		if diff := pretty.Compare(receivedAPIStatements, test.apiStatements); diff != "" {
			t.Errorf("Test %v failed. Received different responses (received/wanted) %v", n, diff)
			continue
		}
	}
}

func Test_stringArrayToString(t *testing.T) {
	testcases := map[string]struct {
		arrayString    []string
		expectedString string
	}{
		"OkCase": {
			arrayString: []string{
				"asd",
				"123",
				"456",
				"zxc",
			},
			expectedString: "asd;123;456;zxc",
		},
		"OkCase2": {
			arrayString:    []string{},
			expectedString: "",
		},
	}

	for n, test := range testcases {
		receivedString := stringArrayToString(test.arrayString)
		// Check response
		if diff := pretty.Compare(receivedString, test.expectedString); diff != "" {
			t.Errorf("Test %v failed. Received different responses (received/wanted) %v", n, diff)
			continue
		}
	}
}
