import { Routes } from '@angular/router';

export const routes: Routes = [
  {
    path: '',
    loadComponent: () =>
      import('./pages/home/home.component').then(m => m.HomeComponent)
  },
  {
    path: 'movies/:movieId/seats',
    loadComponent: () =>
      import('./pages/movie-seats/movie-seats.component').then(m => m.MovieSeatsComponent)
  },
  {
    path: 'movies/:movieId/payment',
    loadComponent: () =>
      import('./pages/payment/payment.component').then(m => m.PaymentComponent)
  },
  {
    path: 'movies/:movieId/booking-success/:bookingId',
    loadComponent: () =>
      import('./pages/booking-success/booking-success.component').then(m => m.BookingSuccessComponent)
  },
  {
    path: 'movies/:movieId',
    loadComponent: () =>
      import('./pages/movie-detail/movie-detail.component').then(m => m.MovieDetailComponent)
  },
  {
    path: 'login',
    loadComponent: () =>
      import('./pages/auth/login/login.component').then(m => m.LoginComponent)
  },
  {
    path: 'signup',
    loadComponent: () =>
      import('./pages/auth/signup/signup.component').then(m => m.SignupComponent)
  }
];
