import React, { useState } from "react";
import { Copy, Check } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Recording } from "@/lib/api";
import { formatJSON, formatBody } from "@/lib/formatters";

interface ResponsePanelProps {
  recording: Recording;
}

/**
 * Displays response details: status, headers, body
 */
export function ResponsePanel({ recording }: ResponsePanelProps) {
  const [copied, setCopied] = useState(false);

  const copyToClipboard = () => {
    const responseData = formatJSON({
      status: recording.response.status,
      headers: recording.response.headers,
      body: recording.response.body,
    });
    navigator.clipboard.writeText(responseData);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="bg-card border rounded-md">
      <div className="flex items-center justify-between p-3 bg-muted/30 border-b">
        <h3 className="font-semibold">Response</h3>
        <Button size="sm" variant="ghost" onClick={copyToClipboard}>
          {copied ? (
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
          <pre className="text-xs bg-muted p-3 rounded-md overflow-x-auto mt-1 font-mono">
            {formatJSON(recording.response.headers)}
          </pre>
        </div>

        <div>
          <label className="text-sm font-medium text-muted-foreground">
            Body
          </label>
          <pre className="text-xs bg-muted p-3 rounded-md overflow-x-auto mt-1 font-mono max-h-96">
            {formatBody(recording.response.body)}
          </pre>
        </div>
      </div>
    </div>
  );
}
