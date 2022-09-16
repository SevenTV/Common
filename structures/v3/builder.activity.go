package structures

import (
	"time"

	"github.com/seventv/common/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ActivityBuilder struct {
	Update   UpdateMap
	Activity Activity

	initial Activity
}

func NewActivityBuilder(activity Activity) *ActivityBuilder {
	return &ActivityBuilder{
		Update:   UpdateMap{},
		initial:  activity,
		Activity: activity,
	}
}

func (ab *ActivityBuilder) Initial() Activity {
	return ab.initial
}

func (ab *ActivityBuilder) SetUserID(id primitive.ObjectID) *ActivityBuilder {
	ab.Activity.State.UserID = id
	ab.Update.Set("state.user_id", id)

	return ab
}

func (ab *ActivityBuilder) SetType(t ActivityType) *ActivityBuilder {
	ab.Activity.Type = t
	ab.Update.Set("type", t)

	return ab
}

func (ab *ActivityBuilder) SetName(name ActivityName) *ActivityBuilder {
	ab.Activity.Name = name
	ab.Update.Set("name", name)

	return ab
}

func (ab *ActivityBuilder) SetStatus(status ActivityStatus) *ActivityBuilder {
	ab.Activity.Status = status
	ab.Update.Set("status", status)

	return ab
}

func (ab *ActivityBuilder) SetTimespan(start, end time.Time) *ActivityBuilder {
	ab.Activity.State.Timespan.Start = start
	ab.Activity.State.Timespan.End = utils.Ternary(end.IsZero(), nil, &end)
	ab.Update.Set("state.timespan", ab.Activity.State.Timespan)

	return ab
}

func (ab *ActivityBuilder) SetObject(kind ObjectKind, id primitive.ObjectID) *ActivityBuilder {
	ab.Activity.Object = &ActivityObject{
		Kind: kind,
		ID:   id,
	}
	ab.Update.Set("object", ab.Activity.Object)

	return ab
}
