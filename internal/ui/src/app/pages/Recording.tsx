import React, { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { useParams, useNavigate } from "react-router";
import RecordingDetailDrawer from "../components/RecordingDetailDrawer";

interface RecordingSummary {
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

interface RecordingListResponse {
  recordings: RecordingSummary[];
  total: number;
  page: number;
  limit: number;
  hasMore: boolean;
}

async function fetchRecordings(
  page: number,
  limit: number,
  provider: string,
  search: string,
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

export default function Recording() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [page] = useState(1);
  const [limit] = useState(50);
  const [provider] = useState("");
  const [search] = useState("");

  // Fetch recordings list for the sidebar navigation
  const { data } = useQuery({
    queryKey: ["recordings", page, limit, provider, search],
    queryFn: () => fetchRecordings(page, limit, provider, search),
    refetchInterval: 10000, // Auto-refresh every 10 seconds
    refetchIntervalInBackground: true,
  });

  if (!id) {
    navigate("/recordings");
    return null;
  }

  if (!data?.recordings) {
    return (
      <div className="flex items-center justify-center py-12">
        <div className="text-gray-500">Loading recordings...</div>
      </div>
    );
  }

  return (
    <RecordingDetailDrawer
      recordings={data.recordings}
      initialRecordingId={id}
      onClose={() => navigate("/recordings")}
      onNavigate={(id) => navigate(`/recordings/${id}`)}
    />
  );
}
