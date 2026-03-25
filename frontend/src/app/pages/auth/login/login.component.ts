import { Component } from '@angular/core';
import { Router, RouterLink, ActivatedRoute } from '@angular/router';
import { ReactiveFormsModule, FormBuilder, FormGroup, Validators } from '@angular/forms';
import { AuthService, formatAuthHttpError } from '../../../services/auth.service';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'app-login',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, RouterLink],
  templateUrl: './login.component.html',
  styleUrl: './login.component.scss'
})
export class LoginComponent {
  form: FormGroup;
  error = '';
  loading = false;
  /** Shown when user came from "See All" or after signup */
  signInMessage = '';

  constructor(
    private fb: FormBuilder,
    private auth: AuthService,
    private router: Router,
    private route: ActivatedRoute
  ) {
    this.form = this.fb.nonNullable.group({
      email: ['', [Validators.required, Validators.email]],
      password: ['', Validators.required]
    });
    this.route.queryParams.subscribe((params) => {
      if (params['signedUp'] === 'true') {
        this.signInMessage = 'Account created. Please sign in.';
      } else if (params['from'] === 'see-all-movies') {
        this.signInMessage = 'Sign in to browse the full movie catalog.';
      } else {
        this.signInMessage = '';
      }
    });
  }

  onSubmit() {
    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }
    this.error = '';
    this.loading = true;
    this.auth.login(this.form.getRawValue()).subscribe({
      next: res => {
        this.loading = false;
        this.auth.setSessionFromLogin(res);
        const pendingRaw = localStorage.getItem('pending_booking_hold');
        if (pendingRaw) {
          try {
            const pending = JSON.parse(pendingRaw) as {
              movieId?: string;
              theaterId?: string;
              showTime?: string;
            };
            if (pending.movieId && pending.theaterId && pending.showTime) {
              void this.router.navigate(['/movies', pending.movieId, 'seats'], {
                queryParams: {
                  theaterId: pending.theaterId,
                  showTime: pending.showTime
                }
              });
              return;
            }
          } catch {
            // If parsing fails, fall back to default behavior.
          }
        }
        void this.router.navigate(['/']);
      },
      error: err => {
        this.loading = false;
        this.error = formatAuthHttpError(err);
      }
    });
  }
}
