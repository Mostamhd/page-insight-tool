import { useState } from 'react';
import { analyzeUrl } from './api/analyze';
import { UrlForm } from './components/UrlForm';
import { ResultsTable } from './components/ResultsTable';
import { ErrorMessage } from './components/ErrorMessage';
import { Spinner } from './components/Spinner';
import type { AnalyzeResponse, ErrorResponse } from './types';
import './App.css';

type State =
  | { status: 'idle' }
  | { status: 'loading' }
  | { status: 'success'; data: AnalyzeResponse }
  | { status: 'error'; error: ErrorResponse };

function App() {
  const [state, setState] = useState<State>({ status: 'idle' });

  const handleSubmit = async (url: string) => {
    setState({ status: 'loading' });
    try {
      const data = await analyzeUrl(url);
      setState({ status: 'success', data });
    } catch (err) {
      if (err && typeof err === 'object' && 'statusCode' in err) {
        setState({ status: 'error', error: err as ErrorResponse });
      } else {
        setState({
          status: 'error',
          error: { statusCode: 0, message: 'An unexpected error occurred.' },
        });
      }
    }
  };

  return (
    <div className="app">
      <h1>Page Insight</h1>
      <UrlForm onSubmit={handleSubmit} loading={state.status === 'loading'} />
      {state.status === 'loading' && <Spinner />}
      {state.status === 'success' && <ResultsTable data={state.data} />}
      {state.status === 'error' && <ErrorMessage error={state.error} />}
    </div>
  );
}

export default App;
