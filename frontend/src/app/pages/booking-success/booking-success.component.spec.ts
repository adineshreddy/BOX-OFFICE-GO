import { BehaviorSubject, of, throwError } from 'rxjs';
import { vi } from 'vitest';
import { BookingSuccessComponent } from './booking-success.component';

function createRouteStub() {
  const paramMap$ = new BehaviorSubject({
    get: (k: string) => {
      if (k === 'movieId') return 'mov_001';
      if (k === 'bookingId') return 'bok_001';
      return null;
    }
  });
  const queryParamMap$ = new BehaviorSubject({
    get: (k: string) => {
      if (k === 'seats') return 'D5,D6';
      if (k === 'total') return '24';
      return null;
    }
  });
  return {
    paramMap: paramMap$.asObservable(),
    queryParamMap: queryParamMap$.asObservable()
  } as any;
}

describe('BookingSuccessComponent', () => {
  it('parses booking route/query data on init', () => {
    const component = new BookingSuccessComponent(
      createRouteStub(),
      { downloadTicket: vi.fn().mockReturnValue(of(new Blob(['pdf'], { type: 'application/pdf' }))) } as any
    );

    expect(component.movieId()).toBe('mov_001');
    expect(component.bookingId()).toBe('bok_001');
    expect(component.seats()).toEqual(['D5', 'D6']);
    expect(component.total()).toBe(24);
  });

  it('sets error when ticket download fails', () => {
    const component = new BookingSuccessComponent(
      createRouteStub(),
      { downloadTicket: vi.fn().mockReturnValue(throwError(() => new Error('fail'))) } as any
    );

    component.downloadTicket();

    expect(component.downloading()).toBe(false);
    expect(component.downloadError()).toBe('Could not download ticket right now. Please try again.');
  });
});
