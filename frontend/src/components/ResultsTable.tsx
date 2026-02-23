import type { AnalyzeResponse } from '../types';

interface Props {
  data: AnalyzeResponse;
}

const headingOrder = ['h1', 'h2', 'h3', 'h4', 'h5', 'h6'];

export function ResultsTable({ data }: Props) {
  return (
    <div className="results">
      <table>
        <tbody>
          <tr>
            <th>HTML Version</th>
            <td>{data.htmlVersion}</td>
          </tr>
          <tr>
            <th>Page Title</th>
            <td>{data.title || <em>No title</em>}</td>
          </tr>
        </tbody>
      </table>

      <h3>Headings</h3>
      <table>
        <thead>
          <tr>
            {headingOrder.map((h) => (
              <th key={h}>{h.toUpperCase()}</th>
            ))}
          </tr>
        </thead>
        <tbody>
          <tr>
            {headingOrder.map((h) => (
              <td key={h}>{data.headings[h] ?? 0}</td>
            ))}
          </tr>
        </tbody>
      </table>

      <h3>Links</h3>
      <table>
        <tbody>
          <tr>
            <th>Internal Links</th>
            <td>{data.internalLinks}</td>
          </tr>
          <tr>
            <th>External Links</th>
            <td>{data.externalLinks}</td>
          </tr>
          <tr>
            <th>Inaccessible Links</th>
            <td>{data.inaccessibleLinks}</td>
          </tr>
        </tbody>
      </table>

      <h3>Login Form</h3>
      <span className={`badge ${data.hasLoginForm ? 'badge-yes' : 'badge-no'}`}>
        {data.hasLoginForm ? 'Yes' : 'No'}
      </span>
    </div>
  );
}
