import type { ErrorResponse } from '../types';

interface Props {
  error: ErrorResponse;
}

export function ErrorMessage({ error }: Props) {
  return (
    <div className="error-message">
      <strong>Error {error.statusCode}</strong>
      <p>{error.message}</p>
    </div>
  );
}
