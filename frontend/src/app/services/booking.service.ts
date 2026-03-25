import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { environment } from '../../environments/environment';

@Injectable({ providedIn: 'root' })
export class BookingService {
  constructor(private http: HttpClient) {}

  createBookingHold(data: { userId: string; showtimeId: string; seatNumbers: string[] }) {
    return this.http.post(`${environment.apiUrl}/bookings/holds`, data);
  }

  checkoutBookingHold(data: { holdId: string; userId: string }) {
    return this.http.post(`${environment.apiUrl}/bookings/checkout`, data);
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
