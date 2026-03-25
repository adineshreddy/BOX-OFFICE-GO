import { Component, OnInit, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { MovieService } from '../../services/movie.service';
import { SeatItem, SeatMapResponse, ShowDetails } from '../../models/movie.models';
import { AuthService } from '../../services/auth.service';
import { BookingService } from '../../services/booking.service';
import { HttpErrorResponse } from '@angular/common/http';

type PendingBookingHold = {
  movieId: string;
  theaterId: string;
  showTime: string;
  showtimeId: string;
  ticketCount: number;
  selectedSeats: string[];
};

const PENDING_HOLD_KEY = 'pending_booking_hold';

@Component({
  selector: 'app-movie-seats',
  standalone: true,
  imports: [CommonModule, RouterLink, FormsModule],
  templateUrl: './movie-seats.component.html',
  styleUrl: './movie-seats.component.scss'
})
export class MovieSeatsComponent implements OnInit {
  movieId = signal('');
  theaterId = signal<string | null>(null);
  showTime = signal<string | null>(null);
  loading = signal(true);
  error = signal<string | null>(null);

  showDetails = signal<ShowDetails | null>(null);
  seatMap = signal<SeatMapResponse | null>(null);

  ticketPromptOpen = signal(true);
  ticketCount = signal(2);
  selectedSeats = signal<string[]>([]);
  selectionError = signal<string | null>(null);
  simulatedSoldSeats = signal<Set<string>>(new Set());

  holdInProgress = signal(false);
  holdError = signal<string | null>(null);
  holdSuccess = signal<string | null>(null);
  holdCreated = signal(false);
  holdId = signal<string | null>(null);

  pendingHoldAttempted = signal(false);

  constructor(
    private route: ActivatedRoute,
    private movieService: MovieService,
    private bookingService: BookingService,
    private authService: AuthService,
    private router: Router
  ) {}

  ngOnInit() {
    this.route.paramMap.subscribe(pm => {
      this.movieId.set(pm.get('movieId') ?? '');
    });
    this.route.queryParamMap.subscribe(q => {
      this.theaterId.set(q.get('theaterId'));
      this.showTime.set(q.get('showTime'));
      this.loadSeatExperience();
    });
  }

  private loadSeatExperience() {
    const movieId = this.movieId();
    const theaterId = this.theaterId();
    const showTime = this.showTime();
    if (!movieId || !theaterId || !showTime) {
      this.loading.set(false);
      this.error.set('Missing theater or show time. Go back and pick a showtime.');
      return;
    }

    this.loading.set(true);
    this.error.set(null);
    this.selectionError.set(null);
    this.holdError.set(null);
    this.holdSuccess.set(null);
    this.holdCreated.set(false);
    this.holdId.set(null);
    this.pendingHoldAttempted.set(false);
    this.selectedSeats.set([]);
    this.ticketPromptOpen.set(true);

    this.movieService.getShowDetailsBySelection(movieId, theaterId, showTime).subscribe({
      next: details => this.showDetails.set(details),
      error: () => this.error.set('Could not load show details. Please go back and try again.')
    });

    this.movieService.getSeatMapBySelection(movieId, theaterId, showTime).subscribe({
      next: seatMap => {
        this.seatMap.set(seatMap);
        this.simulatedSoldSeats.set(this.buildSimulatedSoldSeats(seatMap));
        this.loading.set(false);
        this.tryAutoHoldFromPending();
      },
      error: () => {
        this.error.set('Could not load theater seat map. Please go back and pick another show.');
        this.loading.set(false);
      }
    });
  }

  chooseTicketCount(count: number) {
    if (this.holdCreated() || this.holdInProgress()) {
      return;
    }
    this.ticketCount.set(count);
    this.selectedSeats.set([]);
    this.selectionError.set(null);
    this.ticketPromptOpen.set(false);
  }

  onTicketCountChange(rawValue: string) {
    if (this.holdCreated() || this.holdInProgress()) {
      return;
    }
    const parsed = Number(rawValue);
    if (!Number.isInteger(parsed) || parsed < 1) {
      return;
    }
    this.ticketCount.set(parsed);
    const current = this.selectedSeats();
    if (current.length > parsed) {
      this.selectedSeats.set(current.slice(0, parsed));
    }
    this.selectionError.set(null);
  }

  ticketOptions(): number[] {
    const maxSeats = this.selectableSeatCount();
    const max = Math.max(1, Math.min(maxSeats, 10));
    return Array.from({ length: max }, (_, i) => i + 1);
  }

  isSeatSold(seat: SeatItem): boolean {
    if (!seat.isAvailable || seat.isHeld) {
      return true;
    }
    return this.simulatedSoldSeats().has(seat.seatNumber);
  }

  isSeatUnavailable(seat: SeatItem): boolean {
    return this.isSeatSold(seat);
  }

  isSeatSelected(seatNumber: string): boolean {
    return this.selectedSeats().includes(seatNumber);
  }

  toggleSeat(seat: SeatItem) {
    if (this.holdCreated() || this.holdInProgress()) {
      return;
    }
    if (this.isSeatUnavailable(seat)) {
      return;
    }

    const current = this.selectedSeats();
    const alreadySelected = current.includes(seat.seatNumber);
    if (alreadySelected) {
      this.selectedSeats.set(current.filter(s => s !== seat.seatNumber));
      this.selectionError.set(null);
      return;
    }

    if (current.length >= this.ticketCount()) {
      this.selectionError.set(`You can select only ${this.ticketCount()} seat(s).`);
      return;
    }

    this.selectedSeats.set([...current, seat.seatNumber]);
    this.selectionError.set(null);
  }

  confirmSeatsHold() {
    if (this.holdInProgress() || this.holdCreated()) {
      return;
    }
    this.selectionError.set(null);
    this.holdError.set(null);
    this.holdSuccess.set(null);

    const map = this.seatMap();
    if (!map) {
      this.holdError.set('Seat map not loaded yet.');
      return;
    }

    const selected = this.selectedSeats();
    if (selected.length !== this.ticketCount()) {
      this.selectionError.set(`Select exactly ${this.ticketCount()} seat(s).`);
      return;
    }

    const user = this.authService.getUser();
    if (!user?.id) {
      // Persist the pending selection so that after login we can immediately create the hold.
      const pending: PendingBookingHold = {
        movieId: this.movieId(),
        theaterId: this.theaterId() ?? '',
        showTime: this.showTime() ?? '',
        showtimeId: map.showtimeId,
        ticketCount: this.ticketCount(),
        selectedSeats: selected
      };
      if (!pending.theaterId || !pending.showTime) {
        this.holdError.set('Missing booking context. Please go back and reselect seats.');
        return;
      }
      localStorage.setItem(PENDING_HOLD_KEY, JSON.stringify(pending));
      // Redirect to login; on success we send the user back to this seat page automatically.
      void this.router.navigate(['/login'], { queryParams: { from: 'booking' } });
      return;
    }

    this.holdInProgress.set(true);
    this.bookingService
      .createBookingHold({
        userId: user.id,
        showtimeId: map.showtimeId,
        seatNumbers: selected
      })
      .subscribe({
        next: (res: any) => {
          const hold = res?.hold ?? res;
          this.holdId.set(hold?.holdId ?? null);
          this.holdCreated.set(true);
          this.holdInProgress.set(false);
          this.holdSuccess.set('Seats held successfully.');
          this.ticketPromptOpen.set(false);
          this.refreshSeatMap(map.showtimeId);
          this.clearPendingHold();
        },
        error: (err: HttpErrorResponse) => {
          this.holdInProgress.set(false);
          // Another user could have taken the seats between the time they were displayed and this request.
          if (err.status === 409) {
            this.selectionError.set('Some seats were just taken. Please reselect.');
            this.selectedSeats.set([]);
            this.holdError.set(null);
            this.clearPendingHold();
          } else {
            this.holdError.set('Could not hold seats. Please try again.');
            this.clearPendingHold();
          }
          this.refreshSeatMap(map.showtimeId);
        }
      });
  }

  private refreshSeatMap(_showtimeId: string) {
    const movieId = this.movieId();
    const theaterId = this.theaterId();
    const showTime = this.showTime();
    if (!movieId || !theaterId || !showTime) {
      return;
    }

    this.movieService.getSeatMapBySelection(movieId, theaterId, showTime).subscribe({
      next: seatMap => {
        this.seatMap.set(seatMap);
        this.simulatedSoldSeats.set(this.buildSimulatedSoldSeats(seatMap));
      }
    });
  }

  private selectableSeatCount(): number {
    const map = this.seatMap();
    if (!map) {
      return 0;
    }
    return map.rows.flatMap(row => row.seats).filter(seat => !this.isSeatSold(seat)).length;
  }

  private buildSimulatedSoldSeats(map: SeatMapResponse): Set<string> {
    const candidateSeats = map.rows.flatMap(row =>
      row.seats.filter(seat => seat.isAvailable && !seat.isHeld).map(seat => seat.seatNumber)
    );
    const soldCount = Math.floor(candidateSeats.length * 0.2);
    if (soldCount <= 0) {
      return new Set();
    }

    // Truly random per page refresh (uniform across all candidate seats).
    const shuffled = [...candidateSeats];
    for (let i = shuffled.length - 1; i > 0; i--) {
      const j = Math.floor(Math.random() * (i + 1));
      [shuffled[i], shuffled[j]] = [shuffled[j], shuffled[i]];
    }

    return new Set(shuffled.slice(0, soldCount));
  }

  private tryAutoHoldFromPending() {
    const pending = this.getPendingHold();
    if (!pending) {
      return;
    }

    if (this.pendingHoldAttempted()) {
      return;
    }

    if (
      pending.movieId !== this.movieId() ||
      pending.theaterId !== (this.theaterId() ?? '') ||
      pending.showTime !== (this.showTime() ?? '')
    ) {
      return;
    }

    const user = this.authService.getUser();
    if (!user?.id) {
      return;
    }

    // Only attempt once per page load (prevents re-creating holds in case seat-map refresh fires again).
    this.pendingHoldAttempted.set(true);

    this.ticketCount.set(pending.ticketCount);
    this.selectedSeats.set(pending.selectedSeats);
    this.ticketPromptOpen.set(false);

    // Attempt hold immediately (preference B).
    this.confirmSeatsHold();
  }

  private getPendingHold(): PendingBookingHold | null {
    try {
      const raw = localStorage.getItem(PENDING_HOLD_KEY);
      if (!raw) {
        return null;
      }
      return JSON.parse(raw) as PendingBookingHold;
    } catch {
      return null;
    }
  }

  private clearPendingHold() {
    localStorage.removeItem(PENDING_HOLD_KEY);
  }
}
