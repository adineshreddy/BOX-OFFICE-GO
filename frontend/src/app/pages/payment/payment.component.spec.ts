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
    component.cardNumber.set('4111111111111111');
    component.cardExpiry.set('12/30');
    component.cardCvv.set('123');
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
    component.cardNumber.set('4111111111111111');
    component.cardExpiry.set('12/30');
    component.cardCvv.set('123');
    component.submitPayment();

    expect(component.error()).toBe('booking operation failed');
  });

  it('does not call checkout when card fields are empty', () => {
    const checkoutBookingHold = vi.fn().mockReturnValue(of({ booking: null }));

    const component = new PaymentComponent(
      createRouteStub(),
      { navigate: vi.fn().mockResolvedValue(true) } as any,
      { getUser: vi.fn().mockReturnValue({ id: 'u1' }) } as any,
      { checkoutBookingHold, releaseBookingHold: vi.fn() } as any
    );

    component.ngOnInit();
    component.submitPayment();

    expect(checkoutBookingHold).not.toHaveBeenCalled();
    expect(component.error()).toBe('Card number, expiry and CVV are required.');
  });

  it('submits net banking details when selected', () => {
    const checkoutBookingHold = vi.fn().mockReturnValue(of({ booking: { bookingId: 'bok_002' } }));

    const component = new PaymentComponent(
      createRouteStub(),
      { navigate: vi.fn().mockResolvedValue(true) } as any,
      { getUser: vi.fn().mockReturnValue({ id: 'u1' }) } as any,
      { checkoutBookingHold, releaseBookingHold: vi.fn() } as any
    );

    component.ngOnInit();
    component.paymentMethod.set('netbanking');
    component.bankName.set('Test Bank');
    component.accountNumber.set('123456789');
    component.routingNumber.set('021000021');
    component.submitPayment();

    expect(checkoutBookingHold).toHaveBeenCalledWith(
      expect.objectContaining({
        paymentMethod: 'netbanking',
        cardNumber: 'Test Bank:123456789',
        cardExpiry: '021000021',
        cardCvv: 'NA'
      })
    );
  });

  it('shows error for invalid card number length', () => {
    const checkoutBookingHold = vi.fn().mockReturnValue(of({ booking: null }));

    const component = new PaymentComponent(
      createRouteStub(),
      { navigate: vi.fn().mockResolvedValue(true) } as any,
      { getUser: vi.fn().mockReturnValue({ id: 'u1' }) } as any,
      { checkoutBookingHold, releaseBookingHold: vi.fn() } as any
    );

    component.ngOnInit();
    component.cardNumber.set('1234');
    component.cardExpiry.set('12/30');
    component.cardCvv.set('123');
    component.submitPayment();

    expect(checkoutBookingHold).not.toHaveBeenCalled();
    expect(component.error()).toBe('Card number must be 13 to 19 digits.');
  });
});
