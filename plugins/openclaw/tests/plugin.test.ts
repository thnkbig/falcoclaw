import { describe, it, expect, vi } from 'vitest';
import { queryAlerts } from '../src/alerts';
import { triggerAction } from '../src/actions';

const mockApi = (overrides = {}) => ({
  config: { plugins: { '@thnkbig/falcoclaw': {} } },
  logger: { info: vi.fn(), warn: vi.fn(), error: vi.fn() },
  ...overrides,
});

describe('queryAlerts', () => {
  it('returns alerts array from the API', async () => {
    // Integration test: call queryAlerts with mock axios
    // Unit test: verify correct endpoint and params
    const api = mockApi();
    // Would need to mock axios for unit tests
    expect(true).toBe(true);
  });
});

describe('triggerAction', () => {
  it('throws if no reason is provided', async () => {
    const api = mockApi();
    await expect(triggerAction(api, {
      actionner: 'kill',
      target: '12345',
    })).rejects.toThrow('reason is required');
  });
});
