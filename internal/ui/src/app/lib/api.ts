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
