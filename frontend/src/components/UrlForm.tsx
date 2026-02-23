import { useState, type FormEvent } from 'react';

interface Props {
  onSubmit: (url: string) => void;
  loading: boolean;
}

export function UrlForm({ onSubmit, loading }: Props) {
  const [url, setUrl] = useState('');

  const handleSubmit = (e: FormEvent) => {
    e.preventDefault();
    if (url.trim()) {
      onSubmit(url.trim());
    }
  };

  return (
    <form className="url-form" onSubmit={handleSubmit}>
      <input
        type="url"
        placeholder="https://example.com"
        value={url}
        onChange={(e) => setUrl(e.target.value)}
        required
        disabled={loading}
      />
      <button type="submit" disabled={loading || !url.trim()}>
        Analyze
      </button>
    </form>
  );
}
