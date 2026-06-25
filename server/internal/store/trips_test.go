package store

import (
	"testing"
)

// newTestStore 用 SQLite 記憶體 DB 建一個乾淨的 store(毫秒級,免外部依賴)。
func newTestStore(t *testing.T) *Store {
	t.Helper()
	s, err := Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("open test store: %v", err)
	}
	return s
}

// TestFindOrCreateTrip 驗證「以區間事件為骨架」的歸組邏輯。
func TestFindOrCreateTrip(t *testing.T) {
	s := newTestStore(t)
	ch := "ch_test"

	// 1. 區間事件:住宿 2026-06-29 ~ 2026-07-01,框出行程範圍 → 新建 trip。
	hotelTrip, err := s.FindOrCreateTrip(ch, "2026-06-29", "2026-07-01", "hotel")
	if err != nil {
		t.Fatalf("hotel: %v", err)
	}
	if hotelTrip == nil {
		t.Fatal("住宿(有時間)應建立 trip,卻回 nil")
	}

	// 2. 單點事件:2026-06-30 開會,落在住宿範圍內 → 應歸入同一 trip。
	meetingTrip, err := s.FindOrCreateTrip(ch, "2026-06-30 14:00", "", "meeting")
	if err != nil {
		t.Fatalf("meeting: %v", err)
	}
	if meetingTrip == nil || *meetingTrip != *hotelTrip {
		t.Fatalf("範圍內的開會應歸入住宿 trip(%v),卻得 %v", deref(hotelTrip), deref(meetingTrip))
	}

	// 3. 範圍外事件:2026-08-15 看牙醫 → 應自成新 trip。
	dentistTrip, err := s.FindOrCreateTrip(ch, "2026-08-15", "", "dentist")
	if err != nil {
		t.Fatalf("dentist: %v", err)
	}
	if dentistTrip == nil || *dentistTrip == *hotelTrip {
		t.Fatalf("範圍外的看牙醫應自成新 trip,卻得 %v(住宿是 %v)", deref(dentistTrip), deref(hotelTrip))
	}

	// 4. 無時間事件 → 不歸組(回 nil)。
	noTimeTrip, err := s.FindOrCreateTrip(ch, "", "", "someday todo")
	if err != nil {
		t.Fatalf("noTime: %v", err)
	}
	if noTimeTrip != nil {
		t.Fatalf("無時間事件不應歸組,卻得 %v", deref(noTimeTrip))
	}

	// 5. 應只有 2 個 trip(住宿行程、看牙醫)。
	trips, err := s.ListTripsByChannel(ch)
	if err != nil {
		t.Fatalf("list trips: %v", err)
	}
	if len(trips) != 2 {
		t.Fatalf("應有 2 個 trip,卻有 %d 個", len(trips))
	}
	t.Logf("trips: %+v", trips)
}

// TestFindOrCreateTrip_ExpandsRange 驗證歸入時 trip 範圍會擴張。
func TestFindOrCreateTrip_ExpandsRange(t *testing.T) {
	s := newTestStore(t)
	ch := "ch_expand"

	// 先建一個 6/29~6/30 的 trip。
	trip, _ := s.FindOrCreateTrip(ch, "2026-06-29", "2026-06-30", "trip")
	// 加一個 7/01 的事件,與 6/30 相鄰(7/01 <= 6/30? 否)— 改用落在範圍邊界內驗證擴張。
	// 6/30 的事件落在範圍內,但加一個 6/28 的事件應把起點往前擴張到 6/28。
	early, _ := s.FindOrCreateTrip(ch, "2026-06-28", "2026-06-29 12:00", "earlier")
	if early == nil || *early != *trip {
		t.Fatalf("6/28~6/29 與 6/29~6/30 重疊,應歸同 trip")
	}

	trips, _ := s.ListTripsByChannel(ch)
	if len(trips) != 1 {
		t.Fatalf("應仍是 1 個 trip(擴張而非新建),卻有 %d 個", len(trips))
	}
	if trips[0].Start != "2026-06-28" {
		t.Fatalf("trip 起點應擴張到 2026-06-28,卻是 %s", trips[0].Start)
	}
	t.Logf("擴張後 trip: %s ~ %s", trips[0].Start, trips[0].End)
}

func deref(p *string) string {
	if p == nil {
		return "nil"
	}
	return *p
}
