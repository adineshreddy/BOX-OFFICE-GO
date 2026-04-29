/** Stub show APIs so seats page works after redirect (real DB may not have this showtime). */
function stubSeatPageApis(apiBase: string) {
  cy.intercept('GET', `${apiBase}/shows/details*`, {
    statusCode: 200,
    body: {
      showtimeId: 'st_stub',
      movieId: 'mov_001',
      movieTitle: 'Test Movie',
      theaterId: 'th_001',
      theaterName: 'Test Theater',
      city: 'Test City',
      addressLine1: '1 Main St',
      screenName: 'Screen 1',
      startTime: '2026-04-13T18:30:00Z',
      language: 'English',
      format: '2D',
      basePrice: 12,
      durationMinutes: 120,
      availableSeats: 4,
      totalSeats: 4,
      unavailableSeats: 0
    }
  }).as('showDetails');

  cy.intercept('GET', `${apiBase}/shows/seat-map*`, {
    statusCode: 200,
    body: {
      showtimeId: 'st_stub',
      movieTitle: 'Test Movie',
      theaterName: 'Test Theater',
      screenName: 'Screen 1',
      showTime: '2026-04-13T18:30:00Z',
      totalSeats: 4,
      availableSeats: 4,
      unavailableSeats: 0,
      rows: [
        {
          rowLabel: 'A',
          seats: [
            {
              seatNumber: 'A1',
              rowLabel: 'A',
              seatIndex: 1,
              seatType: 'REGULAR',
              priceFactor: 1,
              isAvailable: true,
              isHeld: false
            },
            {
              seatNumber: 'A2',
              rowLabel: 'A',
              seatIndex: 2,
              seatType: 'REGULAR',
              priceFactor: 1,
              isAvailable: true,
              isHeld: false
            }
          ]
        },
        {
          rowLabel: 'B',
          seats: [
            {
              seatNumber: 'B1',
              rowLabel: 'B',
              seatIndex: 3,
              seatType: 'REGULAR',
              priceFactor: 1,
              isAvailable: true,
              isHeld: false
            },
            {
              seatNumber: 'B2',
              rowLabel: 'B',
              seatIndex: 4,
              seatType: 'REGULAR',
              priceFactor: 1,
              isAvailable: true,
              isHeld: false
            }
          ]
        }
      ]
    }
  }).as('seatMap');
}

describe('Payment and booking success flow', () => {
  it('completes payment and navigates to booking success', () => {
    const apiBase = 'http://localhost:8080/api/v1';

    cy.intercept('POST', `${apiBase}/bookings/checkout`, {
      statusCode: 200,
      body: {
        message: 'checkout successful',
        booking: {
          bookingId: 'bok_001',
          holdId: 'hold_001',
          seatNumbers: ['D5', 'D6'],
          status: 'CONFIRMED',
          totalAmount: 24
        }
      }
    }).as('checkout');

    cy.intercept('GET', `${apiBase}/bookings/bok_001/ticket`, {
      statusCode: 200,
      headers: { 'content-type': 'application/pdf' },
      body: 'PDF_CONTENT'
    }).as('ticket');

    cy.visit('/movies/mov_001/payment?theaterId=th_001&showTime=2026-04-13T18:30:00Z&holdId=hold_001&holdExpiresAt=2099-04-13T18:37:00Z', {
      onBeforeLoad(win) {
        win.localStorage.setItem(
          'auth_user',
          JSON.stringify({
            id: 'u1',
            name: 'Test User',
            phone: '1234567890',
            email: 'test@example.com',
            isAdmin: false
          })
        );
        win.localStorage.setItem('token', 'test-token');
      }
    });

    cy.get('#card-number', { timeout: 10000 }).type('4111111111111111');
    cy.get('#card-expiry').type('12/30');
    cy.get('#card-cvv').type('123');
    cy.contains('button', 'Pay now', { timeout: 10000 }).click();
    cy.wait('@checkout');
    cy.url({ timeout: 10000 }).should('include', '/booking-success/bok_001');
    cy.contains('Your tickets are locked in!', { timeout: 10000 }).should('be.visible');
    cy.contains('button', 'Download ticket (PDF)').click();
    cy.wait('@ticket');
  });

  it('releases hold when payment timer expires', () => {
    const apiBase = 'http://localhost:8080/api/v1';

    stubSeatPageApis(apiBase);

    cy.intercept('DELETE', `${apiBase}/bookings/holds/hold_001`, {
      statusCode: 200,
      body: { message: 'booking hold released successfully' }
    }).as('releaseHold');

    cy.visit('/movies/mov_001/payment?theaterId=th_001&showTime=2026-04-13T18:30:00Z&holdId=hold_001&holdExpiresAt=2000-04-13T18:37:00Z', {
      onBeforeLoad(win) {
        win.localStorage.setItem(
          'auth_user',
          JSON.stringify({
            id: 'u1',
            name: 'Test User',
            phone: '1234567890',
            email: 'test@example.com',
            isAdmin: false
          })
        );
        win.localStorage.setItem('token', 'test-token');
      }
    });

    cy.wait('@releaseHold');
    cy.url({ timeout: 10000 }).should('include', '/seats');
    cy.url().should('include', 'expired=1');
    cy.wait(['@showDetails', '@seatMap'], { timeout: 15000 });
    cy.contains('Back to showtimes', { timeout: 10000 }).should('be.visible');
  });
});
