import { Component, inject } from '@angular/core';
import { ToastService } from '../../services/toast.service';

@Component({
  selector: 'app-toast',
  standalone: true,
  template: `
    @if (toast.message(); as msg) {
      <div class="toast" role="status" aria-live="polite">{{ msg }}</div>
    }
  `,
  styles: `
    :host {
      display: contents;
    }
    .toast {
      position: fixed;
      bottom: 1.5rem;
      left: 50%;
      transform: translateX(-50%);
      z-index: 10000;
      max-width: min(420px, calc(100vw - 2rem));
      padding: 0.75rem 1.25rem;
      border-radius: 12px;
      font-size: 0.9rem;
      font-weight: 600;
      color: #f0fdf4;
      background: linear-gradient(135deg, rgba(22, 101, 52, 0.95), rgba(21, 128, 61, 0.92));
      border: 1px solid rgba(134, 239, 172, 0.45);
      box-shadow:
        0 12px 40px rgba(0, 0, 0, 0.35),
        0 0 0 1px rgba(255, 255, 255, 0.06) inset;
      animation: toast-in 0.22s ease-out;
    }
    @keyframes toast-in {
      from {
        opacity: 0;
        transform: translateX(-50%) translateY(12px);
      }
      to {
        opacity: 1;
        transform: translateX(-50%) translateY(0);
      }
    }
  `
})
export class ToastComponent {
  readonly toast = inject(ToastService);
}
