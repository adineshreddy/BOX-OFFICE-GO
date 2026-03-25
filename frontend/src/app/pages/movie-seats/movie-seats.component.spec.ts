import { vi } from 'vitest';
import type { SeatItem } from '../../models/movie.models';
import { MovieSeatsComponent } from './movie-seats.component';

describe('MovieSeatsComponent', () => {
  it('toggleSeat() enforces selection limit and toggles seat', () => {
    const component = new MovieSeatsComponent(
      {} as any,
      {} as any,
      {} as any,
      {} as any,
      { navigate: vi.fn() } as any
    );

    const seat1 = { seatNumber: 'A1', isAvailable: true, isHeld: false } as any as SeatItem;
    const seat2 = { seatNumber: 'A2', isAvailable: true, isHeld: false } as any as SeatItem;

    component.ticketCount.set(1);
    component.selectedSeats.set(['A1']);
    component.selectionError.set(null);

    component.toggleSeat(seat2);

    expect(component.selectedSeats()).toEqual(['A1']);
    expect(component.selectionError()).toBe('You can select only 1 seat(s).');

    component.toggleSeat(seat1);

    expect(component.selectedSeats()).toEqual([]);
    expect(component.selectionError()).toBeNull();
  });
});

