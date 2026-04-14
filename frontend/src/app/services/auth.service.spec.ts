import { HttpErrorResponse } from '@angular/common/http';
import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';
import { environment } from '../../environments/environment';
import { Router } from '@angular/router';
import { vi } from 'vitest';
import { AuthService, formatAuthHttpError } from './auth.service';

function ensureLocalStorage() {
  if (typeof globalThis.localStorage !== 'undefined') {
    return;
  }

  const store = new Map<string, string>();
  const localStorageMock = {
    getItem: (key: string) => (store.has(key) ? store.get(key)! : null),
    setItem: (key: string, value: string) => {
      store.set(key, value);
    },
    removeItem: (key: string) => {
      store.delete(key);
    },
    clear: () => {
      store.clear();
    }
  };

  Object.defineProperty(globalThis, 'localStorage', { value: localStorageMock });
}

describe('formatAuthHttpError', () => {
  beforeEach(() => {
    ensureLocalStorage();
  });

  it('formats HttpErrorResponse and Error instances', () => {
    const httpErr = new HttpErrorResponse({
      status: 400,
      error: {
        message: 'Invalid credentials',
        fields: { email: 'Email is not valid' }
      }
    });

    expect(formatAuthHttpError(httpErr)).toBe('Invalid credentials Email is not valid');

    const err = new Error('Boom');
    expect(formatAuthHttpError(err)).toBe('Boom');

    expect(formatAuthHttpError('nope')).toBe('Something went wrong');
  });
});

describe('AuthService', () => {
  const USER_STORAGE_KEY = 'auth_user';

  let service: AuthService;
  let routerNavigate: ReturnType<typeof vi.fn>;
  let httpMock: HttpTestingController;

  beforeEach(async () => {
    ensureLocalStorage();
    globalThis.localStorage.clear();

    routerNavigate = vi.fn();

    await TestBed.configureTestingModule({
      imports: [HttpClientTestingModule],
      providers: [
        { provide: Router, useValue: { navigate: routerNavigate } }
      ]
    }).compileComponents();

    service = TestBed.inject(AuthService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('setSessionFromLogin persists user and updates the signal', () => {
    const res = {
      message: 'ok',
      accessToken: 'jwt-token',
      tokenType: 'Bearer',
      expiresAt: new Date().toISOString(),
      user: {
        id: 'u1',
        name: 'Test User',
        phone: '1234567890',
        email: 'test@example.com',
        isAdmin: false
      }
    };

    service.setSessionFromLogin(res);

    expect(globalThis.localStorage.getItem(USER_STORAGE_KEY)).toBe(JSON.stringify(res.user));
    expect(globalThis.localStorage.getItem('token')).toBe('jwt-token');
    expect(service.getUser()).toEqual(res.user);
    expect(service.getToken()).toBe('jwt-token');
    expect(service.isLoggedIn()).toBe(true);
  });

  it('logout clears storage and navigates to /login', () => {
    globalThis.localStorage.setItem(
      USER_STORAGE_KEY,
      JSON.stringify({
        id: 'u1',
        name: 'Test User',
        phone: '1234567890',
        email: 'test@example.com'
      })
    );
    globalThis.localStorage.setItem('token', 'abc');
    globalThis.localStorage.setItem('role', 'admin');

    service.logout();

    const req = httpMock.expectOne(`${environment.apiUrl}/auth/logout`);
    expect(req.request.method).toBe('POST');
    req.flush({ message: 'logout successful' });

    expect(globalThis.localStorage.getItem(USER_STORAGE_KEY)).toBeNull();
    expect(globalThis.localStorage.getItem('token')).toBeNull();
    expect(globalThis.localStorage.getItem('role')).toBeNull();
    expect(service.getUser()).toBeNull();
    expect(routerNavigate).toHaveBeenCalledWith(['/login']);
  });
});

