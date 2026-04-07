import axios from 'axios';

export interface TriggerActionOptions {
  actionner: string;
  target: string;
  params?: Record<string, any>;
  reason?: string;
  dry_run?: boolean;
}

function getEndpoint(api: any): string {
  const cfg = api.config?.plugins?.['@thnkbig/falcoclaw'] ?? {};
  return cfg.endpoint ?? 'http://localhost:2804';
}

function getApiKey(api: any): string | undefined {
  return api.config?.plugins?.['@thnkbig/falcoclaw']?.apiKey;
}

/** Trigger a FalcoClaw response action. */
export async function triggerAction(api: any, options: TriggerActionOptions): Promise<any> {
  const endpoint = getEndpoint(api);
  const apiKey = getApiKey(api);

  if (!options.reason) {
    throw new Error('A reason is required when manually triggering a FalcoClaw action.');
  }

  try {
    const response = await axios.post(
      endpoint + '/api/actions',
      {
        actionner: options.actionner,
        target: options.target,
        parameters: options.params ?? {},
        reason: options.reason,
        dry_run: options.dry_run ?? false,
      },
      {
        headers: {
          'Content-Type': 'application/json',
          ...(apiKey ? { Authorization: 'Bearer ' + apiKey } : {}),
        },
        timeout: 10000,
      }
    );
    return response.data;
  } catch (err: any) {
    api.logger.error('[falcoclaw] triggerAction failed: ' + err.message);
    throw new Error('FalcoClaw action trigger failed: ' + err.message);
  }
}
