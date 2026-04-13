import { BehaviorSubject, of, throwError } from 'rxjs';
import { vi } from 'vitest';
import { PaymentComponent } from './payment.component';

function createRouteStub() {
  const paramMap$ = new BehaviorSubject({
    get: (k: string) => (k === 'movieId' ? 'mov_001' : null)
  });
  const queryParamMap$ = new BehaviorSubject({
    get: (k: string) => {
      if (k === 'theaterId') return 'th_001';
      if (k === 'showTime') return '2026-04-13T18:30:00Z';
      if (k === 'holdId') return 'hold_001';
      if (k === 'holdExpiresAt') return new Date(Date.now() + 5 * 60 * 1000).toISOString();
      return null;
    }
  });
  return {
    paramMap: paramMap$.asObservable(),
    queryParamMap: queryParamMap$.asObservable()
  } as any;
}

describe('PaymentComponent', () => {
  it('submits checkout and navigates to booking-success on success', () => {
    const checkoutBookingHold = vi.fn().mockReturnValue(
      of({
        booking: {
          bookingId: 'bok_001',
          holdId: 'hold_001',
          totalAmount: 24,
          seatNumbers: ['D5', 'D6'],
          status: 'CONFIRMED'
        }
      })
    );

    const component = new PaymentComponent(
      createRouteStub(),
      { navigate: vi.fn().mockResolvedValue(true) } as any,
      { getUser: vi.fn().mockReturnValue({ id: 'u1' }) } as any,
      { checkoutBookingHold, releaseBookingHold: vi.fn() } as any
    );

    component.ngOnInit();
    component.submitPayment();

    expect(checkoutBookingHold).toHaveBeenCalled();
    expect(component.processing()).toBe(false);
  });

  it('shows server message on checkout failure', () => {
    const checkoutBookingHold = vi.fn().mockReturnValue(
      throwError(() => ({
        status: 500,
        error: { message: 'booking operation failed' }
      }))
    );

    const component = new PaymentComponent(
      createRouteStub(),
      { navigate: vi.fn().mockResolvedValue(true) } as any,
      { getUser: vi.fn().mockReturnValue({ id: 'u1' }) } as any,
      { checkoutBookingHold, releaseBookingHold: vi.fn() } as any
    );

    component.ngOnInit();
    component.submitPayment();

    expect(component.error()).toBe('booking operation failed');
  });
});
