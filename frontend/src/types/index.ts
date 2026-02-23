export interface AnalyzeRequest {
  url: string;
}

export interface AnalyzeResponse {
  htmlVersion: string;
  title: string;
  headings: Record<string, number>;
  internalLinks: number;
  externalLinks: number;
  inaccessibleLinks: number;
  hasLoginForm: boolean;
}

export interface ErrorResponse {
  statusCode: number;
  message: string;
}
