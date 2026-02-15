import { Injectable } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { environment } from '../../environments/environment';

@Injectable({ providedIn: 'root' })
export class MovieService {
  constructor(private http: HttpClient) {}

  getMovies(query: any) {
    let params = new HttpParams();
    Object.keys(query).forEach(key => {
      if (query[key]) params = params.set(key, query[key]);
    });
    return this.http.get(`${environment.apiUrl}/movies`, { params });
  }

  getMovie(id: string) {
    return this.http.get(`${environment.apiUrl}/movies/${id}`);
  }

  getShows(movieId: string) {
    return this.http.get(`${environment.apiUrl}/movies/${movieId}/shows`);
  }
}
