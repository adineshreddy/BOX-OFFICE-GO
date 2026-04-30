import { Component, inject, signal, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterLink } from '@angular/router';
import { MovieService } from '../../services/movie.service';
import { ApiMovie } from '../../models/movie.models';

@Component({
  selector: 'app-browse-movies',
  standalone: true,
  imports: [CommonModule, RouterLink],
  templateUrl: './browse-movies.component.html',
  styleUrl: './browse-movies.component.scss'
})
export class BrowseMoviesComponent implements OnInit {
  private readonly movieService = inject(MovieService);

  movies = signal<ApiMovie[]>([]);
  loadingMovies = signal(true);
  moviesError = signal<string | null>(null);

  readonly posterFallback =
    'https://placehold.co/300x450/312e81/f472b6?text=Poster';

  ngOnInit() {
    this.movieService.listMovies({ limit: '100', sort: 'title' }).subscribe({
      next: res => {
        this.movies.set(res.movies);
        this.loadingMovies.set(false);
      },
      error: () => {
        this.moviesError.set('Could not load movies. Start the API (see README) and refresh.');
        this.loadingMovies.set(false);
      }
    });
  }

  posterUrl(movie: ApiMovie): string {
    return movie.posterUrl || this.posterFallback;
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
