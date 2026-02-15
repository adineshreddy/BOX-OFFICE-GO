#!/bin/bash

mkdir -p src/app/core/interceptors
mkdir -p src/app/core/guards
mkdir -p src/app/models
mkdir -p src/app/services
mkdir -p src/app/pages/auth
mkdir -p src/app/pages/movies
mkdir -p src/app/pages/booking
mkdir -p src/app/pages/admin
mkdir -p src/app/shared

# Environment
cat > src/environments/environment.ts <<EOF
export const environment = {
  production: false,
  apiUrl: 'http://localhost:8080/api'
};
EOF

# Auth Service
cat > src/app/services/auth.service.ts <<EOF
import { Injectable, signal } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { environment } from '../../environments/environment';
import { Router } from '@angular/router';

@Injectable({ providedIn: 'root' })
export class AuthService {
  private token = signal<string | null>(localStorage.getItem('token'));
  private role = signal<string | null>(localStorage.getItem('role'));

  constructor(private http: HttpClient, private router: Router) {}

  login(data: any) {
    return this.http.post<any>(\`\${environment.apiUrl}/auth/login\`, data);
  }

  signup(data: any) {
    return this.http.post<any>(\`\${environment.apiUrl}/auth/signup\`, data);
  }

  setSession(res: any) {
    localStorage.setItem('token', res.token);
    localStorage.setItem('role', res.role);
    this.token.set(res.token);
    this.role.set(res.role);
  }

  logout() {
    localStorage.clear();
    this.token.set(null);
    this.role.set(null);
    this.router.navigate(['/login']);
  }

  isLoggedIn() {
    return !!this.token();
  }

  isAdmin() {
    return this.role() === 'ADMIN';
  }

  getToken() {
    return this.token();
  }
}
EOF

# Auth Interceptor
cat > src/app/core/interceptors/auth.interceptor.ts <<EOF
import { HttpInterceptorFn } from '@angular/common/http';
import { inject } from '@angular/core';
import { AuthService } from '../../services/auth.service';

export const authInterceptor: HttpInterceptorFn = (req, next) => {
  const auth = inject(AuthService);
  const token = auth.getToken();

  if (token) {
    req = req.clone({
      setHeaders: { Authorization: \`Bearer \${token}\` }
    });
  }

  return next(req);
};
EOF

# Movie Service
cat > src/app/services/movie.service.ts <<EOF
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
    return this.http.get(\`\${environment.apiUrl}/movies\`, { params });
  }

  getMovie(id: string) {
    return this.http.get(\`\${environment.apiUrl}/movies/\${id}\`);
  }

  getShows(movieId: string) {
    return this.http.get(\`\${environment.apiUrl}/movies/\${movieId}/shows\`);
  }
}
EOF

# Booking Service
cat > src/app/services/booking.service.ts <<EOF
import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { environment } from '../../environments/environment';

@Injectable({ providedIn: 'root' })
export class BookingService {
  constructor(private http: HttpClient) {}

  getSeats(showId: string) {
    return this.http.get(\`\${environment.apiUrl}/shows/\${showId}/seats\`);
  }

  createBooking(data: any) {
    return this.http.post(\`\${environment.apiUrl}/bookings\`, data);
  }

  getMyBookings() {
    return this.http.get(\`\${environment.apiUrl}/bookings/me\`);
  }

  cancelBooking(id: string) {
    return this.http.post(\`\${environment.apiUrl}/bookings/\${id}/cancel\`, {});
  }

  downloadTicket(id: string) {
    return this.http.get(
      \`\${environment.apiUrl}/bookings/\${id}/ticket\`,
      { responseType: 'blob' }
    );
  }
}
EOF

echo "Frontend setup complete 🚀"
