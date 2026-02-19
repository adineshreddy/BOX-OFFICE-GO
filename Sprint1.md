# Sprint 1 Report

## Sprint Overview
Sprint 1 focused on building core foundations for the ticket booking platform with parallel frontend and backend tracks. The team prioritized authentication, movie catalog, and theater/showtime discovery while keeping FE-BE integration partial for this sprint.

## User Stories

### FE-00: Setup Authentication Pages
As a user, I want dedicated signup and login pages so that I can create an account and sign in before checkout-related actions.

### FE-01: Browse Movies on Home Screen
As a user, I want to browse movies on the home page so that I can discover what to watch.

### FE-02: View Movie Details and Select Showtime
As a user, I want to open a movie and view detailed info with available showtimes so that I can choose a suitable show.

### FE-03: View Theater Seat Map and Select Seats
As a user, I want to see the seat map and choose seats so that I can prepare my booking before payment.

### FE-04: Enforce Login Before Payment
As a user, I want browsing and movie selection to be possible without login, but login must be mandatory before payment.

### BE-01: Provide Movie Catalog APIs
As a frontend client, I want APIs for listing movies so that the application can show movie catalog data.

### BE-02: Provide Showtime and Theater Layout APIs
As a frontend client, I want APIs for showtimes and theater layout data so that users can select theater and seat options.

### BE-03: Provide Authentication APIs
As a frontend client, I want signup and login APIs so that users can authenticate securely.

### BE-04: Provide Booking Hold and Checkout APIs
As a frontend client, I want booking hold and checkout APIs so that selected seats can be reserved and purchased.

### BE-05: View Theaters and Showtimes for Selected Movie
As a user, when selecting a movie, I want to see theaters and upcoming showtimes so that I can pick where and when to watch.

---

## Issues Planned for Sprint 1
(From team tracker board)

- BOX-OFFICE-GO #15 — FE-00: setup authentication pages
- BOX-OFFICE-GO #8 — FE-01: Browse Movies on Home Screen
- BOX-OFFICE-GO #4 — FE-02: View Movie Details and Select Showtime
- BOX-OFFICE-GO #5 — FE-03: View Theater Seat Map and Select Seats
- BOX-OFFICE-GO #6 — FE-04: Enforce Login Before Payment
- BOX-OFFICE-GO #7 — BE-01: Provide Movie Catalog APIs
- BOX-OFFICE-GO #9 — BE-02: Provide Showtime and Theater Layout APIs
- BOX-OFFICE-GO #10 — BE-03: Provide Authentication APIs
- BOX-OFFICE-GO #11 — BE-04: Provide Booking Hold and Checkout APIs
- BOX-OFFICE-GO #18 — BE-05: View Theaters and Showtimes for Selected Movie

---

## Successfully Completed

### Frontend
- BOX-OFFICE-GO #15 (FE-00) — Completed
  - Signup/Login pages created.
- BOX-OFFICE-GO #8 (FE-01) — Completed
  - Home screen movie browsing UI implemented.

### Backend
- BOX-OFFICE-GO #7 (BE-01) — Completed
  - Movie catalog API implemented (`GET /api/v1/movies`).
- BOX-OFFICE-GO #10 (BE-03) — Completed
  - Authentication APIs implemented (`POST /api/v1/auth/signup`, `POST /api/v1/auth/login`) with validation and password hashing.
- BOX-OFFICE-GO #18 (BE-05) — Completed
  - Theaters + showtimes for selected movie API implemented (`GET /api/v1/movies/{movieId}/theaters`) with date filter support and runtime-gap validation.

---

## Not Completed (or Partially Completed) and Why

### In Progress
- BOX-OFFICE-GO #9 (BE-02) — In Progress
  - Reason: Theater/showtime support is partially available, but full theater layout + seat-level endpoints are not fully completed yet.

### Not Started / Pending
- BOX-OFFICE-GO #4 (FE-02) — Not completed
  - Reason: FE detail page + showtime selection flow still pending integration.
- BOX-OFFICE-GO #5 (FE-03) — Not completed
  - Reason: Seat-map UI and seat selection flow are pending backend seat-layout endpoints.
- BOX-OFFICE-GO #6 (FE-04) — Not completed
  - Reason: End-to-end checkout/login gating flow not finalized in this sprint.
- BOX-OFFICE-GO #11 (BE-04) — Not completed
  - Reason: Booking hold, checkout, and ticket lifecycle APIs were deferred to next sprint due to scope and dependency order.

---

## Sprint 1 Outcome Summary
- Core authentication and movie discovery foundations are complete.
- Movie-to-theater showtime discovery is complete at API level.
- Seat layout and booking pipeline remain for the next sprint, along with FE flows dependent on those APIs.
