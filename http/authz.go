package http

import (
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"github.com/tecsisa/authorizr/api"
	"net/http"
)

// Requests

type AuthorizeResourcesRequest struct {
	Action    string   `json:", omitempty"`
	Resources []string `json:", omitempty"`
}

// Responses

type AuthorizeResourcesResponse struct {
	ResourcesAllowed []string `json:"ResourcesAllowed, omitempty"`
}

func (a *WorkerHandler) HandleAuthorizeResources(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	userID := a.worker.Authenticator.RetrieveUserID(*r)

	// Decode request
	request := AuthorizeResourcesRequest{}
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		a.worker.Logger.Errorln(err)
		a.RespondBadRequest(r, &userID, w, &api.Error{Code: api.INVALID_PARAMETER_ERROR, Message: err.Error()})
		return
	}

	// Retrieve allowed resources
	a.worker.Logger.Debugf("Action %v, Resources %v", request.Action, request.Resources)
	result, err := a.worker.AuthzApi.GetAuthorizedExternalResources(userID, request.Action, request.Resources)
	if err != nil {
		a.worker.Logger.Errorln(err)
		// Transform to API errors
		apiError := err.(*api.Error)
		switch apiError.Code {
		case api.INVALID_PARAMETER_ERROR:
			a.RespondBadRequest(r, &userID, w, apiError)
		case api.UNAUTHORIZED_RESOURCES_ERROR:
			a.RespondForbidden(r, &userID, w, apiError)
		default: // Unexpected API error
			a.RespondInternalServerError(r, &userID, w)
		}
		return
	}

	response := AuthorizeResourcesResponse{
		ResourcesAllowed: result,
	}

	a.RespondOk(r, &userID, w, response)

}
