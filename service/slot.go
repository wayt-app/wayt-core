package service

import (
	"errors"
	"fmt"
	"math"
	"sort"

	"github.com/wayt-app/wayt-core/model"
	"github.com/wayt-app/wayt-core/repository"
)

// AssignedTable is one line in an optimal room assignment.
type AssignedTable struct {
	TableTypeID uint   `json:"table_type_id"`
	Name        string `json:"name"`
	Capacity    int    `json:"capacity"`
	Count       int    `json:"count"`
}

// SlotRoom represents a room's availability for a given time slot.
type SlotRoom struct {
	RoomID      *uint           `json:"room_id"`
	RoomName    string          `json:"room_name"`
	Assignment  []AssignedTable `json:"assignment"`  // optimal table assignment
	TotalSeats  int             `json:"total_seats"` // seats covered by assignment
	WastedSeats int             `json:"wasted_seats"`
	TableCount  int             `json:"table_count"`
}

// SlotResult is a single bookable time slot.
type SlotResult struct {
	StartTime    string     `json:"start_time"`
	EndTime      string     `json:"end_time"`
	Rooms        []SlotRoom `json:"rooms"`         // all rooms with an available assignment
	AutoAssigned *SlotRoom  `json:"auto_assigned"` // best room (min wasted, then min tables)
}

// SlotTable is kept for backward compat with CheckAvailability.
type SlotTable struct {
	TableTypeID  uint   `json:"table_type_id"`
	Name         string `json:"name"`
	Capacity     int    `json:"capacity"`
	TotalTables  int    `json:"total_tables"`
	Available    int64  `json:"available"`
	TablesNeeded int    `json:"tables_needed"`
}

type SlotService interface {
	GetSlots(branchID uint, dateStr string, guests int, roomID *uint) ([]SlotResult, error)
}

type slotService struct {
	branchRepo       repository.BranchRepository
	tableTypeRepo    repository.TableTypeRepository
	bookingRepo      repository.BookingRepository
	bookingTableRepo repository.BookingTableRepository
}

func NewSlotService(
	branchRepo repository.BranchRepository,
	tableTypeRepo repository.TableTypeRepository,
	bookingRepo repository.BookingRepository,
	bookingTableRepo repository.BookingTableRepository,
) SlotService {
	return &slotService{
		branchRepo:       branchRepo,
		tableTypeRepo:    tableTypeRepo,
		bookingRepo:      bookingRepo,
		bookingTableRepo: bookingTableRepo,
	}
}

// roomPool groups physical tables belonging to the same room.
type roomPool struct {
	roomID   *uint
	roomName string
	tables   []model.TableType // all active physical tables in this room
}

func roomKey(roomID *uint) string {
	if roomID == nil {
		return ""
	}
	return fmt.Sprintf("%d", *roomID)
}

// tableGroupKey is used by booking.go for grouping by capacity+room.
func tableGroupKey(capacity int, roomID *uint) string {
	if roomID == nil {
		return fmt.Sprintf("%d|", capacity)
	}
	return fmt.Sprintf("%d|%d", capacity, *roomID)
}

func (s *slotService) GetSlots(branchID uint, dateStr string, guests int, roomID *uint) ([]SlotResult, error) {
	branch, err := s.branchRepo.FindByID(branchID)
	if err != nil {
		return nil, errors.New("cabang tidak ditemukan")
	}
	if !branch.IsActive {
		return nil, errors.New("cabang tidak aktif")
	}
	if branch.OpenFrom == "" || branch.OpenTo == "" {
		return nil, errors.New("jam operasional belum diatur untuk cabang ini")
	}

	date, err := parseDate(dateStr)
	if err != nil {
		return nil, errors.New("format tanggal tidak valid, gunakan YYYY-MM-DD")
	}
	if date.Before(today()) {
		return nil, errors.New("tanggal tidak boleh di masa lalu")
	}
	if date.After(today().AddDate(0, 0, 30)) {
		return nil, errors.New("booking maksimal 30 hari ke depan")
	}

	interval := branch.SlotIntervalMinutes
	if interval <= 0 {
		interval = 30
	}
	duration := branch.DefaultDurationMinutes
	if duration <= 0 {
		duration = 120
	}

	startTimes := generateSlots(branch.OpenFrom, branch.OpenTo, interval, duration)
	if len(startTimes) == 0 {
		return nil, errors.New("tidak ada slot tersedia untuk jam operasional yang diatur")
	}

	tableTypes, err := s.tableTypeRepo.FindByBranchAndRoom(branchID, roomID)
	if err != nil {
		return nil, err
	}

	// Group tables by room.
	roomMap := map[string]*roomPool{}
	var roomOrder []string
	for _, tt := range tableTypes {
		if !tt.IsActive {
			continue
		}
		key := roomKey(tt.RoomID)
		if _, ok := roomMap[key]; !ok {
			roomName := ""
			if tt.Room != nil {
				roomName = tt.Room.Name
			}
			roomMap[key] = &roomPool{roomID: tt.RoomID, roomName: roomName}
			roomOrder = append(roomOrder, key)
		}
		roomMap[key].tables = append(roomMap[key].tables, tt)
	}

	// Collect all table type IDs for availability query.
	var allTypeIDs []uint
	for _, rp := range roomMap {
		for _, tt := range rp.tables {
			allTypeIDs = append(allTypeIDs, tt.ID)
		}
	}

	nowStr := ""
	isToday := date.Equal(today())
	if isToday {
		nowStr = nowWIB().Format("15:04")
	}

	var results []SlotResult
	for _, start := range startTimes {
		end := addMinutes(start, duration)
		if isToday && start <= nowStr {
			continue
		}

		// Fetch booked count per table_type_id from booking_tables for this slot.
		bookedByTypeID, err := s.bookingTableRepo.SumByTypeIDsAndSlot(allTypeIDs, date, start, end, 0)
		if err != nil {
			bookedByTypeID = map[uint]int64{}
		}

		var availRooms []SlotRoom
		for _, key := range roomOrder {
			rp := roomMap[key]

			// Build available-count per table type for this room.
			var avails []typeAvail
			for _, tt := range rp.tables {
				free := 1 - int(bookedByTypeID[tt.ID])
				if free > 0 {
					avails = append(avails, typeAvail{
						id:       tt.ID,
						name:     tt.Name,
						capacity: tt.Capacity,
						count:    free,
					})
				}
			}

			if guests <= 0 {
				// No guest count: just show room as available
				availRooms = append(availRooms, SlotRoom{
					RoomID:   rp.roomID,
					RoomName: rp.roomName,
				})
				continue
			}

			// Compute optimal assignment for this room.
			assign := computeOptimalAssignment(guests, avails)
			if assign == nil {
				continue // room can't accommodate
			}

			wasted := 0
			tableCount := 0
			var assigned []AssignedTable
			for _, a := range assign {
				wasted += a.count*a.capacity - guests
				if wasted < 0 { // reset: wasted is cumulative only at end
					wasted = 0
				}
				tableCount += a.count
				assigned = append(assigned, AssignedTable{
					TableTypeID: a.id,
					Name:        a.name,
					Capacity:    a.capacity,
					Count:       a.count,
				})
			}
			// Recalculate wasted correctly
			total := 0
			for _, a := range assign {
				total += a.count * a.capacity
			}
			wasted = total - guests

			availRooms = append(availRooms, SlotRoom{
				RoomID:      rp.roomID,
				RoomName:    rp.roomName,
				Assignment:  assigned,
				TotalSeats:  total,
				WastedSeats: wasted,
				TableCount:  tableCount,
			})
		}

		// Sort availRooms: min wasted, then min tables
		sort.Slice(availRooms, func(i, j int) bool {
			if availRooms[i].WastedSeats != availRooms[j].WastedSeats {
				return availRooms[i].WastedSeats < availRooms[j].WastedSeats
			}
			return availRooms[i].TableCount < availRooms[j].TableCount
		})

		slot := SlotResult{StartTime: start, EndTime: end, Rooms: availRooms}
		if len(availRooms) > 0 {
			auto := availRooms[0]
			slot.AutoAssigned = &auto
		}
		results = append(results, slot)
	}

	return results, nil
}

// typeAvail is used internally for optimal assignment.
type typeAvail struct {
	id       uint
	name     string
	capacity int
	count    int // available units
}

type typeAssign struct {
	id       uint
	name     string
	capacity int
	count    int
}

// computeOptimalAssignment finds the combination of tables with minimum wasted seats,
// then minimum table count, that seats at least `guests` people.
// Returns nil if impossible.
func computeOptimalAssignment(guests int, avails []typeAvail) []typeAssign {
	if guests <= 0 || len(avails) == 0 {
		return nil
	}

	// Sort descending by capacity for better pruning.
	sort.Slice(avails, func(i, j int) bool { return avails[i].capacity > avails[j].capacity })

	bestWasted := math.MaxInt32
	bestTables := math.MaxInt32
	var best []typeAssign

	var dfs func(idx, remaining int, cur []typeAssign)
	dfs = func(idx, remaining int, cur []typeAssign) {
		if remaining <= 0 {
			wasted := -remaining
			tables := 0
			for _, a := range cur {
				tables += a.count
			}
			if wasted < bestWasted || (wasted == bestWasted && tables < bestTables) {
				bestWasted = wasted
				bestTables = tables
				best = append([]typeAssign{}, cur...)
			}
			return
		}
		if idx >= len(avails) {
			return
		}
		t := avails[idx]
		maxUse := t.count
		ceil := (remaining + t.capacity - 1) / t.capacity
		if ceil < maxUse {
			maxUse = ceil
		}
		// Try from maxUse down to 0 (larger use first for early termination).
		for use := maxUse; use >= 0; use-- {
			var newCur []typeAssign
			if use > 0 {
				newCur = append(append([]typeAssign{}, cur...), typeAssign{
					id:       t.id,
					name:     t.name,
					capacity: t.capacity,
					count:    use,
				})
			} else {
				newCur = cur
			}
			dfs(idx+1, remaining-use*t.capacity, newCur)
		}
	}
	dfs(0, guests, nil)

	if best == nil {
		return nil
	}
	return best
}

// generateSlots returns "HH:MM" start times from openFrom up to
// (openTo - durationMinutes), stepping by intervalMinutes.
func generateSlots(openFrom, openTo string, intervalMinutes, durationMinutes int) []string {
	fromMins := timeToMinutes(openFrom)
	toMins := timeToMinutes(openTo)
	if fromMins < 0 || toMins <= fromMins {
		return nil
	}

	lastStart := toMins - durationMinutes
	if lastStart < fromMins {
		return nil
	}

	var slots []string
	for cur := fromMins; cur <= lastStart; cur += intervalMinutes {
		slots = append(slots, minutesToTime(cur))
	}
	return slots
}

// timeToMinutes converts "HH:MM" to minutes since midnight. Returns -1 on error.
func timeToMinutes(t string) int {
	if len(t) != 5 || t[2] != ':' {
		return -1
	}
	h := int(t[0]-'0')*10 + int(t[1]-'0')
	m := int(t[3]-'0')*10 + int(t[4]-'0')
	if h < 0 || h > 23 || m < 0 || m > 59 {
		return -1
	}
	return h*60 + m
}

// minutesToTime converts minutes since midnight to "HH:MM".
func minutesToTime(mins int) string {
	return fmt.Sprintf("%02d:%02d", mins/60, mins%60)
}
