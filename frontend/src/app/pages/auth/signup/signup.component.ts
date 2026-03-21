import { Component } from '@angular/core';
import { Router, RouterLink } from '@angular/router';
import {
  AbstractControl,
  ReactiveFormsModule,
  FormBuilder,
  FormGroup,
  ValidationErrors,
  Validators
} from '@angular/forms';
import { AuthService, formatAuthHttpError } from '../../../services/auth.service';
import { CommonModule } from '@angular/common';

function passwordsMatchValidator(group: AbstractControl): ValidationErrors | null {
  const fg = group as FormGroup;
  const p = fg.get('password')?.value as string | undefined;
  const c = fg.get('confirmPassword')?.value as string | undefined;
  if (!p || !c) {
    return null;
  }
  return p === c ? null : { passwordMismatch: true };
}

@Component({
  selector: 'app-signup',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, RouterLink],
  templateUrl: './signup.component.html',
  styleUrl: './signup.component.scss'
})
export class SignupComponent {
  form: FormGroup;
  error = '';
  loading = false;

  constructor(
    private fb: FormBuilder,
    private auth: AuthService,
    private router: Router
  ) {
    this.form = this.fb.nonNullable.group(
      {
        name: ['', [Validators.required, Validators.minLength(2)]],
        phone: ['', [Validators.required, Validators.pattern(/^[0-9+]{8,15}$/)]],
        email: ['', [Validators.required, Validators.email]],
        password: ['', [Validators.required, Validators.minLength(8)]],
        confirmPassword: ['', Validators.required]
      },
      { validators: [passwordsMatchValidator] }
    );
  }

  onSubmit() {
    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }
    this.error = '';
    this.loading = true;
    const v = this.form.getRawValue();
    this.auth
      .signup({
        name: v.name,
        phone: v.phone,
        email: v.email,
        password: v.password,
        confirmPassword: v.confirmPassword
      })
      .subscribe({
        next: () => {
          this.loading = false;
          void this.router.navigate(['/login'], { queryParams: { signedUp: 'true' } });
        },
        error: err => {
          this.loading = false;
          this.error = formatAuthHttpError(err);
        }
      });
  }
}
