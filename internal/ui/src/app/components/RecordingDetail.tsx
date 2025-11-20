import React, { useState, useEffect } from "react";
import { format } from "date-fns";
import { ArrowLeft, Copy, Check, Loader2 } from "lucide-react";
import { Button } from "./ui/button";
import { useQuery } from "@tanstack/react-query";
import { useNavigate, useSearchParams } from "react-router";

interface ParsedStream {
  text: string;
  metadata: Record<string, any>;
  eventCounts: Record<string, number>;
}

interface Recording {
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
    durationMs: number;
  };
  error?: string;
}

interface RecordingDetailProps {
  recordingId: string;
}

async function fetchRecording(id: string): Promise<Recording> {
  const response = await fetch(`/api/recordings/${id}`);
  if (!response.ok) {
    throw new Error("Failed to fetch recording");
  }
  return response.json();
}

async function fetchParsedRecording(id: string): Promise<ParsedStream> {
  const response = await fetch(`/api/recordings/${id}/parse`);
  if (!response.ok) {
    throw new Error("Failed to parse recording");
  }
  return response.json();
}

export default function RecordingDetail({ recordingId }: RecordingDetailProps) {
  const [copiedSection, setCopiedSection] = useState<string | null>(null);
  const [searchParams, setSearchParams] = useSearchParams();
  const activeTab = searchParams.get("tab") || "request";
  const navigate = useNavigate();

  const setActiveTab = (tab: string) => {
    setSearchParams(
      (prev) => {
        const newParams = new URLSearchParams(prev);
        newParams.set("tab", tab);
        return newParams;
      },
      { replace: true },
    );
  };

  const { data: recording, isLoading: isLoadingRecording } = useQuery({
    queryKey: ["recording", recordingId],
    queryFn: () => fetchRecording(recordingId),
    enabled: !!recordingId,
  });

  const {
    data: parsedData,
    isLoading: isParsing,
    error: parseError,
  } = useQuery({
    queryKey: ["parsed", recordingId],
    queryFn: () => fetchParsedRecording(recordingId),
    enabled: activeTab === "parsed" && !!recording?.response.streaming,
  });

  const copyToClipboard = (text: string, section: string) => {
    navigator.clipboard.writeText(text);
    setCopiedSection(section);
    setTimeout(() => setCopiedSection(null), 2000);
  };

  const formatJSON = (obj: any) => {
    try {
      return JSON.stringify(obj, null, 2);
    } catch {
      return String(obj);
    }
  };

  const formatBody = (body: any) => {
    if (typeof body === "string") {
      return body;
    }
    return formatJSON(body);
  };

  const getStatusColor = (status: number) => {
    if (status >= 200 && status < 300)
      return "text-green-600 dark:text-green-400 bg-green-50 dark:bg-green-900/20";
    if (status >= 400 && status < 500)
      return "text-yellow-600 dark:text-yellow-400 bg-yellow-50 dark:bg-yellow-900/20";
    if (status >= 500)
      return "text-red-600 dark:text-red-400 bg-red-50 dark:bg-red-900/20";
    return "text-gray-600 dark:text-gray-400 bg-gray-50 dark:bg-gray-800";
  };

  if (isLoadingRecording && !recording) {
    return (
      <div className="flex-1 flex items-center justify-center py-12">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
        <span className="ml-2 text-muted-foreground">Loading recording...</span>
      </div>
    );
  }

  if (!recording) {
    return (
      <div className="flex-1 flex items-center justify-center py-12">
        <span className="ml-2 text-muted-foreground">Recording not found.</span>
      </div>
    );
  }

  return (
    <div className="w-full h-full flex flex-col bg-background text-foreground">
      {/* Header */}
      <div className="flex items-center justify-between p-6 border-b bg-card">
        <div className="flex items-center gap-4">
          <button
            onClick={() => navigate("/recordings")}
            className="p-2 hover:bg-muted rounded-md transition-colors"
          >
            <ArrowLeft className="h-5 w-5" />
          </button>
          <div>
            <h2 className="text-xl font-bold">Recording Details</h2>
            {recording && (
              <p className="text-sm text-muted-foreground font-mono mt-1">
                {recording.id}
              </p>
            )}
          </div>
        </div>
      </div>

      {/* Main Content */}
      <div className="flex-1 flex flex-col overflow-hidden bg-background">
        {/* Metadata & Tabs Header */}
        <div className="bg-card border-b">
          <div className="p-6 pb-0">
            <div className="grid grid-cols-2 gap-4 mb-6">
              <div>
                <label className="text-sm font-medium text-muted-foreground">
                  Timestamp
                </label>
                <p className="text-sm mt-1">
                  {format(
                    new Date(recording.timestamp),
                    "MMM d, yyyy HH:mm:ss",
                  )}
                </p>
              </div>
              <div>
                <label className="text-sm font-medium text-muted-foreground">
                  Provider
                </label>
                <p className="text-sm mt-1">
                  <span className="inline-flex items-center px-2 py-1 rounded text-xs font-medium bg-muted text-muted-foreground">
                    {recording.provider}
                  </span>
                </p>
              </div>
              <div>
                <label className="text-sm font-medium text-muted-foreground">
                  Duration
                </label>
                <p className="text-sm mt-1">{recording.timing.durationMs}ms</p>
              </div>
              <div>
                <label className="text-sm font-medium text-muted-foreground">
                  Status
                </label>
                <p className="text-sm mt-1">
                  <span
                    className={`inline-flex items-center px-2 py-1 rounded text-xs font-medium ${getStatusColor(recording.response.status)}`}
                  >
                    {recording.response.status}
                  </span>
                </p>
              </div>
            </div>

            {recording.error && (
              <div className="p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-md mb-6">
                <label className="text-sm font-medium text-red-800 dark:text-red-300">
                  Error
                </label>
                <p className="text-sm text-red-700 dark:text-red-400 mt-1 font-mono">
                  {recording.error}
                </p>
              </div>
            )}

            {/* Tabs */}
            <div className="flex gap-6 border-b -mb-px">
              {["request", "response", "parsed"].map((tab) => (
                <button
                  key={tab}
                  onClick={() => setActiveTab(tab)}
                  className={`
                    pb-3 text-sm font-medium transition-colors border-b-2
                    ${
                      activeTab === tab
                        ? "border-primary text-primary"
                        : "border-transparent text-muted-foreground hover:text-foreground hover:border-border"
                    }
                  `}
                >
                  {tab.charAt(0).toUpperCase() +
                    tab.slice(1) +
                    (tab === "parsed" ? " Response" : "")}
                </button>
              ))}
            </div>
          </div>
        </div>

        {/* Tab Content */}
        <div className="flex-1 overflow-y-auto p-6 bg-muted/10">
          {activeTab === "request" && (
            <div className="bg-card border rounded-md">
              <div className="flex items-center justify-between p-3 bg-muted/30 border-b">
                <h3 className="font-semibold">Request</h3>
                <Button
                  size="sm"
                  variant="ghost"
                  onClick={() =>
                    copyToClipboard(
                      formatJSON({
                        method: recording.request.method,
                        path: recording.request.path,
                        query: recording.request.query,
                        headers: recording.request.headers,
                        body: recording.request.body,
                      }),
                      "request",
                    )
                  }
                >
                  {copiedSection === "request" ? (
                    <>
                      <Check className="h-4 w-4 mr-1" />
                      Copied
                    </>
                  ) : (
                    <>
                      <Copy className="h-4 w-4 mr-1" />
                      Copy
                    </>
                  )}
                </Button>
              </div>
              <div className="p-4 space-y-3">
                <div>
                  <label className="text-sm font-medium text-muted-foreground">
                    Endpoint
                  </label>
                  <p className="text-sm mt-1 font-mono">
                    {recording.request.method} {recording.request.path}
                    {recording.request.query && `?${recording.request.query}`}
                  </p>
                </div>

                <div>
                  <label className="text-sm font-medium text-muted-foreground">
                    Headers
                  </label>
                  <pre className="text-xs mt-1 p-3 bg-muted/50 rounded overflow-x-auto">
                    {formatJSON(recording.request.headers)}
                  </pre>
                </div>

                {recording.request.body && (
                  <div>
                    <label className="text-sm font-medium text-muted-foreground">
                      Body
                    </label>
                    <pre className="text-xs mt-1 p-3 bg-muted/50 rounded overflow-x-auto overflow-y-auto max-h-[500px] w-full whitespace-pre-wrap">
                      {formatBody(recording.request.body)}
                    </pre>
                  </div>
                )}
              </div>
            </div>
          )}

          {activeTab === "response" && (
            <div className="bg-card border rounded-md">
              <div className="flex items-center justify-between p-3 bg-muted/30 border-b">
                <div className="flex items-center gap-2">
                  <h3 className="font-semibold">Response</h3>
                  {recording.response.streaming && (
                    <span className="text-xs px-2 py-1 bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-300 rounded">
                      Streaming
                    </span>
                  )}
                </div>
                <Button
                  size="sm"
                  variant="ghost"
                  onClick={() =>
                    copyToClipboard(
                      formatJSON({
                        status: recording.response.status,
                        headers: recording.response.headers,
                        body: recording.response.body,
                      }),
                      "response",
                    )
                  }
                >
                  {copiedSection === "response" ? (
                    <>
                      <Check className="h-4 w-4 mr-1" />
                      Copied
                    </>
                  ) : (
                    <>
                      <Copy className="h-4 w-4 mr-1" />
                      Copy
                    </>
                  )}
                </Button>
              </div>
              <div className="p-4 space-y-3">
                <div>
                  <label className="text-sm font-medium text-muted-foreground">
                    Headers
                  </label>
                  <pre className="text-xs mt-1 p-3 bg-muted/50 rounded overflow-x-auto">
                    {formatJSON(recording.response.headers)}
                  </pre>
                </div>

                {recording.response.body && (
                  <div>
                    <label className="text-sm font-medium text-muted-foreground">
                      Body
                    </label>
                    <pre className="text-xs mt-1 p-3 bg-muted/50 rounded overflow-x-auto overflow-y-auto max-h-[500px] w-full whitespace-pre-wrap">
                      {formatBody(recording.response.body)}
                    </pre>
                  </div>
                )}
              </div>
            </div>
          )}

          {activeTab === "parsed" && (
            <div className="bg-card border rounded-md">
              <div className="flex items-center justify-between p-3 bg-muted/30 border-b">
                <h3 className="font-semibold">Parsed Response</h3>
              </div>
              <div className="p-4">
                {!recording.response.streaming ? (
                  <div className="text-center py-8 text-muted-foreground">
                    This recording is not streaming, so there is no parsed view
                    available.
                  </div>
                ) : (
                  <div className="mt-1">
                    {isParsing && (
                      <div className="flex items-center justify-center p-8">
                        <Loader2 className="h-6 w-6 animate-spin text-primary" />
                        <span className="ml-2 text-muted-foreground">
                          Parsing stream...
                        </span>
                      </div>
                    )}
                    {parseError && (
                      <div className="p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded text-red-800 dark:text-red-300 text-sm">
                        Failed to parse stream: {(parseError as Error).message}
                      </div>
                    )}
                    {parsedData && !isParsing && (
                      <div className="space-y-4">
                        {/* Reconstructed Output */}
                        <div>
                          <h4 className="text-sm font-semibold mb-2">
                            Reconstructed Output
                          </h4>
                          <div className="p-3 bg-muted/50 rounded border">
                            {parsedData.text ? (
                              <p className="text-sm whitespace-pre-wrap">
                                {parsedData.text}
                              </p>
                            ) : (
                              <p className="text-sm text-muted-foreground italic">
                                No text content
                              </p>
                            )}
                          </div>
                        </div>

                        {/* Metadata */}
                        {Object.keys(parsedData.metadata).length > 0 && (
                          <div>
                            <h4 className="text-sm font-semibold mb-2">
                              Metadata
                            </h4>
                            <div className="grid grid-cols-2 gap-2">
                              {Object.entries(parsedData.metadata).map(
                                ([key, value]) => (
                                  <div
                                    key={key}
                                    className="p-2 bg-muted/50 rounded border"
                                  >
                                    <div className="text-xs font-medium text-muted-foreground">
                                      {key}
                                    </div>
                                    <div className="text-sm mt-1">
                                      {typeof value === "object" ? (
                                        <pre className="text-xs whitespace-pre-wrap overflow-x-auto">
                                          {JSON.stringify(value, null, 2)}
                                        </pre>
                                      ) : (
                                        String(value)
                                      )}
                                    </div>
                                  </div>
                                ),
                              )}
                            </div>
                          </div>
                        )}

                        {/* Event Summary */}
                        {Object.keys(parsedData.eventCounts).length > 0 && (
                          <div>
                            <h4 className="text-sm font-semibold mb-2">
                              Event Summary
                            </h4>
                            <div className="p-3 bg-muted/50 rounded border">
                              <div className="flex flex-wrap gap-2">
                                {Object.entries(parsedData.eventCounts).map(
                                  ([eventType, count]) => (
                                    <span
                                      key={eventType}
                                      className="inline-flex items-center px-2 py-1 rounded text-xs bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-300"
                                    >
                                      {eventType}: {count}
                                    </span>
                                  ),
                                )}
                              </div>
                            </div>
                          </div>
                        )}
                      </div>
                    )}
                  </div>
                )}
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

