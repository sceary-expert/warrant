package authz

import (
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/warrant-dev/warrant/pkg/middleware"
	"github.com/warrant-dev/warrant/pkg/service"
)

// GetRoutes registers all route handlers for this module
func (svc FeatureService) Routes() []service.Route {
	return []service.Route{
		// create
		{
			Pattern: "/v1/features",
			Method:  "POST",
			Handler: service.NewRouteHandler(svc, CreateHandler),
		},

		// get
		{
			Pattern: "/v1/features",
			Method:  "GET",
			Handler: middleware.ChainMiddleware(
				service.NewRouteHandler(svc, ListHandler),
				middleware.ListMiddleware[FeatureListParamParser],
			),
		},
		{
			Pattern: "/v1/features/{featureId}",
			Method:  "GET",
			Handler: service.NewRouteHandler(svc, GetHandler),
		},

		// update
		{
			Pattern: "/v1/features/{featureId}",
			Method:  "POST",
			Handler: service.NewRouteHandler(svc, UpdateHandler),
		},
		{
			Pattern: "/v1/features/{featureId}",
			Method:  "PUT",
			Handler: service.NewRouteHandler(svc, UpdateHandler),
		},

		// delete
		{
			Pattern: "/v1/features/{featureId}",
			Method:  "DELETE",
			Handler: service.NewRouteHandler(svc, DeleteHandler),
		},
	}
}

func CreateHandler(svc FeatureService, w http.ResponseWriter, r *http.Request) error {
	var newFeature FeatureSpec
	err := service.ParseJSONBody(r.Body, &newFeature)
	if err != nil {
		return err
	}

	createdFeature, err := svc.Create(r.Context(), newFeature)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, createdFeature)
	return nil
}

func GetHandler(svc FeatureService, w http.ResponseWriter, r *http.Request) error {
	featureIdParam := mux.Vars(r)["featureId"]
	featureId, err := url.QueryUnescape(featureIdParam)
	if err != nil {
		return service.NewInvalidParameterError("featureId", "")
	}

	feature, err := svc.GetByFeatureId(r.Context(), featureId)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, feature)
	return nil
}

func ListHandler(svc FeatureService, w http.ResponseWriter, r *http.Request) error {
	listParams := middleware.GetListParamsFromContext(r.Context())
	features, err := svc.List(r.Context(), listParams)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, features)
	return nil
}

func UpdateHandler(svc FeatureService, w http.ResponseWriter, r *http.Request) error {
	var updateFeature UpdateFeatureSpec
	err := service.ParseJSONBody(r.Body, &updateFeature)
	if err != nil {
		return err
	}

	featureIdParam := mux.Vars(r)["featureId"]
	featureId, err := url.QueryUnescape(featureIdParam)
	if err != nil {
		return service.NewInvalidParameterError("featureId", "")
	}

	updatedFeature, err := svc.UpdateByFeatureId(r.Context(), featureId, updateFeature)
	if err != nil {
		return err
	}

	service.SendJSONResponse(w, updatedFeature)
	return nil
}

func DeleteHandler(svc FeatureService, w http.ResponseWriter, r *http.Request) error {
	featureId := mux.Vars(r)["featureId"]
	if featureId == "" {
		return service.NewMissingRequiredParameterError("featureId")
	}

	err := svc.DeleteByFeatureId(r.Context(), featureId)
	if err != nil {
		return err
	}

	return nil
}
