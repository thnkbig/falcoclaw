import axios from 'axios';

export interface FalcoClawAlert {
  id: string;
  timestamp: string;
  falco_rule: string;
  priority: string;
  process_name: string;
  process_pid: number;
  hostname: string;
  output: string;
  action_taken?: string;
  tags?: string[];
}

export interface QueryAlertsOptions {
  priority?: string;
  since?: string;
  limit?: number;
  rule?: string;
  hostname?: string;
}

function getEndpoint(api: any): string {
  const cfg = api.config?.plugins?.['@thnkbig/falcoclaw'] ?? {};
  return cfg.endpoint ?? 'http://localhost:2804';
}

function getApiKey(api: any): string | undefined {
  return api.config?.plugins?.['@thnkbig/falcoclaw']?.apiKey;
}

/** Query FalcoClaw alert history. */
export async function queryAlerts(api: any, options: QueryAlertsOptions): Promise<FalcoClawAlert[]> {
  const endpoint = getEndpoint(api);
  const apiKey = getApiKey(api);
  const params: Record<string, string | number> = {};
  if (options.priority) params.priority = options.priority;
  if (options.since) params.since = options.since;
  if (options.limit) params.limit = options.limit;
  if (options.rule) params.rule = options.rule;
  if (options.hostname) params.hostname = options.hostname;

  try {
    const response = await axios.get(endpoint + '/api/alerts', {
      headers: apiKey ? { Authorization: 'Bearer ' + apiKey } : {},
      params,
      timeout: 10000,
    });
    return response.data.alerts ?? response.data ?? [];
  } catch (err: any) {
    api.logger.error('[falcoclaw] queryAlerts failed: ' + err.message);
    throw new Error('FalcoClaw alert query failed: ' + err.message);
  }
}
