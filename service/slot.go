package service

import (
	"errors"
	"fmt"
	"sort"

	"github.com/wayt-app/wayt-core/repository"
)

// SlotTable represents availability of one table type within a time slot.
type SlotTable struct {
	TableTypeID  uint   `json:"table_type_id"`
	Name         string `json:"name"`
	Capacity     int    `json:"capacity"`
	TotalTables  int    `json:"total_tables"`
	Available    int64  `json:"available"`
	TablesNeeded int    `json:"tables_needed"`
}

// SlotResult is a single bookable time slot with available table types.
// Tables are sorted by capacity ascending (smallest-that-fits first) to support auto-assignment.
type SlotResult struct {
	StartTime       string      `json:"start_time"`        // "HH:MM"
	EndTime         string      `json:"end_time"`          // "HH:MM"
	Tables          []SlotTable `json:"tables"`            // sorted by capacity ASC, available > 0
	AutoAssigned    *SlotTable  `json:"auto_assigned"`     // smallest table that fits guest count
}

type SlotService interface {
	GetSlots(branchID uint, dateStr string, guests int) ([]SlotResult, error)
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

func (s *slotService) GetSlots(branchID uint, dateStr string, guests int) ([]SlotResult, error) {
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

	tableTypes, err := s.tableTypeRepo.FindByBranch(branchID)
	if err != nil {
		return nil, err
	}

	// Fetch all active bookings for this branch+date in ONE query,
	// then compute overlaps per slot in memory — avoids N×M DB round trips.
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

		// Skip slots already past (only for today), using WIB timezone
		if isToday && start <= nowStr {
			continue
		}

		// Compute booked tables per table_type_id for this slot from in-memory bookings
		bookedByType := make(map[uint]int64)
		for _, b := range activeBookings {
			if b.StartTime < end && b.EndTime > start {
				bookedByType[b.TableTypeID] += int64(b.TablesCount)
			}
		}

		var slotTables []SlotTable
		var availableTables []SlotTable
		for _, tt := range tableTypes {
			if !tt.IsActive {
				continue
			}
			// Calculate tables needed via combining; skip if even all tables combined can't fit
			tablesNeeded := 1
			if guests > 0 {
				tablesNeeded = (guests + tt.Capacity - 1) / tt.Capacity
				if tablesNeeded > tt.TotalTables {
					continue
				}
			}
			booked := bookedByType[tt.ID]
			available := int64(tt.TotalTables) - booked
			if available < 0 {
				available = 0
			}
			st := SlotTable{
				TableTypeID:  tt.ID,
				Name:         tt.Name,
				Capacity:     tt.Capacity,
				TotalTables:  tt.TotalTables,
				Available:    available,
				TablesNeeded: tablesNeeded,
			}
			slotTables = append(slotTables, st)
			// Available for auto-assign only if enough tables are free for this group
			if available >= int64(tablesNeeded) {
				availableTables = append(availableTables, st)
			}
		}

		// Sort slotTables by capacity ascending for display
		sort.Slice(slotTables, func(i, j int) bool {
			return slotTables[i].Capacity < slotTables[j].Capacity
		})

		// Sort availableTables for auto-assign:
		// Primary: minimum wasted seats (tablesNeeded×capacity - guests)
		// Secondary: minimum tables needed
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
			// Auto-assign: pick best fit (min wasted seats, then min tables)
			auto := availableTables[0]
			slot.AutoAssigned = &auto
		}
		// Always include slot (even if full) so customer can join waiting list
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
