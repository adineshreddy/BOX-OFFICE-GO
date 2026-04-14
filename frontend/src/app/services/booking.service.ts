import { Injectable } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { environment } from '../../environments/environment';

/** Matches GET /bookings list items from the API. */
export interface UserBooking {
  bookingId: string;
  holdId: string;
  userId: string;
  showtimeId: string;
  seatNumbers: string[];
  status: string;
  totalAmount: number;
  confirmedAt: string;
  movieTitle: string;
  theaterName: string;
  city: string;
  screenName: string;
  showTime: string;
  language: string;
  format: string;
}

@Injectable({ providedIn: 'root' })
export class BookingService {
  constructor(private http: HttpClient) {}

  createBookingHold(data: { showtimeId: string; seatNumbers: string[] }) {
    return this.http.post<{ message: string; hold: { holdId: string; holdExpiresAt: string; totalAmount: number } }>(
      `${environment.apiUrl}/bookings/holds`,
      data
    );
  }

  checkoutBookingHold(data: { holdId: string; paymentMethod: string; idempotencyKey: string }) {
    return this.http.post<{
      message: string;
      booking: { bookingId: string; holdId: string; totalAmount: number; seatNumbers: string[]; status: string };
    }>(`${environment.apiUrl}/bookings/checkout`, data);
  }

  releaseBookingHold(holdId: string) {
    return this.http.delete<{ message: string }>(`${environment.apiUrl}/bookings/holds/${holdId}`);
  }

  getMyBookings() {
    return this.http.get<{ bookings: UserBooking[] }>(`${environment.apiUrl}/bookings`);
  }

  cancelBooking(bookingId: string) {
    const params = new HttpParams().set('bookingId', bookingId);
    return this.http.delete<{ message: string }>(`${environment.apiUrl}/bookings`, { params });
  }

  downloadTicket(id: string) {
    return this.http.get(`${environment.apiUrl}/bookings/${id}/ticket`, { responseType: 'blob' });
  }
}
