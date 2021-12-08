package dao

import "testing"

func TestMessagesStruct(t *testing.T) {
	message := &Messages{
		ID:          1,
		SeqId:       2,
		TSeqId:      3,
		FSeqId:      4,
		To:          5,
		From:        6,
		Content:     "test",
		Group:       7,
		ContentType: "html",
		Topic:       8,
		Status:      9,
		Origin:      10,
		CreatedAt:   11,
	}

	testMap := map[string]interface{}{
		"ID":          "1",
		"SeqId":       "2",
		"TSeqId":      "3",
		"FSeqId":      "4",
		"To":          "5",
		"From":        "6",
		"Content":     "test",
		"Group":       "7",
		"ContentType": "html",
		"Topic":       "8",
		"Status":      "9",
		"Origin":      "10",
		"CreatedAt":   "11",
	}

	message2 := &Messages{}
	if err := message2.FromMap(testMap); err != nil {
		t.Fatal(err)
	}

	message3 := &Messages{}
	if err := message3.FromMap(message.ToMap()); err != nil {
		t.Fatal(err)
	}
}
