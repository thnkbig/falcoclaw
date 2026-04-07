export interface FalcoClawInvestigation {
  alert_id: string;
  rule: string;
  priority: string;
  hostname: string;
  process_tree?: any;
  files_accessed?: string[];
  network_connections?: any[];
  recommended_actions?: string[];
  context: Record<string, any>;
}

/** Handle FalcoClaw investigation dispatch from webhook. */
export async function onInvestigation(api: any, investigation: FalcoClawInvestigation): Promise<void> {
  api.logger.info('[falcoclaw] Investigation alert=' + investigation.alert_id + ' host=' + investigation.hostname);
  // Format into agent-readable summary via formatInvestigation()
}

export function formatInvestigation(inv: FalcoClawInvestigation): string {
  const lines: string[] = [
    '## FalcoClaw Investigation — Alert ' + inv.alert_id,
    '',
    '**Rule:** ' + inv.rule,
    '**Priority:** ' + inv.priority,
    '**Host:** ' + inv.hostname,
    '',
  ];
  if (inv.process_tree) {
    lines.push('**Process Tree:**');
    lines.push(formatTree(inv.process_tree, 0));
    lines.push('');
  }
  if (inv.files_accessed?.length) {
    lines.push('**Files Accessed:** ' + inv.files_accessed.join(', '));
    lines.push('');
  }
  if (inv.network_connections?.length) {
    for (const c of inv.network_connections) {
      lines.push('**Connection:** ' + c.source + ' -> ' + c.destination + ' (' + c.protocol + ')');
    }
    lines.push('');
  }
  if (inv.recommended_actions?.length) {
    lines.push('**Recommended Actions:**');
    for (const a of inv.recommended_actions) lines.push('  - ' + a);
  }
  return lines.join('
');
}

function formatTree(node: any, depth: number): string {
  if (!node) return '';
  const indent = '  '.repeat(depth);
  let out = indent + node.name + ' (PID ' + node.pid + ')';
  if (node.children?.length) {
    for (const c of node.children) out += '
' + formatTree(c, depth + 1);
  }
  return out;
}
