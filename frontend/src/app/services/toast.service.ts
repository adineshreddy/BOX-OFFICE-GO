import { Injectable, signal } from '@angular/core';

@Injectable({ providedIn: 'root' })
export class ToastService {
  readonly message = signal<string | null>(null);
  private timer: ReturnType<typeof setTimeout> | null = null;

  /** Shows a short-lived message (e.g. success feedback). */
  show(text: string, durationMs = 4000): void {
    if (this.timer) {
      clearTimeout(this.timer);
      this.timer = null;
    }
    this.message.set(text);
    this.timer = setTimeout(() => {
      this.message.set(null);
      this.timer = null;
    }, durationMs);
  }
}
