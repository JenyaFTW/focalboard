package utils

import (
	"encoding/json"
	"time"

	"github.com/mattermost/focalboard/server/model"

	mm_model "github.com/mattermost/mattermost-server/v6/model"
)

type IDType byte

const (
	IDTypeNone      IDType = '7'
	IDTypeWorkspace IDType = 'w'
	IDTypeBoard     IDType = 'b'
	IDTypeCard      IDType = 'c'
	IDTypeView      IDType = 'v'
	IDTypeSession   IDType = 's'
	IDTypeUser      IDType = 'u'
	IDTypeToken     IDType = 'k'
	IDTypeBlock     IDType = 'a'
)

// NewId is a globally unique identifier.  It is a [A-Z0-9] string 27
// characters long.  It is a UUID version 4 Guid that is zbased32 encoded
// with the padding stripped off, and a one character alpha prefix indicating the
// type of entity or a `7` if unknown type.
func NewID(idType IDType) string {
	return string(idType) + mm_model.NewId()
}

// BlockType2IDType returns an appropriate IDType for the specified BlockType.
func BlockType2IDType(blockType model.BlockType) IDType {
	switch blockType {
	case model.TypeBoard:
		return IDTypeBoard
	case model.TypeCard:
		return IDTypeCard
	case model.TypeView:
		return IDTypeView
	case model.TypeText, model.TypeComment:
		return IDTypeBlock
	}
	return IDTypeNone
}

// GetMillis is a convenience method to get milliseconds since epoch.
func GetMillis() int64 {
	return mm_model.GetMillis()
}

// GetMillisForTime is a convenience method to get milliseconds since epoch for provided Time.
func GetMillisForTime(thisTime time.Time) int64 {
	return mm_model.GetMillisForTime(thisTime)
}

// GetTimeForMillis is a convenience method to get time.Time for milliseconds since epoch.
func GetTimeForMillis(millis int64) time.Time {
	return mm_model.GetTimeForMillis(millis)
}

// SecondsToMillis is a convenience method to convert seconds to milliseconds.
func SecondsToMillis(seconds int64) int64 {
	return seconds * 1000
}

func StructToMap(v interface{}) (m map[string]interface{}) {
	b, _ := json.Marshal(v)
	_ = json.Unmarshal(b, &m)
	return
}
