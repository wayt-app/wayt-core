package service

import (
	"log"
	"sort"

	"github.com/wayt-app/wayt-core/model"
)

// ProcessNoShows marks confirmed bookings as no_show if they started more than 15 minutes ago,
// then triggers auto-promote for each freed slot.
func (s *bookingService) ProcessNoShows() error {
	candidates, err := s.repo.FindNoShowCandidates(15)
	if err != nil {
		return err
	}
	for _, b := range candidates {
		if err := s.repo.UpdateStatus(b.ID, model.BookingStatusNoShow); err != nil {
			continue
		}
		_ = s.autoPromote(&b)
	}
	return nil
}

// autoPromote promotes the earliest waiting list entry for the same slot when a booking is cancelled.
func (s *bookingService) autoPromote(cancelled *model.Booking) error {
	branch, err := s.branchRepo.FindByID(cancelled.BranchID)
	if err != nil {
		return nil
	}
	waiters, err := s.repo.FindWaitingListForSlot(cancelled.BranchID, cancelled.BookingDate, cancelled.StartTime, cancelled.EndTime)
	if err != nil || len(waiters) == 0 {
		return nil
	}
	tableTypes, err := s.tableTypeRepo.FindByBranch(cancelled.BranchID)
	if err != nil {
		return nil
	}
	sort.Slice(tableTypes, func(i, j int) bool {
		return tableTypes[i].Capacity < tableTypes[j].Capacity
	})

	for _, waiter := range waiters {
		tablesNeeded := waiter.TablesCount
		if tablesNeeded <= 0 {
			// Fallback for old bookings without tables_count
			tablesNeeded = 1
		}
		// Find a table type group that can fit this waiter
		for _, tt := range tableTypes {
			if !tt.IsActive {
				continue
			}
			needed := (waiter.GuestCount + tt.Capacity - 1) / tt.Capacity
			grpT, _ := s.tableTypeRepo.FindByGroup(tt.BranchID, tt.Name, tt.Capacity, tt.RoomID)
			if len(grpT) == 0 || needed > len(grpT) {
				continue
			}
			grpI := make([]uint, len(grpT))
			for i, t := range grpT { grpI[i] = t.ID }
			booked, _ := s.repo.CountOverlappingByGroup(grpI, cancelled.BookingDate, cancelled.StartTime, cancelled.EndTime, waiter.ID)
			if booked+int64(needed) <= int64(len(grpT)) {
				// Assign this table type and promote
				if err := s.repo.UpdateTableType(waiter.ID, tt.ID); err != nil {
					log.Printf("[WARN] autoPromote booking #%d: gagal update table type: %v", waiter.ID, err)
					continue
				}
				status := model.BookingStatusConfirmed
				if branch.RequireConfirmation {
					status = model.BookingStatusPending
				}
				if err := s.repo.UpdateStatus(waiter.ID, status); err != nil {
					log.Printf("[WARN] autoPromote booking #%d: gagal update status: %v", waiter.ID, err)
					continue
				}
				waiter.TableTypeID = tt.ID
				waiter.EndTime = cancelled.EndTime
				go s.sendBookingNotif(&waiter, "promoted")
				go s.sendBookingEmail(&waiter, "promoted")
				go s.sendInAppNotif(&waiter, "promoted")
				return nil // promote one at a time
			}
		}
	}
	return nil
}

// ProcessReminders sends H-1 reminders for tomorrow's bookings (pending or confirmed).
// Designed to be called by a background job (e.g., every hour).
func (s *bookingService) ProcessReminders() error {
	candidates, err := s.repo.FindReminderCandidates()
	if err != nil {
		return err
	}
	for _, b := range candidates {
		b := b // capture loop variable
		s.sendReminderNotif(&b)
		if err := s.repo.MarkReminderSent(b.ID); err != nil {
			log.Printf("[WARN] ProcessReminders: gagal mark reminder sent booking #%d: %v", b.ID, err)
		}
	}
	return nil
}
