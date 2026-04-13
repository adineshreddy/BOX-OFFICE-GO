/** Matches login response user payload from the API. */
export interface AuthUser {
  id: string;
  name: string;
  phone: string;
  email: string;
  isAdmin?: boolean;
  isActive?: boolean;
  isVerified?: boolean;
  lastLoginAt?: string;
  createdAt?: string;
  updatedAt?: string;
}

export interface LoginResponse {
  message: string;
  accessToken: string;
  tokenType: string;
  expiresAt: string;
  user: AuthUser;
}

export interface SignupResponse {
  message: string;
  user: {
    id: string;
    name: string;
    phone: string;
    email: string;
    createdAt: string;
  };
}

export interface SignupRequest {
  name: string;
  phone: string;
  email: string;
  password: string;
  confirmPassword: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}
