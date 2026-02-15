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
    return this.http.post<any>(`${environment.apiUrl}/auth/login`, data);
  }

  signup(data: any) {
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
