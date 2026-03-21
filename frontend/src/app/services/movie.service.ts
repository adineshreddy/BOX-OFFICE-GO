import { Injectable } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '../../environments/environment';
import {
  ApiMovie,
  MovieListResponse,
  MovieTheaterListResponse
} from '../models/movie.models';

@Injectable({ providedIn: 'root' })
export class MovieService {
  constructor(private http: HttpClient) {}

  listMovies(query: Record<string, string | undefined> = {}): Observable<MovieListResponse> {
    let params = new HttpParams();
    Object.keys(query).forEach(key => {
      const v = query[key];
      if (v != null && v !== '') {
        params = params.set(key, v);
      }
    });
    return this.http.get<MovieListResponse>(`${environment.apiUrl}/movies`, { params });
  }

  getMovie(id: string): Observable<ApiMovie> {
    return this.http.get<ApiMovie>(`${environment.apiUrl}/movies/${encodeURIComponent(id)}`);
  }

  getTheatersForMovie(movieId: string, date?: string): Observable<MovieTheaterListResponse> {
    let params = new HttpParams();
    if (date) {
      params = params.set('date', date);
    }
    return this.http.get<MovieTheaterListResponse>(
      `${environment.apiUrl}/movies/${encodeURIComponent(movieId)}/theaters`,
      { params }
    );
  }
}
