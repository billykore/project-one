import { afterEach, vi } from 'vitest';

// Keep React's act warnings quiet in the jsdom test environment.
// ponytail: this is the smallest shared switch for React 19 testing.
globalThis.IS_REACT_ACT_ENVIRONMENT = true;

afterEach(() => {
  vi.restoreAllMocks();
  document.body.innerHTML = '';
});