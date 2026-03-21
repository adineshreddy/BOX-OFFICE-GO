import { Injectable, signal, computed } from '@angular/core';
import { HttpClient, HttpErrorResponse } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '../../environments/environment';
import { Router } from '@angular/router';
import {
  AuthUser,
  LoginRequest,
  LoginResponse,
  SignupRequest,
  SignupResponse
} from '../models/auth.models';

const USER_STORAGE_KEY = 'auth_user';

export function formatAuthHttpError(err: unknown): string {
  if (err instanceof HttpErrorResponse && err.error && typeof err.error === 'object') {
    const e = err.error as { message?: string; fields?: Record<string, string> };
    const fieldText = e.fields ? Object.values(e.fields).join(' ') : '';
    if (e.message || fieldText) {
      return [e.message, fieldText].filter(Boolean).join(' ');
    }
  }
  if (err instanceof Error) {
    return err.message;
  }
  return 'Something went wrong';
}

@Injectable({ providedIn: 'root' })
export class AuthService {
  private user = signal<AuthUser | null>(this.readStoredUser());

  readonly isLoggedInSignal = computed(() => !!this.user());

  constructor(private http: HttpClient, private router: Router) {}

  private readStoredUser(): AuthUser | null {
    try {
      const raw = localStorage.getItem(USER_STORAGE_KEY);
      if (!raw) {
        return null;
      }
      return JSON.parse(raw) as AuthUser;
    } catch {
      return null;
    }
  }

  login(data: LoginRequest): Observable<LoginResponse> {
    return this.http.post<LoginResponse>(`${environment.apiUrl}/auth/login`, data);
  }

  signup(data: SignupRequest): Observable<SignupResponse> {
    return this.http.post<SignupResponse>(`${environment.apiUrl}/auth/signup`, data);
  }

  /** Persists session from API responses (backend does not issue JWT). */
  setSessionFromLogin(res: LoginResponse) {
    localStorage.setItem(USER_STORAGE_KEY, JSON.stringify(res.user));
    this.user.set(res.user);
  }

  logout() {
    localStorage.removeItem(USER_STORAGE_KEY);
    localStorage.removeItem('token');
    localStorage.removeItem('role');
    this.user.set(null);
    this.router.navigate(['/login']);
  }

  isLoggedIn(): boolean {
    return this.isLoggedInSignal();
  }

  getUser(): AuthUser | null {
    return this.user();
  }

  /** No JWT from API; interceptor skips Authorization when null. */
  getToken(): string | null {
    return null;
  }

  isAdmin(): boolean {
    return this.user()?.isAdmin === true;
  }
}
