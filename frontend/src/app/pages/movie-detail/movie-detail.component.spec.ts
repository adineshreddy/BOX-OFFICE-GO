import { vi } from 'vitest';
import { MovieDetailComponent } from './movie-detail.component';

describe('MovieDetailComponent', () => {
  it('posterUrl() returns provided url or the fallback', () => {
    const component = new MovieDetailComponent({} as any, { navigate: vi.fn() } as any, {} as any);

    expect(component.posterUrl('http://example.com/poster.jpg')).toBe('http://example.com/poster.jpg');
    expect(component.posterUrl(null)).toBe(component.posterFallback);
    expect(component.posterUrl(undefined)).toBe(component.posterFallback);
  });

  it('onPosterError() swaps to fallback only once', () => {
    const component = new MovieDetailComponent({} as any, { navigate: vi.fn() } as any, {} as any);

    const img = { dataset: {} as Record<string, string>, src: '' } as any as HTMLImageElement;
    const event = { target: img } as any;

    component.onPosterError(event);

    expect(img.dataset['usedFallback']).toBe('1');
    expect(img.src).toBe(component.posterFallback);

    const srcAfterFirst = img.src;
    component.onPosterError(event);

    expect(img.src).toBe(srcAfterFirst);
  });
});

