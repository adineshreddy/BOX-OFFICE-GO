# BoxOfficeGo

## Project Description
BoxOfficeGo is a movie ticket booking platform with a Go-based backend and an Angular frontend. The system will allow users to browse movies and showtimes, select seats, and book tickets, while admins can manage movies, theaters, and pricing.

## Requirements
- Go 1.22+
- Node.js 20+ and npm
- Neon Postgres database URL for backend (`DATABASE_URL`)

## Run Application
1. **Start backend**
   - `cd backend`
   - copy env template once: `cp .env.example .env`
   - set `DATABASE_URL` in `backend/.env`
   - run: `go run ./cmd/api`
   - backend runs on `http://localhost:8080`
2. **Start frontend**
   - open new terminal, `cd frontend`
   - install deps: `npm install`
   - run dev server: `npm start`
   - frontend runs on `http://localhost:4200`

## How to Use
- Sign up or log in from the home page.
- Browse movies from the home carousel or open **Movies** in the header (`/movies`) for the full list; open a title for details.
- On movie details, choose a **show date** within the next **14 days** (past dates and far-future dates are blocked in the UI and API).
- Pick theater/showtime and continue to seat selection.
- Select seats and continue to checkout.
- Complete payment with:
  - payment method
  - card number (**15** digits for Amex, **16** for most cards; test card `4111111111111111` is 16 digits)
  - card expiry (MM/YY)
  - card CVV
- After payment success, view booking confirmation and download ticket PDF.
- Visit **My Bookings** to view, cancel, or download tickets for existing bookings.

## Testing
- Frontend unit tests: `cd frontend && npm test`
- Frontend e2e tests: `cd frontend && npm run cypress:run`
- Backend tests: `cd backend && go test ./...`

## Git hooks (optional)
From the repo root, enable shared hooks (currently removes a trailing `Made-with: Cursor` line from commit messages if your editor adds it):

```bash
git config core.hooksPath .githooks
```

This only affects this clone; it is not stored in the remote. Each teammate can run the command once after cloning.

## Members
### Front-end

<table style="border-collapse: collapse;" border="1" cellspacing="0" cellpadding="6">
	<thead>
		<tr>
			<th style="border: 1px solid #000;">Name</th>
			<th style="border: 1px solid #000;">UFID</th>
		</tr>
	</thead>
	<tbody>
		<tr>
			<td style="border: 1px solid #000;"><a href="https://github.com/bvsahith">Venkata Sahith Bathini</a></td>
			<td style="border: 1px solid #000;">95306942</td>
		</tr>
		<tr>
			<td style="border: 1px solid #000;"><a href="https://github.com/sandeepelluri11">Sandeep Elluri</a></td>
			<td style="border: 1px solid #000;">16404314</td>
		</tr>
	</tbody>
</table>

### Back-end

<table style="border-collapse: collapse;" border="1" cellspacing="0" cellpadding="6">
	<thead>
		<tr>
			<th style="border: 1px solid #000;">Name</th>
			<th style="border: 1px solid #000;">UFID</th>
		</tr>
	</thead>
	<tbody>
		<tr>
			<td style="border: 1px solid #000;"><a href="https://github.com/adineshreddy">Dinesh Reddy Ande</a></td>
			<td style="border: 1px solid #000;">58723541</td>
		</tr>
		<tr>
			<td style="border: 1px solid #000;"><a href="https://github.com/saaketh-coder">Saaketh Bala</a></td>
			<td style="border: 1px solid #000;">86294284</td>
		</tr>
	</tbody>
</table>

