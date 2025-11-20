import React, { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { useNavigate } from "react-router";
import { format } from "date-fns";
import { Search, RefreshCw } from "lucide-react";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "../components/ui/table";
import { Button } from "../components/ui/button";
import { Input } from "../components/ui/input";

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

export default function Recordings() {
  const navigate = useNavigate();
  const [page, setPage] = useState(1);
  const [limit] = useState(50);
  const [provider, setProvider] = useState("");
  const [search, setSearch] = useState("");
  const [searchInput, setSearchInput] = useState("");

  // Fetch recordings list with auto-refresh every 10 seconds
  const { data, isLoading, error, refetch, isFetching } = useQuery({
    queryKey: ["recordings", page, limit, provider, search],
    queryFn: () => fetchRecordings(page, limit, provider, search),
    refetchInterval: 10000, // Auto-refresh every 10 seconds
    refetchIntervalInBackground: true,
  });

  const handleSearch = () => {
    setSearch(searchInput);
    setPage(1);
  };

  const handleClearFilters = () => {
    setProvider("");
    setSearch("");
    setSearchInput("");
    setPage(1);
  };

  const handleRefresh = () => {
    refetch();
  };

  const truncateId = (id: string) => {
    return id.substring(0, 8);
  };

  const getStatusColor = (status: number) => {
    if (status >= 200 && status < 300) return "text-green-600 dark:text-green-400";
    if (status >= 400 && status < 500) return "text-yellow-600 dark:text-yellow-400";
    if (status >= 500) return "text-red-600 dark:text-red-400";
    return "text-muted-foreground";
  };

  const formatBytes = (bytes: number): string => {
    if (bytes === 0) return "0 B";
    if (bytes < 1024) return `${bytes} B`;
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
  };

  const getSizeColor = (bytes: number): string => {
    const tenKB = 10 * 1024;
    const oneMB = 1024 * 1024;

    if (bytes < tenKB) return "text-green-600 dark:text-green-400";
    if (bytes < oneMB) return "text-yellow-600 dark:text-yellow-400";
    return "text-red-600 dark:text-red-400";
  };

  return (
    <div className="space-y-4">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-foreground">Recordings</h1>
          <p className="text-sm text-muted-foreground">
            {data?.total ? `${data.total} total recordings` : "Loading..."}
          </p>
        </div>
        <Button
          onClick={handleRefresh}
          disabled={isFetching}
          className="flex items-center gap-2"
        >
          <RefreshCw
            className={`h-4 w-4 ${isFetching ? "animate-spin" : ""}`}
          />
          Refresh
        </Button>
      </div>

      {/* Filters */}
      <div className="flex gap-4 items-end">
        <div className="flex-1">
          <label className="text-sm font-medium mb-1 block text-muted-foreground">Search</label>
          <div className="flex gap-2">
            <Input
              placeholder="Search by ID, path, or error..."
              value={searchInput}
              onChange={(e) => setSearchInput(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === "Enter") {
                  handleSearch();
                }
              }}
            />
            <Button onClick={handleSearch} className="flex items-center gap-2">
              <Search className="h-4 w-4" />
              Search
            </Button>
          </div>
        </div>
        <div className="w-48">
          <label className="text-sm font-medium mb-1 block text-muted-foreground">Provider</label>
          <select
            value={provider}
            onChange={(e) => {
              setProvider(e.target.value);
              setPage(1);
            }}
            className="w-full px-3 py-2 border rounded-md bg-background text-foreground border-input"
          >
            <option value="">All Providers</option>
            <option value="openai">OpenAI</option>
            <option value="claude">Claude</option>
            <option value="gemini">Gemini</option>
          </select>
        </div>
        {(provider || search) && (
          <Button variant="outline" onClick={handleClearFilters}>
            Clear Filters
          </Button>
        )}
      </div>

      {/* Error State */}
      {error && (
        <div className="p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-md text-red-800 dark:text-red-300">
          Error loading recordings: {(error as Error).message}
        </div>
      )}

      {/* Loading State */}
      {isLoading && (
        <div className="flex items-center justify-center py-12">
          <div className="text-muted-foreground">Loading recordings...</div>
        </div>
      )}

      {/* Table */}
      {!isLoading && data && (
        <>
          <div className="border rounded-md bg-card">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>ID</TableHead>
                  <TableHead>Timestamp</TableHead>
                  <TableHead>Provider</TableHead>
                  <TableHead>Method</TableHead>
                  <TableHead>Path</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Duration</TableHead>
                  <TableHead>Size</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {data.recordings.length === 0 ? (
                  <TableRow>
                    <TableCell
                      colSpan={8}
                      className="text-center py-8 text-muted-foreground"
                    >
                      No recordings found
                    </TableCell>
                  </TableRow>
                ) : (
                  data.recordings.map((recording) => (
                    <TableRow
                      key={recording.id}
                      onClick={() => navigate(`/recordings/${recording.id}`)}
                    >
                      <TableCell className="font-mono text-xs text-muted-foreground">
                        {truncateId(recording.id)}
                      </TableCell>
                      <TableCell className="text-sm text-foreground">
                        {format(
                          new Date(recording.timestamp),
                          "MMM d, HH:mm:ss",
                        )}
                      </TableCell>
                      <TableCell>
                        <span className="inline-flex items-center px-2 py-1 rounded text-xs font-medium bg-muted text-muted-foreground">
                          {recording.provider}
                        </span>
                      </TableCell>
                      <TableCell className="text-sm font-mono text-foreground">
                        {recording.method}
                      </TableCell>
                      <TableCell className="text-sm font-mono max-w-xs truncate text-foreground">
                        {recording.path}
                      </TableCell>
                      <TableCell
                        className={`font-medium ${getStatusColor(recording.status)}`}
                      >
                        {recording.status}
                      </TableCell>
                      <TableCell className="text-sm text-foreground">
                        {recording.duration}ms
                      </TableCell>
                      <TableCell
                        className={`text-sm font-mono ${getSizeColor(recording.responseSize)}`}
                      >
                        {formatBytes(recording.responseSize)}
                      </TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          </div>

          {/* Pagination */}
          {data.recordings.length > 0 && (
            <div className="flex items-center justify-between">
              <div className="text-sm text-muted-foreground">
                Page {data.page} of {Math.ceil(data.total / data.limit)}
              </div>
              <div className="flex gap-2">
                <Button
                  variant="outline"
                  onClick={() => setPage(page - 1)}
                  disabled={page === 1}
                >
                  Previous
                </Button>
                <Button
                  variant="outline"
                  onClick={() => setPage(page + 1)}
                  disabled={!data.hasMore}
                >
                  Next
                </Button>
              </div>
            </div>
          )}
        </>
      )}
    </div>
  );
}
