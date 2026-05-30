package service

import "errors"

func (s *bookingService) CheckAvailability(branchID uint, dateStr, startTime string, guests int, roomID *uint) ([]AvailabilityResult, error) {
	branch, err := s.branchRepo.FindByID(branchID)
	if err != nil {
		return nil, errors.New("cabang tidak ditemukan")
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
	if err := validateTime(startTime); err != nil {
		return nil, err
	}

	endTime := addMinutes(startTime, branch.DefaultDurationMinutes)
	tableTypes, err := s.tableTypeRepo.FindByBranchAndRoom(branchID, roomID)
	if err != nil {
		return nil, err
	}

	// Group physical tables by (name, capacity, room_id)
	type grp struct {
		rep      uint
		name     string
		capacity int
		roomID   *uint
		ids      []uint
	}
	groups := map[string]*grp{}
	var groupOrder []string
	for _, tt := range tableTypes {
		if !tt.IsActive {
			continue
		}
		key := tableGroupKey(tt.Capacity, tt.RoomID)
		if _, ok := groups[key]; !ok {
			groups[key] = &grp{rep: tt.ID, name: tt.Name, capacity: tt.Capacity, roomID: tt.RoomID}
			groupOrder = append(groupOrder, key)
		}
		groups[key].ids = append(groups[key].ids, tt.ID)
	}

	var results []AvailabilityResult
	for _, key := range groupOrder {
		g := groups[key]
		totalTables := len(g.ids)
		tablesNeeded := 1
		if guests > 0 {
			tablesNeeded = (guests + g.capacity - 1) / g.capacity
		}
		booked, err := s.repo.CountOverlappingByGroup(g.ids, date, startTime, endTime, 0)
		if err != nil {
			booked = 0
		}
		available := int64(totalTables) - booked
		results = append(results, AvailabilityResult{
			TableTypeID:  g.rep,
			Name:         g.name,
			Capacity:     g.capacity,
			TotalTables:  totalTables,
			BookedCount:  booked,
			Available:    available,
			TablesNeeded: tablesNeeded,
			StartTime:    startTime,
			EndTime:      endTime,
		})
	}
	return results, nil
}

func (s *bookingService) GetTableStatus(branchID uint, dateStr, startTime string) (*TableStatusResult, error) {
	branch, err := s.branchRepo.FindByID(branchID)
	if err != nil {
		return nil, errors.New("cabang tidak ditemukan")
	}
	if dateStr == "" {
		dateStr = today().Format("2006-01-02")
	}
	date, err := parseDate(dateStr)
	if err != nil {
		return nil, errors.New("format tanggal tidak valid")
	}
	if err := validateTime(startTime); err != nil {
		return nil, err
	}
	endTime := addMinutes(startTime, branch.DefaultDurationMinutes)

	bookedMap, err := s.repo.CountOverlappingByTableType(branchID, date, startTime, endTime)
	if err != nil {
		return nil, err
	}

	tableTypes, err := s.tableTypeRepo.FindByBranch(branchID)
	if err != nil {
		return nil, err
	}

	// Group physical tables; sum booked per group
	type tsGrp struct {
		rep      uint
		name     string
		capacity int
		ids      []uint
	}
	tsGroups := map[string]*tsGrp{}
	var tsOrder []string
	for _, tt := range tableTypes {
		if !tt.IsActive {
			continue
		}
		key := tableGroupKey(tt.Capacity, tt.RoomID)
		if _, ok := tsGroups[key]; !ok {
			tsGroups[key] = &tsGrp{rep: tt.ID, name: tt.Name, capacity: tt.Capacity}
			tsOrder = append(tsOrder, key)
		}
		tsGroups[key].ids = append(tsGroups[key].ids, tt.ID)
	}
	var tables []TableTypeStatus
	for _, key := range tsOrder {
		g := tsGroups[key]
		var booked int64
		for _, id := range g.ids {
			booked += bookedMap[id]
		}
		totalTables := int64(len(g.ids))
		available := totalTables - booked
		if available < 0 {
			available = 0
		}
		tables = append(tables, TableTypeStatus{
			TableTypeID: g.rep,
			Name:        g.name,
			Capacity:    g.capacity,
			TotalTables: int(totalTables),
			Booked:      booked,
			Available:   available,
		})
	}

	return &TableStatusResult{
		Date:      dateStr,
		StartTime: startTime,
		EndTime:   endTime,
		Tables:    tables,
	}, nil
}
