package service

import (
	"errors"
	"fmt"
	"sort"

	"github.com/wayt-app/wayt-core/repository"
)

// SlotTable represents availability of one table group within a time slot.
type SlotTable struct {
	TableTypeID  uint   `json:"table_type_id"` // representative ID (any physical table in the group)
	Name         string `json:"name"`
	Capacity     int    `json:"capacity"`
	TotalTables  int    `json:"total_tables"` // computed: count of physical tables in group
	Available    int64  `json:"available"`
	TablesNeeded int    `json:"tables_needed"`
}

// SlotResult is a single bookable time slot with available table groups.
type SlotResult struct {
	StartTime    string      `json:"start_time"`
	EndTime      string      `json:"end_time"`
	Tables       []SlotTable `json:"tables"`
	AutoAssigned *SlotTable  `json:"auto_assigned"`
}

type SlotService interface {
	GetSlots(branchID uint, dateStr string, guests int, roomID *uint) ([]SlotResult, error)
}

type slotService struct {
	branchRepo    repository.BranchRepository
	tableTypeRepo repository.TableTypeRepository
	bookingRepo   repository.BookingRepository
}

func NewSlotService(
	branchRepo repository.BranchRepository,
	tableTypeRepo repository.TableTypeRepository,
	bookingRepo repository.BookingRepository,
) SlotService {
	return &slotService{
		branchRepo:    branchRepo,
		tableTypeRepo: tableTypeRepo,
		bookingRepo:   bookingRepo,
	}
}

// tableGroup groups physical table rows sharing the same name+capacity+room.
type tableGroup struct {
	representative uint   // ID of the first physical table in the group (used as slot table_type_id)
	name           string
	capacity       int
	roomID         *uint
	ids            []uint // all physical table IDs in the group
}

func tableGroupKey(name string, capacity int, roomID *uint) string {
	if roomID == nil {
		return fmt.Sprintf("%s|%d|", name, capacity)
	}
	return fmt.Sprintf("%s|%d|%d", name, capacity, *roomID)
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

	// Build groups: each group = all physical tables with same name+capacity+room.
	groups := map[string]*tableGroup{}
	var groupOrder []string // preserve insertion order for stable output
	idToGroupKey := map[uint]string{}
	for _, tt := range tableTypes {
		if !tt.IsActive {
			continue
		}
		key := tableGroupKey(tt.Name, tt.Capacity, tt.RoomID)
		if _, ok := groups[key]; !ok {
			groups[key] = &tableGroup{
				representative: tt.ID,
				name:           tt.Name,
				capacity:       tt.Capacity,
				roomID:         tt.RoomID,
			}
			groupOrder = append(groupOrder, key)
		}
		groups[key].ids = append(groups[key].ids, tt.ID)
		idToGroupKey[tt.ID] = key
	}

	// Fetch all active bookings for this branch+date in ONE query.
	activeBookings, err := s.bookingRepo.FindActiveByBranchDate(branchID, date)
	if err != nil {
		return nil, err
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

		// Compute booked tables_count per group from in-memory bookings.
		bookedByGroup := make(map[string]int64)
		for _, b := range activeBookings {
			bStart := b.StartTime
			if len(bStart) > 5 {
				bStart = bStart[:5]
			}
			bEnd := b.EndTime
			if len(bEnd) > 5 {
				bEnd = bEnd[:5]
			}
			if bStart < end && bEnd > start {
				if gk, ok := idToGroupKey[b.TableTypeID]; ok {
					bookedByGroup[gk] += int64(b.TablesCount)
				}
			}
		}

		var slotTables []SlotTable
		var availableTables []SlotTable
		for _, key := range groupOrder {
			g := groups[key]
			totalTables := int64(len(g.ids))
			booked := bookedByGroup[key]
			available := totalTables - booked
			if available < 0 {
				available = 0
			}
			tablesNeeded := 1
			if guests > 0 {
				tablesNeeded = (guests + g.capacity - 1) / g.capacity
			}
			st := SlotTable{
				TableTypeID:  g.representative,
				Name:         g.name,
				Capacity:     g.capacity,
				TotalTables:  int(totalTables),
				Available:    available,
				TablesNeeded: tablesNeeded,
			}
			slotTables = append(slotTables, st)
			if tablesNeeded <= int(totalTables) && available >= int64(tablesNeeded) {
				availableTables = append(availableTables, st)
			}
		}

		// Sort slotTables by capacity ascending for display
		sort.Slice(slotTables, func(i, j int) bool {
			return slotTables[i].Capacity < slotTables[j].Capacity
		})

		// Sort availableTables: min wasted seats first, then min tables needed
		wastedSeats := func(st SlotTable) int {
			if guests <= 0 {
				return 0
			}
			return st.TablesNeeded*st.Capacity - guests
		}
		sort.Slice(availableTables, func(i, j int) bool {
			wi, wj := wastedSeats(availableTables[i]), wastedSeats(availableTables[j])
			if wi != wj {
				return wi < wj
			}
			return availableTables[i].TablesNeeded < availableTables[j].TablesNeeded
		})

		slot := SlotResult{StartTime: start, EndTime: end, Tables: slotTables}
		if len(availableTables) > 0 {
			auto := availableTables[0]
			slot.AutoAssigned = &auto
		}
		results = append(results, slot)
	}

	return results, nil
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
