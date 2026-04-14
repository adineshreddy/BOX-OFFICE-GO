import { CommonModule } from '@angular/common';
import { HttpErrorResponse } from '@angular/common/http';
import { Component, OnDestroy, OnInit, computed, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { AuthService } from '../../services/auth.service';
import { BookingService } from '../../services/booking.service';

@Component({
  selector: 'app-payment',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterLink],
  templateUrl: './payment.component.html',
  styleUrl: './payment.component.scss'
})
export class PaymentComponent implements OnInit, OnDestroy {
  movieId = signal('');
  theaterId = signal<string | null>(null);
  showTime = signal<string | null>(null);
  holdId = signal<string | null>(null);
  holdExpiresAt = signal<Date | null>(null);

  paymentMethod = signal('card');
  loading = signal(true);
  processing = signal(false);
  error = signal<string | null>(null);
  success = signal<string | null>(null);
  timedOut = signal(false);
  holdReleased = signal(false);
  checkoutCompleted = signal(false);

  remainingSeconds = signal(0);
  private timerHandle: ReturnType<typeof setInterval> | null = null;

  readonly formattedTimer = computed(() => {
    const total = Math.max(0, this.remainingSeconds());
    const mm = String(Math.floor(total / 60)).padStart(2, '0');
    const ss = String(total % 60).padStart(2, '0');
    return `${mm}:${ss}`;
  });

  constructor(
    private route: ActivatedRoute,
    private router: Router,
    private authService: AuthService,
    private bookingService: BookingService
  ) {}

  ngOnInit(): void {
    this.route.paramMap.subscribe(pm => {
      this.movieId.set(pm.get('movieId') ?? '');
    });

    this.route.queryParamMap.subscribe(q => {
      this.theaterId.set(q.get('theaterId'));
      this.showTime.set(q.get('showTime'));
      this.holdId.set(q.get('holdId'));

      const expiresRaw = q.get('holdExpiresAt');
      const parsed = expiresRaw ? new Date(expiresRaw) : null;
      this.holdExpiresAt.set(parsed && !Number.isNaN(parsed.getTime()) ? parsed : null);
      this.bootstrapPage();
    });
  }

  ngOnDestroy(): void {
    this.stopTimer();
    if (!this.checkoutCompleted() && !this.holdReleased() && this.holdId()) {
      this.releaseHold();
    }
  }

  submitPayment(): void {
    if (this.processing() || this.timedOut()) {
      return;
    }
    const user = this.authService.getUser();
    const holdId = this.holdId();
    if (!user?.id || !holdId) {
      this.error.set('Missing payment context. Please reselect seats.');
      return;
    }

    this.processing.set(true);
    this.error.set(null);

    this.bookingService
      .checkoutBookingHold({
        holdId,
        paymentMethod: this.paymentMethod(),
        idempotencyKey: this.newIdempotencyKey()
      })
      .subscribe({
        next: res => {
          const booking = res?.booking;
          this.processing.set(false);
          this.checkoutCompleted.set(true);
          const bookingId = booking?.bookingId;
          if (bookingId) {
            void this.router.navigate(['/movies', this.movieId(), 'booking-success', bookingId], {
              queryParams: {
                seats: (booking.seatNumbers ?? []).join(','),
                total: booking.totalAmount ?? null
              }
            });
            return;
          }
          this.success.set('Payment successful. Booking confirmed.');
        },
        error: (err: HttpErrorResponse) => {
          this.processing.set(false);
          const serverMessage =
            err.error && typeof err.error === 'object' && typeof err.error.message === 'string'
              ? err.error.message
              : null;
          if (err.status === 409) {
            this.error.set(serverMessage ?? 'Your seat hold has expired. Please select seats again.');
            void this.navigateBackToSeats(true);
            return;
          }
          if (err.status === 402) {
            this.error.set(serverMessage ?? 'Payment was declined. Please try another method.');
            return;
          }
          if (err.status === 401) {
            this.error.set(serverMessage ?? 'Your session expired. Please sign in again.');
            void this.router.navigate(['/login'], { queryParams: { from: 'booking' } });
            return;
          }
          this.error.set(serverMessage ?? 'Payment failed. Please try again.');
        }
      });
  }

  private bootstrapPage(): void {
    this.loading.set(true);
    this.error.set(null);
    this.success.set(null);
    this.timedOut.set(false);
    this.holdReleased.set(false);
    this.checkoutCompleted.set(false);
    this.stopTimer();

    const user = this.authService.getUser();
    if (!user?.id) {
      void this.router.navigate(['/login'], { queryParams: { from: 'booking' } });
      return;
    }

    if (!this.holdId() || !this.holdExpiresAt()) {
      void this.navigateBackToSeats(false, 'missing_hold');
      return;
    }

    this.startTimer();
    this.loading.set(false);
  }

  private startTimer(): void {
    this.tickTimer();
    this.timerHandle = setInterval(() => this.tickTimer(), 1000);
  }

  private stopTimer(): void {
    if (this.timerHandle) {
      clearInterval(this.timerHandle);
      this.timerHandle = null;
    }
  }

  private tickTimer(): void {
    const expiresAt = this.holdExpiresAt();
    if (!expiresAt) {
      this.remainingSeconds.set(0);
      return;
    }
    const leftSeconds = Math.floor((expiresAt.getTime() - Date.now()) / 1000);
    this.remainingSeconds.set(Math.max(0, leftSeconds));
    if (leftSeconds <= 0 && !this.timedOut()) {
      this.timedOut.set(true);
      this.stopTimer();
      this.error.set('Time is up. Releasing your held seats...');
      this.releaseHold(() => {
        void this.navigateBackToSeats(true);
      });
    }
  }

  private releaseHold(onDone?: () => void): void {
    const holdId = this.holdId();
    if (!holdId || this.holdReleased()) {
      onDone?.();
      return;
    }
    this.holdReleased.set(true);
    this.bookingService.releaseBookingHold(holdId).subscribe({
      next: () => onDone?.(),
      error: () => onDone?.()
    });
  }

  private navigateBackToSeats(expired: boolean, reason: string | null = null): Promise<boolean> {
    return this.router.navigate(['/movies', this.movieId(), 'seats'], {
      queryParams: {
        theaterId: this.theaterId(),
        showTime: this.showTime(),
        expired: expired ? '1' : null,
        reason
      }
    });
  }

  private newIdempotencyKey(): string {
    if (typeof crypto !== 'undefined' && 'randomUUID' in crypto) {
      return crypto.randomUUID();
    }
    return `idem_${Date.now()}`;
  }
}
