import { registerTool, registerHook } from './api';
import { queryAlerts } from './alerts';
import { triggerAction } from './actions';
import { onInvestigation, formatInvestigation } from './investigate';

/** FalcoClaw OpenClaw Plugin — v0.1.0 */
export { queryAlerts, triggerAction, onInvestigation, formatInvestigation };

export default {
  name: '@thnkbig/falcoclaw',
  version: '0.1.0',

  async register(api: any) {
    // Tool: query alert history
    registerTool(api, 'falcoclaw_queryAlerts', {
      description: 'Query FalcoClaw alert history. Returns recent security events matched by Falco rules.',
      parameters: {
        type: 'object',
        properties: {
          priority: { type: 'string', description: 'Minimum severity: EMERGENCY|ALERT|CRITICAL|ERROR|WARNING|NOTICE|INFO|DEBUG' },
          since: { type: 'string', description: 'ISO-8601 timestamp. E.g. 2026-04-07T00:00:00Z' },
          limit: { type: 'number', description: 'Max alerts (default 50, max 500)', default: 50 },
          rule: { type: 'string', description: 'Falco rule name filter. E.g. shell_injection' },
          hostname: { type: 'string', description: 'Filter by hostname' },
        },
      },
      handler: queryAlerts,
    });

    // Tool: trigger response action
    registerTool(api, 'falcoclaw_triggerAction', {
      description: 'Manually trigger a FalcoClaw response action. Use when agents detect suspicious activity.',
      parameters: {
        type: 'object',
        required: ['actionner', 'target'],
        properties: {
          actionner: {
            type: 'string',
            enum: ['kill','block_ip','quarantine','disable_user','stop_service','firewall','script',
                   'openclaw_disable_skill','openclaw_revoke_token','openclaw_restart','agent_notify','agent_investigate'],
            description: 'Response action to execute',
          },
          target: { type: 'string', description: 'Action target: PID, IP, file path, username, etc.' },
          params: { type: 'object', description: 'Action-specific parameters. E.g. { signal: "SIGKILL" } for kill.' },
          reason: { type: 'string', description: 'Human-readable justification. Required. Recorded in audit log.' },
          dry_run: { type: 'boolean', description: 'Validate without executing (default false)', default: false },
        },
      },
      handler: triggerAction,
    });

    // Hook: receive FalcoClaw investigation dispatches
    registerHook(api, 'before_agent_start', async (ctx: any) => {
      if (ctx.falcoclaw?.investigation) {
        await onInvestigation(api, ctx.falcoclaw.investigation);
      }
    });

    api.logger.info('@thnkbig/falcoclaw plugin loaded');
  },
};
