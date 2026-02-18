import { Component, OnInit, OnDestroy, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Router, RouterLink } from '@angular/router';

interface Movie {
  id: string;
  title: string;
  genre: string;
  posterUrl: string;
  likesK: number;
  runtimeMinutes: number;
  certificate: 'A' | 'U/A' | 'U';
  rating: number; // /10
  trailerUrl: string;
}

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

  // 10 English movies with real posters + details
  allMovies: Movie[] = [
    {
      id: '1',
      title: 'The Dark Knight',
      genre: 'Action/Thriller',
      posterUrl: 'https://image.tmdb.org/t/p/w500/qJ2tW6WMUDux911r6m7haRef0WH.jpg',
      likesK: 245,
      runtimeMinutes: 152,
      certificate: 'U/A',
      rating: 9.0,
      trailerUrl: 'https://www.youtube.com/watch?v=EXeTwQWrcwY'
    },
    {
      id: '2',
      title: 'Inception',
      genre: 'Sci-Fi/Thriller',
      posterUrl: 'https://image.tmdb.org/t/p/w500/oYuLEt3zVCKq57qu2F8dT7NIa6f.jpg',
      likesK: 189,
      runtimeMinutes: 148,
      certificate: 'U/A',
      rating: 8.8,
      trailerUrl: 'https://www.youtube.com/watch?v=YoHD9XEInc0'
    },
    {
      id: '3',
      title: 'Interstellar',
      genre: 'Sci-Fi/Drama',
      posterUrl: 'https://image.tmdb.org/t/p/w500/gEU2QniE6E77NI6lCU6MxlNBvIx.jpg',
      likesK: 312,
      runtimeMinutes: 169,
      certificate: 'U/A',
      rating: 8.6,
      trailerUrl: 'https://www.youtube.com/watch?v=zSWdZVtXT7E'
    },
    {
      id: '4',
      title: 'The Matrix',
      genre: 'Action/Sci-Fi',
      posterUrl: 'https://image.tmdb.org/t/p/w500/f89U3ADr1oiB1s9GkdPOEpXUk5H.jpg',
      likesK: 278,
      runtimeMinutes: 136,
      certificate: 'A',
      rating: 8.7,
      trailerUrl: 'https://www.youtube.com/watch?v=vKQi3bBA1y8'
    },
    {
      id: '5',
      title: 'Pulp Fiction',
      genre: 'Crime/Drama',
      posterUrl: 'https://image.tmdb.org/t/p/w500/d5iIlFn5s0ImszYzBPb8JPIfbXD.jpg',
      likesK: 156,
      runtimeMinutes: 154,
      certificate: 'A',
      rating: 8.9,
      trailerUrl: 'https://www.youtube.com/watch?v=s7EdQ4FqbhY'
    },
    {
      id: '6',
      title: 'Fight Club',
      genre: 'Drama/Thriller',
      posterUrl: 'https://image.tmdb.org/t/p/w500/pB8BM7pdSp6B6Ih7QZ4DrQ3PmJK.jpg',
      likesK: 223,
      runtimeMinutes: 139,
      certificate: 'A',
      rating: 8.8,
      trailerUrl: 'https://www.youtube.com/watch?v=SUXWAEX2jlg'
    },
    {
      id: '7',
      title: 'The Shawshank Redemption',
      genre: 'Drama',
      posterUrl: 'https://image.tmdb.org/t/p/w500/q6y0Go1tsGEsmtFryDOJo3dEmqu.jpg',
      likesK: 445,
      runtimeMinutes: 142,
      certificate: 'A',
      rating: 9.3,
      trailerUrl: 'https://www.youtube.com/watch?v=6hB3S9bIaco'
    },
    {
      id: '8',
      title: 'Forrest Gump',
      genre: 'Drama/Romance',
      posterUrl: 'https://image.tmdb.org/t/p/w500/arw2vcBveWOVZr6pxd9XTd1TdQa.jpg',
      likesK: 198,
      runtimeMinutes: 142,
      certificate: 'U/A',
      rating: 8.8,
      trailerUrl: 'https://www.youtube.com/watch?v=bLvqoHBptjg'
    },
    {
      id: '9',
      title: 'The Godfather',
      genre: 'Crime/Drama',
      posterUrl: 'https://image.tmdb.org/t/p/w500/3bhkrj58Vtu7enYsRolD1fZdja1.jpg',
      likesK: 367,
      runtimeMinutes: 175,
      certificate: 'A',
      rating: 9.2,
      trailerUrl: 'https://www.youtube.com/watch?v=sY1S34973zA'
    },
    {
      id: '10',
      title: 'Gladiator',
      genre: 'Action/Drama',
      posterUrl: 'https://image.tmdb.org/t/p/w500/ty8TGRuvJLPUmAR1H1nRIsgwvim.jpg',
      likesK: 289,
      runtimeMinutes: 155,
      certificate: 'A',
      rating: 8.5,
      trailerUrl: 'https://www.youtube.com/watch?v=owK1qxDselE'
    }
  ];

  // Computed: movies visible on current page
  visibleMovies = computed(() => {
    const start = this.currentMoviePage() * this.moviesPerPage;
    const end = start + this.moviesPerPage;
    return this.allMovies.slice(start, end);
  });

  // Computed: check if we can go to next page
  canGoNext = computed(() => {
    return (this.currentMoviePage() + 1) * this.moviesPerPage < this.allMovies.length;
  });

  // Computed: check if we can go to previous page
  canGoPrev = computed(() => {
    return this.currentMoviePage() > 0;
  });

  // Hero banners data - promotional banners
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

  constructor(private router: Router) {}

  ngOnInit() {
    this.startAutoRotate();
  }

  ngOnDestroy() {
    this.stopAutoRotate();
  }

  startAutoRotate() {
    this.stopAutoRotate(); // Clear any existing interval
    this.autoRotateInterval = window.setInterval(() => {
      this.nextHero();
    }, 3000); // Change every 3 seconds
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
    // Restart auto-rotate after manual navigation
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
