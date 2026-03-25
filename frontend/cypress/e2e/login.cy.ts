describe('Login (Cypress)', () => {
  it('logs in and shows Log out on the home page', () => {
    cy.clearLocalStorage();

    cy.intercept('POST', 'http://localhost:8080/api/v1/auth/login', {
      statusCode: 200,
      body: {
        message: 'ok',
        user: {
          id: 'u1',
          name: 'Test User',
          phone: '1234567890',
          email: 'test@example.com',
          isAdmin: false
        }
      }
    }).as('login');

    cy.intercept('GET', 'http://localhost:8080/api/v1/movies', {
      statusCode: 200,
      body: { movies: [] }
    }).as('listMovies');

    cy.visit('/login');

    cy.get('input[placeholder="you@example.com"]').type('test@example.com');
    cy.get('input[type="password"]').type('password123');
    cy.get('button[type="submit"]').click();

    cy.wait('@login');

    cy.contains('button', 'Log out', { timeout: 10000 }).should('be.visible');
    cy.contains('h2.section-title', 'Recommended Movies', { timeout: 10000 }).should('exist');
  });
});

