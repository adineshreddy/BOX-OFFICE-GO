import { Injectable, signal, computed } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, of, throwError } from 'rxjs';
import { delay } from 'rxjs/operators';
import { environment } from '../../environments/environment';
import { Router } from '@angular/router';

/** Mock user store for signup (in-memory). */
const MOCK_USERS: Array<{ email: string; password: string; name: string }> = [
  { email: 'demo@example.com', password: 'demo123', name: 'Demo User' }
];

function mockLogin(data: { email: string; password: string }): Observable<any> {
  const user = MOCK_USERS.find(
    (u) => u.email === data.email && u.password === data.password
  );
  if (!user) {
    return throwError(() => ({
      error: { message: 'Invalid email or password' },
      message: 'Invalid email or password'
    })).pipe(delay(400));
  }
  return of({
    token: 'mock-jwt-' + btoa(data.email).slice(0, 20),
    role: 'USER'
  }).pipe(delay(400));
}

function mockSignup(data: {
  email: string;
  password: string;
  name: string;
}): Observable<any> {
  if (MOCK_USERS.some((u) => u.email === data.email)) {
    return throwError(() => ({
      error: { message: 'Email already registered' },
      message: 'Email already registered'
    })).pipe(delay(400));
  }
  MOCK_USERS.push({
    email: data.email,
    password: data.password,
    name: data.name
  });
  return of({
    token: 'mock-jwt-' + btoa(data.email).slice(0, 20),
    role: 'USER'
  }).pipe(delay(400));
}

@Injectable({ providedIn: 'root' })
export class AuthService {
  private token = signal<string | null>(localStorage.getItem('token'));
  private role = signal<string | null>(localStorage.getItem('role'));
  /** Use in templates so the header updates when login/signup/logout changes. */
  readonly isLoggedInSignal = computed(() => !!this.token());

  constructor(private http: HttpClient, private router: Router) {}

  login(data: any): Observable<any> {
    if (environment.useMockAuth) {
      return mockLogin(data);
    }
    return this.http.post<any>(`${environment.apiUrl}/auth/login`, data);
  }

  signup(data: any): Observable<any> {
    if (environment.useMockAuth) {
      return mockSignup(data);
    }
    return this.http.post<any>(`${environment.apiUrl}/auth/signup`, data);
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

  isLoggedIn(): boolean {
    return this.isLoggedInSignal();
  }

  isAdmin() {
    return this.role() === 'ADMIN';
  }

  getToken() {
    return this.token();
  }
}
