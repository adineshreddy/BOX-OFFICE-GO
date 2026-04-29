import { CommonModule } from '@angular/common';
import { Component, OnInit, computed, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { RouterLink } from '@angular/router';
import { HttpErrorResponse } from '@angular/common/http';
import { BookingService, UserBooking } from '../../services/booking.service';
import { ToastService } from '../../services/toast.service';

@Component({
  selector: 'app-my-bookings',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterLink],
  templateUrl: './my-bookings.component.html',
  styleUrl: './my-bookings.component.scss'
})
export class MyBookingsComponent implements OnInit {
  bookings = signal<UserBooking[]>([]);
  searchQuery = signal('');
  statusFilter = signal<'ALL' | 'CONFIRMED' | 'CANCELLED'>('ALL');
  sortBy = signal<'showtime_desc' | 'showtime_asc' | 'confirmed_desc'>('showtime_desc');
  loading = signal(true);
  listError = signal<string | null>(null);

  downloadingId = signal<string | null>(null);
  downloadError = signal<string | null>(null);

  cancellingId = signal<string | null>(null);
  cancelError = signal<string | null>(null);

  readonly filteredBookings = computed(() => {
    const query = this.searchQuery().trim().toLowerCase();
    const status = this.statusFilter();
    const sortBy = this.sortBy();

    const filtered = this.bookings().filter(b => {
      const normalizedStatus = (b.status ?? '').toUpperCase();
      if (status !== 'ALL' && normalizedStatus !== status) {
        return false;
      }
      if (!query) {
        return true;
      }
      const haystack = [
        b.movieTitle,
        b.theaterName,
        b.city,
        b.screenName,
        ...(b.seatNumbers ?? []),
        b.bookingId
      ]
        .join(' ')
        .toLowerCase();
      return haystack.includes(query);
    });

    return filtered.sort((a, b) => this.compareBookings(a, b, sortBy));
  });

  readonly upcomingBookings = computed(() => this.filteredBookings().filter(b => !this.isPastShowtime(b)));
  readonly pastBookings = computed(() => this.filteredBookings().filter(b => this.isPastShowtime(b)));

  constructor(
    private bookingService: BookingService,
    private toast: ToastService
  ) {}

  ngOnInit(): void {
    this.loadBookings();
  }

  resultLabel(): string {
    const count = this.filteredBookings().length;
    if (count === 0) {
      return 'No matching bookings';
    }
    if (count === 1) {
      return '1 booking';
    }
    return `${count} bookings`;
  }

  loadBookings(): void {
    this.loading.set(true);
    this.listError.set(null);
    this.bookingService.getMyBookings().subscribe({
      next: res => {
        this.bookings.set(res.bookings ?? []);
        this.loading.set(false);
      },
      error: (err: HttpErrorResponse) => {
        this.loading.set(false);
        const msg =
          err.error && typeof err.error === 'object' && typeof err.error.message === 'string'
            ? err.error.message
            : null;
        this.listError.set(msg ?? 'Could not load your bookings. Please try again.');
      }
    });
  }

  isConfirmed(b: UserBooking): boolean {
    return (b.status ?? '').toUpperCase() === 'CONFIRMED';
  }

  /** Cancellation is only allowed >= 1 hour before showtime (same rule as the API). */
  canCancel(b: UserBooking): boolean {
    if (!this.isConfirmed(b)) {
      return false;
    }
    const raw = b.showTime;
    if (!raw) {
      return true;
    }
    const d = new Date(raw);
    if (Number.isNaN(d.getTime())) {
      return true;
    }
    return d.getTime() - Date.now() >= 60 * 60 * 1000;
  }

  formatWhen(b: UserBooking): string {
    const raw = b.showTime;
    if (!raw) {
      return '—';
    }
    const d = new Date(raw);
    if (Number.isNaN(d.getTime())) {
      return raw;
    }
    return d.toLocaleString(undefined, {
      weekday: 'short',
      month: 'short',
      day: 'numeric',
      year: 'numeric',
      hour: 'numeric',
      minute: '2-digit'
    });
  }

  formatConfirmed(b: UserBooking): string {
    const raw = b.confirmedAt;
    if (!raw) {
      return '—';
    }
    const d = new Date(raw);
    if (Number.isNaN(d.getTime())) {
      return raw;
    }
    return d.toLocaleString(undefined, { dateStyle: 'medium', timeStyle: 'short' });
  }

  isPastShowtime(b: UserBooking): boolean {
    const showtimeMs = this.getShowtimeTimestamp(b);
    if (showtimeMs === null) {
      return false;
    }
    return showtimeMs <= Date.now();
  }

  downloadTicket(bookingId: string): void {
    if (this.downloadingId()) {
      return;
    }
    this.downloadingId.set(bookingId);
    this.downloadError.set(null);

    this.bookingService.downloadTicket(bookingId).subscribe({
      next: blob => {
        this.downloadingId.set(null);
        const url = window.URL.createObjectURL(blob);
        const anchor = document.createElement('a');
        anchor.href = url;
        anchor.download = `ticket_${bookingId}.pdf`;
        document.body.appendChild(anchor);
        anchor.click();
        anchor.remove();
        window.URL.revokeObjectURL(url);
        this.toast.show('Ticket downloaded');
      },
      error: () => {
        this.downloadingId.set(null);
        this.downloadError.set('Could not download that ticket. Confirmed bookings only.');
      }
    });
  }

  cancelBooking(booking: UserBooking): void {
    if (!this.canCancel(booking) || this.cancellingId()) {
      return;
    }
    if (!confirm('Cancel this booking? This cannot be undone.')) {
      return;
    }
    this.cancellingId.set(booking.bookingId);
    this.cancelError.set(null);

    this.bookingService.cancelBooking(booking.bookingId).subscribe({
      next: () => {
        this.cancellingId.set(null);
        this.toast.show('Booking cancelled');
        this.loadBookings();
      },
      error: (err: HttpErrorResponse) => {
        this.cancellingId.set(null);
        const msg =
          err.error && typeof err.error === 'object' && typeof err.error.message === 'string'
            ? err.error.message
            : null;
        this.cancelError.set(msg ?? 'Could not cancel this booking.');
      }
    });
  }

  private compareBookings(
    a: UserBooking,
    b: UserBooking,
    sortBy: 'showtime_desc' | 'showtime_asc' | 'confirmed_desc'
  ): number {
    const aShowtime = this.getShowtimeTimestamp(a) ?? 0;
    const bShowtime = this.getShowtimeTimestamp(b) ?? 0;
    const aConfirmed = this.getConfirmedTimestamp(a) ?? 0;
    const bConfirmed = this.getConfirmedTimestamp(b) ?? 0;

    if (sortBy === 'showtime_asc') {
      return aShowtime - bShowtime;
    }
    if (sortBy === 'confirmed_desc') {
      return bConfirmed - aConfirmed;
    }
    return bShowtime - aShowtime;
  }

  private getShowtimeTimestamp(b: UserBooking): number | null {
    const raw = b.showTime;
    if (!raw) {
      return null;
    }
    const d = new Date(raw);
    return Number.isNaN(d.getTime()) ? null : d.getTime();
  }

  private getConfirmedTimestamp(b: UserBooking): number | null {
    const raw = b.confirmedAt;
    if (!raw) {
      return null;
    }
    const d = new Date(raw);
    return Number.isNaN(d.getTime()) ? null : d.getTime();
  }
}
