import { Component, OnInit, OnDestroy, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Router, RouterLink } from '@angular/router';
import { MovieService } from '../../services/movie.service';
import { ApiMovie } from '../../models/movie.models';

@Component({
  selector: 'app-home',
  standalone: true,
  imports: [CommonModule, RouterLink],
  templateUrl: './home.component.html',
  styleUrl: './home.component.scss'
})
export class HomeComponent implements OnInit, OnDestroy {
  currentHeroIndex = signal(0);
  currentMoviePage = signal(0);
  moviesPerPage = 5;
  private autoRotateInterval?: number;

  movies = signal<ApiMovie[]>([]);
  loadingMovies = signal(true);
  moviesError = signal<string | null>(null);

  visibleMovies = computed(() => {
    const list = this.movies();
    const start = this.currentMoviePage() * this.moviesPerPage;
    const end = start + this.moviesPerPage;
    return list.slice(start, end);
  });

  canGoNext = computed(() => {
    const list = this.movies();
    return (this.currentMoviePage() + 1) * this.moviesPerPage < list.length;
  });

  canGoPrev = computed(() => {
    return this.currentMoviePage() > 0;
  });

  heroBanners = [
    {
      type: 'app-promo',
      title: 'Box Office Go App',
      subtitle: 'Coming Soon',
      description: 'Download our brand new app for hassle-free booking',
      buttonText: 'Coming Soon',
      bgColor: 'linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f3460 100%)',
      textColor: '#fff'
    },
    {
      type: 'greeting-combined',
      title: 'CHEERS',
      subtitle: 'Wish a Happy Birthday or Anniversary',
      description: 'on the Big Screen!',
      buttonText: 'Book A Greeting',
      bgColor: 'linear-gradient(135deg, #f093fb 0%, #f5576c 100%)',
      textColor: '#fff',
      emoji: '🎂💕'
    }
  ];

  constructor(
    private router: Router,
    private movieService: MovieService
  ) {}

  ngOnInit() {
    this.startAutoRotate();
    this.loadMovies();
  }

  private loadMovies() {
    this.loadingMovies.set(true);
    this.moviesError.set(null);
    this.movieService.listMovies().subscribe({
      next: res => {
        this.movies.set(res.movies);
        this.loadingMovies.set(false);
        this.currentMoviePage.set(0);
      },
      error: () => {
        this.moviesError.set('Could not load movies. Start the API (see README) and refresh.');
        this.loadingMovies.set(false);
      }
    });
  }

  readonly posterFallback = 'https://placehold.co/300x450/1e1b2e/a78bfa?text=Poster';

  posterUrl(movie: ApiMovie): string {
    if (movie.posterUrl) {
      return movie.posterUrl;
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

  ngOnDestroy() {
    this.stopAutoRotate();
  }

  startAutoRotate() {
    this.stopAutoRotate();
    this.autoRotateInterval = window.setInterval(() => {
      this.nextHero();
    }, 3000);
  }

  stopAutoRotate() {
    if (this.autoRotateInterval) {
      clearInterval(this.autoRotateInterval);
      this.autoRotateInterval = undefined;
    }
  }

  pauseAutoRotate() {
    this.stopAutoRotate();
  }

  resumeAutoRotate() {
    this.startAutoRotate();
  }

  nextHero() {
    this.currentHeroIndex.set(
      (this.currentHeroIndex() + 1) % this.heroBanners.length
    );
  }

  prevHero() {
    this.currentHeroIndex.set(
      this.currentHeroIndex() === 0
        ? this.heroBanners.length - 1
        : this.currentHeroIndex() - 1
    );
  }

  goToBanner(index: number) {
    this.currentHeroIndex.set(index);
    this.stopAutoRotate();
    this.startAutoRotate();
  }

  nextMovies() {
    if (this.canGoNext()) {
      this.currentMoviePage.set(this.currentMoviePage() + 1);
    }
  }

  prevMovies() {
    if (this.canGoPrev()) {
      this.currentMoviePage.set(this.currentMoviePage() - 1);
    }
  }

  goToSignIn() {
    this.router.navigate(['/login'], { queryParams: { from: 'see-all-movies' } });
  }
}
