package validation

import (
	"context"
	"errors"

	"github.com/openfga/openfga/pkg/tuple"
	"github.com/openfga/openfga/pkg/utils"
	"github.com/openfga/openfga/storage"
	openfgapb "go.buf.build/openfga/go/openfga/api/openfga/v1"
)

// ValidateTuple returns whether a *openfgapb.TupleKey is valid
func ValidateTuple(ctx context.Context, backend storage.TypeDefinitionReadBackend, store, authorizationModelID string, tk *openfgapb.TupleKey, dbCallsCounter utils.DBCallCounter) (*openfgapb.Userset, error) {
	if err := tuple.ValidateUser(tk); err != nil {
		return nil, err
	}
	return ValidateObjectsRelations(ctx, backend, store, authorizationModelID, tk, dbCallsCounter)
}

// ValidateObjectsRelations returns whether a tuple's object and relations are valid
func ValidateObjectsRelations(ctx context.Context, backend storage.TypeDefinitionReadBackend, store, modelID string, t *openfgapb.TupleKey, dbCallsCounter utils.DBCallCounter) (*openfgapb.Userset, error) {
	if !tuple.IsValidRelation(t.GetRelation()) {
		return nil, &tuple.InvalidTupleError{Reason: "invalid relation", TupleKey: t}
	}
	if !tuple.IsValidObject(t.GetObject()) {
		return nil, &tuple.InvalidObjectFormatError{TupleKey: t}
	}
	objectType, objectID := tuple.SplitObject(t.GetObject())
	if objectType == "" || objectID == "" {
		return nil, &tuple.InvalidObjectFormatError{TupleKey: t}
	}

	dbCallsCounter.AddReadCall()
	ns, err := backend.ReadTypeDefinition(ctx, store, modelID, objectType)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, &tuple.TypeNotFoundError{TypeName: objectType}
		}
		return nil, err
	}
	userset, ok := ns.Relations[t.Relation]
	if !ok {
		return nil, &tuple.RelationNotFoundError{Relation: t.GetRelation(), TypeName: ns.GetType(), TupleKey: t}
	}
	return userset, nil
}
