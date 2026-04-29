# Frontend (Angular)

## Requirements
- Node.js 20+
- npm
- Running backend API at `http://localhost:8080`

## Setup
```bash
cd frontend
npm install
```

## Run Locally
```bash
npm start
```
Then open `http://localhost:4200`.

## User Flow
- Log in / sign up
- Browse movies and choose a showtime
- Select seats and continue to payment
- Enter payment method, card number, card expiry, and CVV
- Confirm booking and download ticket PDF
- Manage bookings in **My Bookings**

## Tests
- Unit tests: `npm test`
- E2E tests: `npm run cypress:run`

## Build
```bash
npm run build
```
