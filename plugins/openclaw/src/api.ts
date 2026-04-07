/**
 * Minimal OpenClaw plugin API shim.
 */

export interface PluginAPI {
  config: any;
  logger: { info: (msg: string) => void; warn: (msg: string) => void; error: (msg: string) => void };
  registerTool(id: string, descriptor: any, handler: Function): void;
  registerHook(hook: string, fn: Function): void;
}

export function registerTool(api: PluginAPI, id: string, descriptor: any, handler: Function): void {
  if (typeof (api as any).registerTool === 'function') {
    (api as any).registerTool(id, descriptor, handler);
  } else {
    api.logger.warn('registerTool not available for: ' + id);
  }
}

export function registerHook(api: PluginAPI, hook: string, fn: Function): void {
  if (typeof (api as any).registerHook === 'function') {
    (api as any).registerHook(hook, fn);
  }
}
