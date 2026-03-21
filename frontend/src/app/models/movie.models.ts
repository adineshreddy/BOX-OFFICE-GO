export interface ApiMovie {
  id: string;
  title: string;
  description: string;
  genre: string;
  language: string;
  durationMinutes: number;
  releaseDate: string;
  rating: number;
  posterUrl?: string | null;
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface MovieListResponse {
  movies: ApiMovie[];
}

export interface ShowtimeItem {
  showtimeId: string;
  screenName: string;
  startTime: string;
  endTime: string;
  language: string;
  format: string;
  basePrice: number;
}

export interface TheaterSchedule {
  theaterId: string;
  theaterName: string;
  city: string;
  addressLine1: string;
  timezone: string;
  showtimes: ShowtimeItem[];
}

export interface MovieTheaterListResponse {
  movieId: string;
  movieTitle: string;
  durationMinutes: number;
  theaters: TheaterSchedule[];
}
