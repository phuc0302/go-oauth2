package mongo

import (
	"fmt"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// AllEntities returns an entity collection sort by id.
func AllEntities(tableName string, list interface{}) error {
	/* Condition validation */
	if len(tableName) == 0 {
		return fmt.Errorf("Invalid table name.")
	} else if list == nil {
		return fmt.Errorf("Invalid list object.")
	}
	session, database := GetMonotonicSession()
	defer session.Close()

	collection := database.C(tableName)
	err := collection.Find(nil).Sort("_id").All(list)

	return err
}

// AllEntities returns an entity collection base on criterion sort by id.
func AllEntitiesWithCriteria(tableName string, criterion map[string]interface{}, list interface{}) error {
	/* Condition validation */
	if len(tableName) == 0 {
		return fmt.Errorf("Invalid table name.")
	} else if criterion == nil || len(criterion) == 0 {
		return fmt.Errorf("Invalid criterion object.")
	} else if list == nil {
		return fmt.Errorf("Invalid list object.")
	}
	session, database := GetMonotonicSession()
	defer session.Close()

	collection := database.C(tableName)
	err := collection.Find(criterion).Sort("_id").All(list)

	return err
}

// EntityWithID finds entity with ID.
func EntityWithID(tableName string, entityID bson.ObjectId, entity interface{}) error {
	/* Condition validation */
	if len(tableName) == 0 {
		return fmt.Errorf("Invalid table name.")
	} else if entity == nil {
		return fmt.Errorf("Invalid entity object.")
	}
	session, database := GetMonotonicSession()
	defer session.Close()

	collection := database.C(tableName)
	err := collection.FindId(entityID).One(entity)

	return err
}

// EntityWithCriteria finds entity with creteria.
func EntityWithCriteria(tableName string, criterion map[string]interface{}, entity interface{}) error {
	/* Condition validation */
	if len(tableName) == 0 {
		return fmt.Errorf("Invalid table name.")
	} else if criterion == nil || len(criterion) == 0 {
		return fmt.Errorf("Invalid criterion object.")
	} else if entity == nil {
		return fmt.Errorf("Invalid entity object.")
	}
	session, database := GetMonotonicSession()
	defer session.Close()

	collection := database.C(tableName)
	err := collection.Find(criterion).One(entity)

	return err
}

// SaveEntity inserts/updates an entity.
func SaveEntity(tableName string, entityID bson.ObjectId, entity interface{}) error {
	/* Condition validation */
	if len(tableName) == 0 {
		return fmt.Errorf("Invalid table name.")
	} else if entity == nil {
		return fmt.Errorf("Invalid entity object.")
	}
	session, database := GetMonotonicSession()
	defer session.Close()

	session.SetSafe(&mgo.Safe{})
	collection := database.C(tableName)

	_, err := collection.UpsertId(entityID, entity)
	return err
}

// DeleteEntity deletes a record from collection.
func DeleteEntity(tableName string, entityID bson.ObjectId) error {
	/* Condition validation */
	if len(tableName) == 0 {
		return fmt.Errorf("Invalid table name.")
	}
	session, database := GetMonotonicSession()
	defer session.Close()

	collection := database.C(tableName)
	err := collection.RemoveId(entityID)

	return err
}
