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
const TOKEN_STORAGE_KEY = 'token';

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
  private token = signal<string | null>(this.readStoredToken());

  readonly isLoggedInSignal = computed(() => !!this.user());

  constructor(private http: HttpClient, private router: Router) {}

  private readStoredToken(): string | null {
    try {
      return localStorage.getItem(TOKEN_STORAGE_KEY);
    } catch {
      return null;
    }
  }

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

  /** Persists session and bearer token from login response. */
  setSessionFromLogin(res: LoginResponse) {
    localStorage.setItem(USER_STORAGE_KEY, JSON.stringify(res.user));
    localStorage.setItem(TOKEN_STORAGE_KEY, res.accessToken);
    this.user.set(res.user);
    this.token.set(res.accessToken);
  }

  logout() {
    localStorage.removeItem(USER_STORAGE_KEY);
    localStorage.removeItem(TOKEN_STORAGE_KEY);
    localStorage.removeItem('role');
    this.user.set(null);
    this.token.set(null);
    this.router.navigate(['/login']);
  }

  isLoggedIn(): boolean {
    return this.isLoggedInSignal();
  }

  getUser(): AuthUser | null {
    return this.user();
  }

  getToken(): string | null {
    return this.token();
  }

  isAdmin(): boolean {
    return this.user()?.isAdmin === true;
  }
}
