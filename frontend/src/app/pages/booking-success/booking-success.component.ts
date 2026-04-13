import { CommonModule } from '@angular/common';
import { Component, signal } from '@angular/core';
import { ActivatedRoute, RouterLink } from '@angular/router';
import { BookingService } from '../../services/booking.service';

@Component({
  selector: 'app-booking-success',
  standalone: true,
  imports: [CommonModule, RouterLink],
  templateUrl: './booking-success.component.html',
  styleUrl: './booking-success.component.scss'
})
export class BookingSuccessComponent {
  movieId = signal('');
  bookingId = signal('');
  seats = signal<string[]>([]);
  total = signal<number | null>(null);

  downloading = signal(false);
  downloadError = signal<string | null>(null);

  constructor(
    private route: ActivatedRoute,
    private bookingService: BookingService
  ) {
    this.route.paramMap.subscribe(pm => {
      this.movieId.set(pm.get('movieId') ?? '');
      this.bookingId.set(pm.get('bookingId') ?? '');
    });

    this.route.queryParamMap.subscribe(q => {
      const seatsText = q.get('seats');
      this.seats.set(
        seatsText
          ? seatsText
              .split(',')
              .map(s => s.trim())
              .filter(Boolean)
          : []
      );
      const totalRaw = q.get('total');
      const parsed = totalRaw ? Number(totalRaw) : null;
      this.total.set(parsed !== null && Number.isFinite(parsed) ? parsed : null);
    });
  }

  downloadTicket(): void {
    if (this.downloading() || !this.bookingId()) {
      return;
    }
    this.downloading.set(true);
    this.downloadError.set(null);

    this.bookingService.downloadTicket(this.bookingId()).subscribe({
      next: blob => {
        this.downloading.set(false);
        const url = window.URL.createObjectURL(blob);
        const anchor = document.createElement('a');
        anchor.href = url;
        anchor.download = `ticket_${this.bookingId()}.pdf`;
        document.body.appendChild(anchor);
        anchor.click();
        anchor.remove();
        window.URL.revokeObjectURL(url);
      },
      error: () => {
        this.downloading.set(false);
        this.downloadError.set('Could not download ticket right now. Please try again.');
      }
    });
  }
}
