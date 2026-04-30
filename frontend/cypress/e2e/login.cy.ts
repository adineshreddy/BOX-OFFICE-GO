describe('Login -> view movie details -> select showtime -> select seats', () => {
  it('logs in, selects a showtime, then selects seats and continues', () => {
    const apiBase = 'http://localhost:8080/api/v1';

    const user = {
      id: 'u1',
      name: 'Test User',
      phone: '1234567890',
      email: 'test@example.com',
      isAdmin: false
    };

    const movie = {
      id: 'mov_001',
      title: 'Dune: Part Two',
      description: 'A test movie description',
      genre: 'Drama',
      language: 'English',
      durationMinutes: 120,
      releaseDate: '2026-01-01',
      rating: 8.6,
      posterUrl: null,
      isActive: true,
      createdAt: '2026-01-01T00:00:00Z',
      updatedAt: '2026-01-01T00:00:00Z'
    };

    const theater = {
      theaterId: 'th_001',
      theaterName: 'Main Theater',
      city: 'Metropolis',
      addressLine1: '123 Main St',
      timezone: 'UTC',
      showtimes: [
        {
          showtimeId: 'st_001',
          screenName: 'Screen 1',
          startTime: '2026-03-25T18:30:00Z',
          endTime: '2026-03-25T20:30:00Z',
          language: 'English',
          format: '2D',
          basePrice: 10
        }
      ]
    };

    // Seats: create exactly 4 candidate seats (available && not held)
    // so simulatedSoldSeats picks 0 seats (floor(4*0.2) = 0), removing randomness.
    const seatMap = {
      showtimeId: 'st_001',
      movieTitle: movie.title,
      theaterName: theater.theaterName,
      screenName: 'Screen 1',
      showTime: theater.showtimes[0].startTime,
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
    };

    cy.intercept('GET', `${apiBase}/movies`, {
      statusCode: 200,
      body: { movies: [movie] }
    });

    cy.intercept('GET', `${apiBase}/movies/mov_001`, {
      statusCode: 200,
      body: movie
    });

    cy.intercept('GET', `${apiBase}/movies/mov_001/theaters*`, {
      statusCode: 200,
      body: {
        movieId: movie.id,
        movieTitle: movie.title,
        durationMinutes: movie.durationMinutes,
        theaters: [theater]
      }
    });

    cy.intercept('GET', `${apiBase}/shows/details*`, {
      statusCode: 200,
      body: {
        showtimeId: 'st_001',
        movieId: movie.id,
        movieTitle: movie.title,
        theaterId: theater.theaterId,
        theaterName: theater.theaterName,
        city: theater.city,
        addressLine1: theater.addressLine1,
        screenName: theater.showtimes[0].screenName,
        startTime: theater.showtimes[0].startTime,
        language: theater.showtimes[0].language,
        format: theater.showtimes[0].format,
        basePrice: theater.showtimes[0].basePrice,
        durationMinutes: movie.durationMinutes,
        availableSeats: 4,
        totalSeats: 4,
        unavailableSeats: 0
      }
    });

    cy.intercept('GET', `${apiBase}/shows/seat-map*`, {
      statusCode: 200,
      body: seatMap
    }).as('seatMap');

    cy.intercept('POST', `${apiBase}/bookings/holds`, {
      statusCode: 200,
      body: {
        hold: {
          holdId: 'hold_001',
          holdExpiresAt: new Date(Date.now() + 5 * 60 * 1000).toISOString(),
          totalAmount: 20
        }
      }
    }).as('hold');

    cy.intercept('POST', `${apiBase}/auth/login`, {
      statusCode: 200,
      body: {
        message: 'ok',
        accessToken: 'test-token',
        tokenType: 'Bearer',
        expiresAt: new Date(Date.now() + 60 * 60 * 1000).toISOString(),
        user
      }
    }).as('login');

    cy.visit('/login', {
      onBeforeLoad(win) {
        win.localStorage.clear();
      }
    });

    // Login page
    cy.get('input[placeholder="you@example.com"]', { timeout: 10000 }).type(user.email);
    cy.get('input[type="password"]', { timeout: 10000 }).type('password123');
    cy.get('button[type="submit"]', { timeout: 10000 }).click();
    cy.wait('@login');

    // Home -> pick the first movie (LoginComponent navigates to '/')
    cy.contains('button', 'Log out', { timeout: 10000 }).should('be.visible');
    cy.get('.movie-card', { timeout: 10000 }).first().click();

    // Movie detail -> select a showtime, then continue to seats
    cy.get('button.showtime-chip', { timeout: 10000 }).first().should('be.visible').click();
    cy.get('button.btn-next', { timeout: 10000 }).should('be.visible').click();
    cy.url({ timeout: 10000 }).should('include', '/seats');

    // Seats page: pick ticket count (closes the modal)
    cy.contains('button.ticket-pill', '2', { timeout: 10000 }).click();

    // Select exactly 2 seats.
    cy.get('button.seat[aria-label="Seat A1"]').click();
    cy.get('button.seat[aria-label="Seat A2"]').click();

    cy.get('button.btn-confirm').contains('Continue').click();

    cy.wait('@hold');
    cy.url({ timeout: 10000 }).should('include', '/payment');
    cy.contains('Complete your payment').should('be.visible');
  });
});

