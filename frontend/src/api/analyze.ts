import type { AnalyzeResponse, ErrorResponse } from '../types';

export async function analyzeUrl(url: string): Promise<AnalyzeResponse> {
  const resp = await fetch('/api/analyze', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ url }),
  });

  if (!resp.ok) {
    const err: ErrorResponse = await resp.json();
    throw err;
  }

  return resp.json();
}
