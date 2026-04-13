import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { environment } from '../../environments/environment';

@Injectable({ providedIn: 'root' })
export class BookingService {
  constructor(private http: HttpClient) {}

  createBookingHold(data: { userId: string; showtimeId: string; seatNumbers: string[] }) {
    return this.http.post<{ message: string; hold: { holdId: string; holdExpiresAt: string; totalAmount: number } }>(
      `${environment.apiUrl}/bookings/holds`,
      data
    );
  }

  checkoutBookingHold(data: {
    holdId: string;
    userId: string;
    paymentMethod: string;
    idempotencyKey: string;
  }) {
    return this.http.post<{
      message: string;
      booking: { bookingId: string; holdId: string; totalAmount: number; seatNumbers: string[]; status: string };
    }>(`${environment.apiUrl}/bookings/checkout`, data);
  }

  releaseBookingHold(holdId: string) {
    return this.http.delete<{ message: string }>(`${environment.apiUrl}/bookings/holds/${holdId}`);
  }

  getSeats(showId: string) {
    return this.http.get(`${environment.apiUrl}/shows/${showId}/seats`);
  }

  createBooking(data: any) {
    return this.http.post(`${environment.apiUrl}/bookings`, data);
  }

  getMyBookings() {
    return this.http.get(`${environment.apiUrl}/bookings/me`);
  }

  cancelBooking(id: string) {
    return this.http.post(`${environment.apiUrl}/bookings/${id}/cancel`, {});
  }

  downloadTicket(id: string) {
    return this.http.get(
      `${environment.apiUrl}/bookings/${id}/ticket`,
      { responseType: 'blob' }
    );
  }
}
