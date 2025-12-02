export interface RecordingSummary {
  id: string;
  timestamp: string;
  provider: string;
  method: string;
  path: string;
  status: number;
  duration: number;
  responseSize: number;
  error?: string;
}

export interface RecordingListResponse {
  recordings: RecordingSummary[];
  total: number;
  page: number;
  limit: number;
  hasMore: boolean;
}

export interface Recording {
  id: string;
  timestamp: string;
  provider: string;
  request: {
    method: string;
    path: string;
    query: string;
    headers: Record<string, string[]>;
    body: any;
  };
  response: {
    status: number;
    headers: Record<string, string[]>;
    body: any;
    streaming: boolean;
  };
  timing: {
    startedAt: string;
    completedAt: string;
    duration_ms: number;
  };
  error?: string;
}

export interface ParsedStream {
  text: string;
  metadata: Record<string, any>;
  eventCounts: Record<string, number>;
}

export async function fetchRecordings(
  page: number,
  limit: number,
  provider?: string,
  search?: string,
): Promise<RecordingListResponse> {
  const params = new URLSearchParams({
    page: page.toString(),
    limit: limit.toString(),
  });
  if (provider) params.append("provider", provider);
  if (search) params.append("search", search);

  const response = await fetch(`/api/recordings?${params}`);
  if (!response.ok) {
    throw new Error("Failed to fetch recordings");
  }
  return response.json();
}

/**
 * Fetches a single recording by ID
 */
export async function fetchRecording(id: string): Promise<Recording> {
  const response = await fetch(`/api/recordings/${id}`);
  if (!response.ok) {
    throw new Error(`Failed to fetch recording: ${response.statusText}`);
  }
  return response.json();
}

/**
 * Fetches parsed stream data for a recording
 */
export async function fetchParsedRecording(id: string): Promise<ParsedStream> {
  const response = await fetch(`/api/recordings/${id}/parse`);
  if (!response.ok) {
    throw new Error(`Failed to parse recording: ${response.statusText}`);
  }
  return response.json();
}
