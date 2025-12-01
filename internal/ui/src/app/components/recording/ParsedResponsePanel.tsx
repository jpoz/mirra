import React from "react";
import { Loader2 } from "lucide-react";
import { ParsedStream, Recording } from "@/lib/api";

interface ParsedResponsePanelProps {
  recording: Recording;
  parsedData?: ParsedStream;
  isLoading: boolean;
  error?: Error | null;
}

/**
 * Displays parsed streaming response data
 */
export function ParsedResponsePanel({
  recording,
  parsedData,
  isLoading,
  error,
}: ParsedResponsePanelProps) {
  return (
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
            {isLoading && (
              <div className="flex items-center justify-center p-8">
                <Loader2 className="h-6 w-6 animate-spin text-primary" />
                <span className="ml-2 text-muted-foreground">
                  Parsing stream...
                </span>
              </div>
            )}
            {error && (
              <div className="p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded text-red-800 dark:text-red-300 text-sm">
                Failed to parse stream: {error.message}
              </div>
            )}
            {parsedData && !isLoading && (
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
                    <h4 className="text-sm font-semibold mb-2">Metadata</h4>
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
                        )
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
                          )
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
  );
}
