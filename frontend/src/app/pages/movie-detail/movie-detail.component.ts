import { Component, OnInit, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { MovieService } from '../../services/movie.service';
import {
  ApiMovie,
  MovieTheaterListResponse,
  ShowtimeItem,
  TheaterSchedule
} from '../../models/movie.models';

@Component({
  selector: 'app-movie-detail',
  standalone: true,
  imports: [CommonModule, RouterLink, FormsModule],
  templateUrl: './movie-detail.component.html',
  styleUrl: './movie-detail.component.scss'
})
export class MovieDetailComponent implements OnInit {
  movieId = signal('');
  movie = signal<ApiMovie | null>(null);
  theaterData = signal<MovieTheaterListResponse | null>(null);
  loadingMovie = signal(true);
  loadingTheaters = signal(true);
  movieError = signal<string | null>(null);
  showtimesError = signal<string | null>(null);

  selectedDate = signal(this.clampToBookingWindow(this.todayYmd()));

  selectedShowtime = signal<{ showtime: ShowtimeItem; theater: TheaterSchedule } | null>(null);

  constructor(
    private route: ActivatedRoute,
    private router: Router,
    private movieService: MovieService
  ) {}

  ngOnInit() {
    this.route.paramMap.subscribe(pm => {
      const id = pm.get('movieId');
      if (!id) {
        void this.router.navigate(['/']);
        return;
      }
      this.movieId.set(id);
      this.selectedShowtime.set(null);
      this.selectedDate.set(this.clampToBookingWindow(this.selectedDate()));
      this.loadMovie(id);
      this.loadTheaters(id);
    });
  }

  /** Today in local calendar (YYYY-MM-DD). */
  todayYmd(): string {
    return this.toYmd(new Date());
  }

  /** Last day (inclusive) we sell tickets: today + 14 days, local calendar. */
  maxBookingYmd(): string {
    const d = new Date();
    d.setDate(d.getDate() + 14);
    return this.toYmd(d);
  }

  private toYmd(d: Date): string {
    const y = d.getFullYear();
    const m = String(d.getMonth() + 1).padStart(2, '0');
    const day = String(d.getDate()).padStart(2, '0');
    return `${y}-${m}-${day}`;
  }

  /** Keep chosen day within [today, today+14]. */
  clampToBookingWindow(ymd: string): string {
    const min = this.todayYmd();
    const max = this.maxBookingYmd();
    if (ymd < min) {
      return min;
    }
    if (ymd > max) {
      return max;
    }
    return ymd;
  }

  onDateChange() {
    const next = this.clampToBookingWindow(this.selectedDate());
    if (next !== this.selectedDate()) {
      this.selectedDate.set(next);
    }
    this.selectedShowtime.set(null);
    this.loadTheaters(this.movieId());
  }

  private loadMovie(id: string) {
    this.loadingMovie.set(true);
    this.movieError.set(null);
    this.movieService.getMovie(id).subscribe({
      next: m => {
        this.movie.set(m);
        this.loadingMovie.set(false);
      },
      error: () => {
        this.movieError.set('This movie could not be loaded.');
        this.loadingMovie.set(false);
      }
    });
  }

  private loadTheaters(id: string) {
    this.loadingTheaters.set(true);
    this.showtimesError.set(null);
    this.movieService.getTheatersForMovie(id, this.selectedDate()).subscribe({
      next: data => {
        this.theaterData.set(data);
        this.loadingTheaters.set(false);
      },
      error: () => {
        this.showtimesError.set('Showtimes could not be loaded. Try another date.');
        this.loadingTheaters.set(false);
      }
    });
  }

  pickShowtime(theater: TheaterSchedule, st: ShowtimeItem) {
    this.selectedShowtime.set({ theater, showtime: st });
  }

  isSelected(theater: TheaterSchedule, st: ShowtimeItem): boolean {
    const sel = this.selectedShowtime();
    if (!sel) {
      return false;
    }
    return (
      sel.theater.theaterId === theater.theaterId && sel.showtime.showtimeId === st.showtimeId
    );
  }

  continueToSeats() {
    const sel = this.selectedShowtime();
    const id = this.movieId();
    if (!sel) {
      return;
    }
    void this.router.navigate(['/movies', id, 'seats'], {
      queryParams: {
        theaterId: sel.theater.theaterId,
        showTime: sel.showtime.startTime
      }
    });
  }

  readonly posterFallback = 'https://placehold.co/400x600/1e1b2e/a78bfa?text=Poster';

  posterUrl(url?: string | null): string {
    if (url) {
      return url;
    }
    return this.posterFallback;
  }

  onPosterError(event: Event) {
    const img = event.target as HTMLImageElement;
    if (img.dataset['usedFallback']) {
      return;
    }
    img.dataset['usedFallback'] = '1';
    img.src = this.posterFallback;
  }
}
