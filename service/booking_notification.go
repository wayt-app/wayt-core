package service

import (
	"fmt"
	"log"
	"time"

	"github.com/wayt-app/wayt-core/model"
)

// sendBookingNotif sends a WhatsApp notification to the customer for a booking event.
// Errors are ignored — notifications are best-effort.
func (s *bookingService) sendBookingNotif(b *model.Booking, event string) {
	customer, err := s.customerRepo.FindByID(b.CustomerID)
	if err != nil || customer == nil || customer.Phone == "" {
		return
	}
	branch, err := s.branchRepo.FindByID(b.BranchID)
	if err != nil {
		return
	}

	dateStr := b.BookingDate.Format("02 Jan 2006")
	var msg string
	switch event {
	case "confirmed":
		msg = fmt.Sprintf(
			"Halo *%s*! 🎉\n\nBooking Anda di *%s* telah *dikonfirmasi*.\n\n📅 %s\n⏰ %s – %s\n👥 %d tamu\n🔖 ID: #%d\n\nSampai jumpa!",
			customer.Name, branch.Name, dateStr, b.StartTime, b.EndTime, b.GuestCount, b.ID,
		)
	case "pending":
		msg = fmt.Sprintf(
			"Halo *%s*! 📋\n\nBooking Anda di *%s* sedang *menunggu konfirmasi* admin.\n\n📅 %s\n⏰ %s – %s\n👥 %d tamu\n🔖 ID: #%d\n\nKami akan segera mengonfirmasi.",
			customer.Name, branch.Name, dateStr, b.StartTime, b.EndTime, b.GuestCount, b.ID,
		)
	case "waiting_list":
		msg = fmt.Sprintf(
			"Halo *%s*! ⏳\n\nMaaf, slot penuh. Anda telah masuk *waiting list* di *%s*.\n\n📅 %s\n⏰ %s – %s\n👥 %d tamu\n🔖 ID: #%d\n\nKami akan memberitahu jika ada slot kosong.",
			customer.Name, branch.Name, dateStr, b.StartTime, b.EndTime, b.GuestCount, b.ID,
		)
	case "promoted":
		msg = fmt.Sprintf(
			"Halo *%s*! 🎊\n\nKabar baik! Slot tersedia dan booking Anda di *%s* telah *dikonfirmasi*.\n\n📅 %s\n⏰ %s – %s\n👥 %d tamu\n🔖 ID: #%d\n\nSampai jumpa!",
			customer.Name, branch.Name, dateStr, b.StartTime, b.EndTime, b.GuestCount, b.ID,
		)
	case "cancelled":
		msg = fmt.Sprintf(
			"Halo *%s*,\n\nBooking Anda #%d di *%s* pada %s pukul %s telah *dibatalkan*.\n\nTerima kasih.",
			customer.Name, b.ID, branch.Name, dateStr, b.StartTime,
		)
	default:
		return
	}

	if err := s.waSender.Send(customer.Phone, msg); err != nil {
		log.Printf("[WA ERROR] booking #%d customer_id=%d: %v", b.ID, b.CustomerID, err)
	}
}

// sendBookingEmail sends an HTML email notification for a booking event.
// Errors are ignored — notifications are best-effort.
func (s *bookingService) sendBookingEmail(b *model.Booking, event string) {
	customer, err := s.customerRepo.FindByID(b.CustomerID)
	if err != nil || customer == nil {
		return
	}
	recipientEmail := customer.Email
	if b.GuestEmail != "" {
		recipientEmail = b.GuestEmail
	}
	if recipientEmail == "" {
		return
	}

	var emailCfg *model.EmailConfig
	if s.emailConfigRepo != nil {
		emailCfg, _ = s.emailConfigRepo.Get()
	}
	branch, err := s.branchRepo.FindByID(b.BranchID)
	if err != nil {
		return
	}

	dateStr := b.BookingDate.Format("02 Jan 2006")

	var subject, body string
	switch event {
	case "confirmed":
		subject = fmt.Sprintf("Booking #%d Dikonfirmasi — %s", b.ID, branch.Name)
		body = fmt.Sprintf(`
<p>Halo <strong>%s</strong>,</p>
<p>Booking Anda telah <strong style="color:#16a34a">dikonfirmasi</strong>. Berikut detailnya:</p>
<table style="border-collapse:collapse;margin:16px 0">
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">Restoran</td><td><strong>%s</strong></td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">Tanggal</td><td>%s</td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">Waktu</td><td>%s – %s</td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">Tamu</td><td>%d orang</td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">ID Booking</td><td>#%d</td></tr>
</table>
<p>Sampai jumpa!</p>`,
			customer.Name, branch.Name, dateStr, b.StartTime, b.EndTime, b.GuestCount, b.ID)

	case "pending":
		subject = fmt.Sprintf("Booking #%d Menunggu Konfirmasi — %s", b.ID, branch.Name)
		body = fmt.Sprintf(`
<p>Halo <strong>%s</strong>,</p>
<p>Booking Anda sedang <strong>menunggu konfirmasi</strong> dari admin. Kami akan segera mengonfirmasi.</p>
<table style="border-collapse:collapse;margin:16px 0">
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">Restoran</td><td><strong>%s</strong></td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">Tanggal</td><td>%s</td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">Waktu</td><td>%s – %s</td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">Tamu</td><td>%d orang</td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">ID Booking</td><td>#%d</td></tr>
</table>`,
			customer.Name, branch.Name, dateStr, b.StartTime, b.EndTime, b.GuestCount, b.ID)

	case "waiting_list":
		subject = fmt.Sprintf("Booking #%d Masuk Waiting List — %s", b.ID, branch.Name)
		body = fmt.Sprintf(`
<p>Halo <strong>%s</strong>,</p>
<p>Maaf, slot yang Anda pilih sedang penuh. Anda telah masuk <strong>waiting list</strong>.</p>
<table style="border-collapse:collapse;margin:16px 0">
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">Restoran</td><td><strong>%s</strong></td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">Tanggal</td><td>%s</td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">Waktu</td><td>%s – %s</td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">Tamu</td><td>%d orang</td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">ID Booking</td><td>#%d</td></tr>
</table>
<p>Kami akan mengirim email jika ada slot kosong untuk Anda.</p>`,
			customer.Name, branch.Name, dateStr, b.StartTime, b.EndTime, b.GuestCount, b.ID)

	case "promoted":
		subject = fmt.Sprintf("Booking #%d Dikonfirmasi dari Waiting List — %s", b.ID, branch.Name)
		body = fmt.Sprintf(`
<p>Halo <strong>%s</strong>,</p>
<p>Kabar baik! Slot tersedia dan booking Anda dari waiting list telah <strong style="color:#16a34a">dikonfirmasi</strong>.</p>
<table style="border-collapse:collapse;margin:16px 0">
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">Restoran</td><td><strong>%s</strong></td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">Tanggal</td><td>%s</td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">Waktu</td><td>%s – %s</td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">Tamu</td><td>%d orang</td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">ID Booking</td><td>#%d</td></tr>
</table>
<p>Sampai jumpa!</p>`,
			customer.Name, branch.Name, dateStr, b.StartTime, b.EndTime, b.GuestCount, b.ID)

	case "checked_in":
		subject = fmt.Sprintf("Check-in Berhasil — %s", branch.Name)
		body = fmt.Sprintf(`
<p>Halo <strong>%s</strong>,</p>
<p>Anda telah berhasil <strong style="color:#7c3aed">check-in</strong> di <strong>%s</strong>. Selamat menikmati!</p>
<table style="border-collapse:collapse;margin:16px 0">
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">Restoran</td><td><strong>%s</strong></td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">Tanggal</td><td>%s</td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">Waktu</td><td>%s – %s</td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">Tamu</td><td>%d orang</td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">ID Booking</td><td>#%d</td></tr>
</table>
<p>Terima kasih telah memilih kami. Semoga pengalaman makan Anda menyenangkan! 🍽️</p>`,
			customer.Name, branch.Name, branch.Name, dateStr, b.StartTime, b.EndTime, b.GuestCount, b.ID)

	case "cancelled":
		subject = fmt.Sprintf("Booking #%d Dibatalkan — %s", b.ID, branch.Name)
		reasonRow := ""
		if b.CancelReason != "" {
			reasonRow = fmt.Sprintf(`<tr><td style="padding:4px 12px 4px 0;color:#6b7280">Alasan</td><td>%s</td></tr>`, b.CancelReason)
		}
		// Determine contact phone: prefer restaurant phone, fall back to branch phone
		contactPhone := branch.Phone
		if restaurant, err := s.restaurantRepo.FindByID(branch.RestaurantID); err == nil && restaurant.Phone != "" {
			contactPhone = restaurant.Phone
		}
		phoneInfo := ""
		if contactPhone != "" {
			phoneInfo = fmt.Sprintf(`<p>Jika ada pertanyaan, silakan hubungi kami di <strong>%s</strong>.</p>`, contactPhone)
		} else {
			phoneInfo = `<p>Jika ada pertanyaan, hubungi kami langsung.</p>`
		}
		body = fmt.Sprintf(`
<p>Halo <strong>%s</strong>,</p>
<p>Booking Anda telah <strong style="color:#dc2626">dibatalkan</strong> oleh restoran.</p>
<table style="border-collapse:collapse;margin:16px 0">
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">Restoran</td><td><strong>%s</strong></td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">Tanggal</td><td>%s</td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">Waktu</td><td>%s</td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">ID Booking</td><td>#%d</td></tr>
  %s
</table>
%s`,
			customer.Name, branch.Name, dateStr, b.StartTime, b.ID, reasonRow, phoneInfo)

	default:
		return
	}

	var restLogoURL string
	if rest, rerr := s.restaurantRepo.FindByID(branch.RestaurantID); rerr == nil {
		restLogoURL = rest.LogoURL
	}
	wrapped := wrapEmailHTML(body, emailCfg, customer.Name, branch.Name, restLogoURL)
	if err := s.emailSender.Send(recipientEmail, subject, wrapped); err != nil {
		log.Printf("[EMAIL ERROR] booking #%d customer_id=%d: %v", b.ID, b.CustomerID, err)
	}
}

// sendOwnerBookingEmail notifies the restaurant owner when a customer creates a new reservation.
// Only fires for confirmed, pending, and waiting_list events. Errors are ignored — best-effort.
func (s *bookingService) sendOwnerBookingEmail(b *model.Booking, event string) {
	if event != string(model.BookingStatusConfirmed) &&
		event != string(model.BookingStatusPending) &&
		event != string(model.BookingStatusWaitingList) {
		return
	}
	if s.ownerRepo == nil {
		return
	}

	branch, err := s.branchRepo.FindByID(b.BranchID)
	if err != nil || branch == nil {
		return
	}
	restaurant, err := s.restaurantRepo.FindByID(branch.RestaurantID)
	if err != nil || restaurant == nil || restaurant.BusinessOwnerID == nil {
		return
	}
	owner, err := s.ownerRepo.FindByID(*restaurant.BusinessOwnerID)
	if err != nil || owner == nil || owner.Email == "" {
		return
	}
	customer, err := s.customerRepo.FindByID(b.CustomerID)
	if err != nil || customer == nil {
		return
	}

	dateStr := b.BookingDate.Format("02 Jan 2006")

	var statusLabel string
	switch event {
	case string(model.BookingStatusConfirmed):
		statusLabel = `<span style="color:#16a34a;font-weight:bold">Terkonfirmasi Otomatis</span>`
	case string(model.BookingStatusPending):
		statusLabel = `<span style="color:#d97706;font-weight:bold">Menunggu Konfirmasi Anda</span>`
	case string(model.BookingStatusWaitingList):
		statusLabel = `<span style="color:#6b7280;font-weight:bold">Waiting List</span>`
	}

	notesRow := ""
	if b.Notes != "" {
		notesRow = fmt.Sprintf(`<tr><td style="padding:4px 12px 4px 0;color:#6b7280">Catatan</td><td>%s</td></tr>`, b.Notes)
	}

	subject := fmt.Sprintf("Reservasi Baru #%d — %s", b.ID, branch.Name)
	body := fmt.Sprintf(`
<p>Halo <strong>%s</strong>,</p>
<p>Ada reservasi baru masuk di cabang <strong>%s</strong>. Status: %s</p>
<table style="border-collapse:collapse;margin:16px 0">
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">Pelanggan</td><td><strong>%s</strong></td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">No. HP</td><td>%s</td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">Tanggal</td><td>%s</td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">Waktu</td><td>%s – %s</td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">Tamu</td><td>%d orang</td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">ID Booking</td><td>#%d</td></tr>
  %s
</table>`,
		owner.Name, branch.Name, statusLabel,
		customer.Name, customer.Phone,
		dateStr, b.StartTime, b.EndTime, b.GuestCount, b.ID,
		notesRow)

	var emailCfg *model.EmailConfig
	if s.emailConfigRepo != nil {
		emailCfg, _ = s.emailConfigRepo.Get()
	}
	var restLogoURL string
	if rest, rerr := s.restaurantRepo.FindByID(branch.RestaurantID); rerr == nil {
		restLogoURL = rest.LogoURL
	}
	wrapped := wrapEmailHTML(body, emailCfg, owner.Name, branch.Name, restLogoURL)
	if err := s.emailSender.Send(owner.Email, subject, wrapped); err != nil {
		log.Printf("[EMAIL ERROR] notif owner booking #%d owner_id=%d: %v", b.ID, owner.ID, err)
	}
}

// sendInAppNotif creates a DB notification and pushes it via SSE.
func (s *bookingService) sendInAppNotif(b *model.Booking, event string) {
	if s.notifSvc == nil {
		return
	}
	branch, _ := s.branchRepo.FindByID(b.BranchID)
	branchName := ""
	if branch != nil {
		branchName = branch.Name
	}
	dateStr := b.BookingDate.Format("02 Jan 2006")

	var customerTitle, customerMsg string
	switch event {
	case string(model.BookingStatusPending):
		customerTitle = "Booking Menunggu Konfirmasi"
		customerMsg = fmt.Sprintf("Booking #%d di %s pada %s pukul %s menunggu konfirmasi.", b.ID, branchName, dateStr, b.StartTime)
	case string(model.BookingStatusConfirmed):
		customerTitle = "Booking Dikonfirmasi"
		customerMsg = fmt.Sprintf("Booking #%d di %s pada %s pukul %s telah dikonfirmasi.", b.ID, branchName, dateStr, b.StartTime)
	case string(model.BookingStatusWaitingList):
		customerTitle = "Masuk Waiting List"
		customerMsg = fmt.Sprintf("Booking #%d di %s masuk waiting list untuk %s pukul %s.", b.ID, branchName, dateStr, b.StartTime)
	case "promoted":
		customerTitle = "Naik dari Waiting List"
		customerMsg = fmt.Sprintf("Booking #%d di %s berhasil mendapat tempat untuk %s pukul %s.", b.ID, branchName, dateStr, b.StartTime)
	case "cancelled":
		customerTitle = "Booking Dibatalkan"
		customerMsg = fmt.Sprintf("Booking #%d di %s pada %s pukul %s telah dibatalkan.", b.ID, branchName, dateStr, b.StartTime)
	default:
		return
	}

	// Notify customer
	_ = s.notifSvc.SendRef("customer", b.CustomerID, customerTitle, customerMsg, &b.ID)

	// Notify owner: new booking or cancellation
	if event == string(model.BookingStatusPending) || event == string(model.BookingStatusConfirmed) || event == string(model.BookingStatusWaitingList) {
		customer, _ := s.customerRepo.FindByID(b.CustomerID)
		customerName := ""
		if customer != nil {
			customerName = customer.Name
		}
		if branch != nil {
			rest, _ := s.restaurantRepo.FindByBranchID(branch.RestaurantID)
			if rest != nil && rest.BusinessOwnerID != nil {
				ownerMsg := fmt.Sprintf("Booking baru #%d dari %s di %s pada %s pukul %s.", b.ID, customerName, branchName, dateStr, b.StartTime)
				_ = s.notifSvc.SendRef("owner", *rest.BusinessOwnerID, "Booking Baru", ownerMsg, &b.ID)
			}
		}
	}
	if event == "cancelled" {
		customer, _ := s.customerRepo.FindByID(b.CustomerID)
		customerName := ""
		if customer != nil {
			customerName = customer.Name
		}
		if branch != nil {
			rest, _ := s.restaurantRepo.FindByBranchID(branch.RestaurantID)
			if rest != nil && rest.BusinessOwnerID != nil {
				ownerMsg := fmt.Sprintf("Booking #%d dari %s di %s pada %s dibatalkan.", b.ID, customerName, branchName, dateStr)
				_ = s.notifSvc.SendRef("owner", *rest.BusinessOwnerID, "Booking Dibatalkan", ownerMsg, &b.ID)
			}
		}
	}
}

// sendRescheduleNotif notifies customer (in-app, WA, email), owner, and branch staff about a reschedule.
func (s *bookingService) sendRescheduleNotif(b *model.Booking, oldDate time.Time, oldStartTime string) {
	branch, _ := s.branchRepo.FindByID(b.BranchID)
	branchName := ""
	if branch != nil {
		branchName = branch.Name
	}
	customer, _ := s.customerRepo.FindByID(b.CustomerID)
	customerName := ""
	if customer != nil {
		customerName = customer.Name
	}

	oldDateStr := oldDate.Format("02 Jan 2006")
	newDateStr := b.BookingDate.Format("02 Jan 2006")

	customerMsg := fmt.Sprintf("Booking #%d di %s berhasil dijadwalkan ulang dari %s %s ke %s %s. Menunggu konfirmasi.", b.ID, branchName, oldDateStr, oldStartTime, newDateStr, b.StartTime)
	ownerMsg := fmt.Sprintf("Booking #%d dari %s di %s diubah jadwal dari %s %s ke %s %s.", b.ID, customerName, branchName, oldDateStr, oldStartTime, newDateStr, b.StartTime)

	// In-app notification to customer
	if s.notifSvc != nil {
		_ = s.notifSvc.SendRef("customer", b.CustomerID, "Jadwal Booking Diubah", customerMsg, &b.ID)
	}

	// WhatsApp to customer
	if s.waSender != nil && customer != nil && customer.Phone != "" {
		waMsg := fmt.Sprintf(
			"Halo *%s*! 🗓️\n\nJadwal booking Anda telah diubah.\n\n📍 %s\n🔖 ID: #%d\n\n🕐 Jadwal Lama: %s %s\n✅ Jadwal Baru: %s %s\n⏰ %s – %s\n👥 %d tamu\n\nStatus kembali ke *menunggu konfirmasi*.",
			customer.Name, branchName, b.ID, oldDateStr, oldStartTime, newDateStr, b.StartTime, b.StartTime, b.EndTime, b.GuestCount,
		)
		if err := s.waSender.Send(customer.Phone, waMsg); err != nil {
			log.Printf("[WA ERROR] reschedule booking #%d customer_id=%d: %v", b.ID, b.CustomerID, err)
		}
	}

	// Email to customer
	if s.emailSender != nil && customer != nil && customer.Email != "" {
		subject := fmt.Sprintf("Jadwal Booking #%d Diubah — %s", b.ID, branchName)
		body := fmt.Sprintf(`
<p>Halo <strong>%s</strong>,</p>
<p>Jadwal booking Anda telah berhasil diubah. Berikut detailnya:</p>
<table style="border-collapse:collapse;margin:16px 0">
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">Restoran</td><td><strong>%s</strong></td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">ID Booking</td><td>#%d</td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">Jadwal Lama</td><td>%s pukul %s</td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">Jadwal Baru</td><td>%s pukul %s – %s</td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">Tamu</td><td>%d orang</td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">Status</td><td>Menunggu konfirmasi</td></tr>
</table>
<p>Kami akan segera mengonfirmasi jadwal baru Anda.</p>`,
			customer.Name, branchName, b.ID, oldDateStr, oldStartTime, newDateStr, b.StartTime, b.EndTime, b.GuestCount)
		if err := s.emailSender.Send(customer.Email, subject, body); err != nil {
			log.Printf("[EMAIL ERROR] reschedule booking #%d customer_id=%d: %v", b.ID, b.CustomerID, err)
		}
	}

	// In-app notification to owner
	if s.notifSvc != nil && branch != nil {
		rest, _ := s.restaurantRepo.FindByBranchID(branch.RestaurantID)
		if rest != nil && rest.BusinessOwnerID != nil {
			_ = s.notifSvc.SendRef("owner", *rest.BusinessOwnerID, "Jadwal Booking Diubah", ownerMsg, &b.ID)
		}
	}

	// In-app notification to all active staff in the branch
	if s.notifSvc != nil && s.staffRepo != nil && branch != nil {
		staffList, _ := s.staffRepo.FindByBranchID(b.BranchID)
		for _, st := range staffList {
			_ = s.notifSvc.SendRef("staff", st.ID, "Jadwal Booking Diubah", ownerMsg, &b.ID)
		}
	}
}

func (s *bookingService) sendReminderNotif(b *model.Booking) {
	customer, _ := s.customerRepo.FindByID(b.CustomerID)
	branch, _ := s.branchRepo.FindByID(b.BranchID)
	if customer == nil || branch == nil {
		return
	}
	dateStr := b.BookingDate.Format("02 Jan 2006")

	// In-app
	if s.notifSvc != nil {
		_ = s.notifSvc.SendRef("customer", b.CustomerID,
			"Pengingat Booking Besok",
			fmt.Sprintf("Booking #%d di %s besok (%s) pukul %s. Jangan lupa hadir ya!", b.ID, branch.Name, dateStr, b.StartTime),
			&b.ID,
		)
	}

	// WhatsApp
	if s.waSender != nil && customer.Phone != "" {
		msg := fmt.Sprintf(
			"Halo *%s*! 🔔\n\nIni pengingat booking Anda *besok*.\n\n📍 %s\n📅 %s\n⏰ %s – %s\n👥 %d tamu\n🔖 ID: #%d\n\nSampai jumpa!",
			customer.Name, branch.Name, dateStr, b.StartTime, b.EndTime, b.GuestCount, b.ID,
		)
		if err := s.waSender.Send(customer.Phone, msg); err != nil {
			log.Printf("[WA ERROR] reminder booking #%d customer_id=%d: %v", b.ID, b.CustomerID, err)
		}
	}

	// Email
	if s.emailSender != nil && customer.Email != "" {
		subject := fmt.Sprintf("Pengingat Booking Besok — %s", branch.Name)
		body := fmt.Sprintf(`
<p>Halo <strong>%s</strong>,</p>
<p>Ini pengingat bahwa Anda memiliki booking <strong>besok</strong>!</p>
<table style="border-collapse:collapse;margin:16px 0">
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">Restoran</td><td><strong>%s</strong></td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">Tanggal</td><td>%s</td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">Waktu</td><td>%s – %s</td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">Tamu</td><td>%d orang</td></tr>
  <tr><td style="padding:4px 12px 4px 0;color:#6b7280">ID Booking</td><td>#%d</td></tr>
</table>
<p>Sampai jumpa besok! 🎉</p>`,
			customer.Name, branch.Name, dateStr, b.StartTime, b.EndTime, b.GuestCount, b.ID)
		if err := s.emailSender.Send(customer.Email, subject, body); err != nil {
			log.Printf("[EMAIL ERROR] reminder booking #%d customer_id=%d: %v", b.ID, b.CustomerID, err)
		}
	}
}
