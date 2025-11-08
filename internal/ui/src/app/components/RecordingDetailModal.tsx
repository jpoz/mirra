import React, { useState } from 'react'
import { format } from 'date-fns'
import { X, Copy, Check } from 'lucide-react'
import { Button } from './ui/button'

interface Recording {
  id: string
  timestamp: string
  provider: string
  request: {
    method: string
    path: string
    query: string
    headers: Record<string, string[]>
    body: any
  }
  response: {
    status: number
    headers: Record<string, string[]>
    body: any
    streaming: boolean
  }
  timing: {
    startedAt: string
    completedAt: string
    durationMs: number
  }
  error?: string
}

interface RecordingDetailModalProps {
  recording: Recording
  onClose: () => void
}

export default function RecordingDetailModal({
  recording,
  onClose,
}: RecordingDetailModalProps) {
  const [copiedSection, setCopiedSection] = useState<string | null>(null)

  const copyToClipboard = (text: string, section: string) => {
    navigator.clipboard.writeText(text)
    setCopiedSection(section)
    setTimeout(() => setCopiedSection(null), 2000)
  }

  const formatJSON = (obj: any) => {
    try {
      return JSON.stringify(obj, null, 2)
    } catch {
      return String(obj)
    }
  }

  const formatBody = (body: any) => {
    if (typeof body === 'string') {
      // Return string as-is to preserve newlines
      return body
    }
    return formatJSON(body)
  }

  const getStatusColor = (status: number) => {
    if (status >= 200 && status < 300) return 'text-green-600 bg-green-50'
    if (status >= 400 && status < 500) return 'text-yellow-600 bg-yellow-50'
    if (status >= 500) return 'text-red-600 bg-red-50'
    return 'text-gray-600 bg-gray-50'
  }

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
      <div className="bg-white rounded-lg shadow-xl max-w-4xl w-full max-h-[90vh] flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between p-6 border-b">
          <div>
            <h2 className="text-xl font-bold">Recording Details</h2>
            <p className="text-sm text-gray-600 font-mono mt-1">{recording.id}</p>
          </div>
          <button
            onClick={onClose}
            className="p-2 hover:bg-gray-100 rounded-md transition-colors"
          >
            <X className="h-5 w-5" />
          </button>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto p-6 space-y-6">
          {/* Metadata */}
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="text-sm font-medium text-gray-600">Timestamp</label>
              <p className="text-sm mt-1">
                {format(new Date(recording.timestamp), 'MMM d, yyyy HH:mm:ss')}
              </p>
            </div>
            <div>
              <label className="text-sm font-medium text-gray-600">Provider</label>
              <p className="text-sm mt-1">
                <span className="inline-flex items-center px-2 py-1 rounded text-xs font-medium bg-gray-100">
                  {recording.provider}
                </span>
              </p>
            </div>
            <div>
              <label className="text-sm font-medium text-gray-600">Duration</label>
              <p className="text-sm mt-1">{recording.timing.durationMs}ms</p>
            </div>
            <div>
              <label className="text-sm font-medium text-gray-600">Status</label>
              <p className="text-sm mt-1">
                <span
                  className={`inline-flex items-center px-2 py-1 rounded text-xs font-medium ${getStatusColor(
                    recording.response.status
                  )}`}
                >
                  {recording.response.status}
                </span>
              </p>
            </div>
          </div>

          {/* Error */}
          {recording.error && (
            <div className="p-4 bg-red-50 border border-red-200 rounded-md">
              <label className="text-sm font-medium text-red-800">Error</label>
              <p className="text-sm text-red-700 mt-1 font-mono">{recording.error}</p>
            </div>
          )}

          {/* Request */}
          <div className="border rounded-md">
            <div className="flex items-center justify-between p-3 bg-gray-50 border-b">
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
                    'request'
                  )
                }
              >
                {copiedSection === 'request' ? (
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
                <label className="text-sm font-medium text-gray-600">Endpoint</label>
                <p className="text-sm mt-1 font-mono">
                  {recording.request.method} {recording.request.path}
                  {recording.request.query && `?${recording.request.query}`}
                </p>
              </div>

              <div>
                <label className="text-sm font-medium text-gray-600">Headers</label>
                <pre className="text-xs mt-1 p-3 bg-gray-50 rounded overflow-x-auto">
                  {formatJSON(recording.request.headers)}
                </pre>
              </div>

              {recording.request.body && (
                <div>
                  <label className="text-sm font-medium text-gray-600">Body</label>
                  <pre className="text-xs mt-1 p-3 bg-gray-50 rounded overflow-x-auto max-h-64 overflow-y-auto whitespace-pre-wrap">
                    {formatBody(recording.request.body)}
                  </pre>
                </div>
              )}
            </div>
          </div>

          {/* Response */}
          <div className="border rounded-md">
            <div className="flex items-center justify-between p-3 bg-gray-50 border-b">
              <div className="flex items-center gap-2">
                <h3 className="font-semibold">Response</h3>
                {recording.response.streaming && (
                  <span className="text-xs px-2 py-1 bg-blue-100 text-blue-700 rounded">
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
                    'response'
                  )
                }
              >
                {copiedSection === 'response' ? (
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
                <label className="text-sm font-medium text-gray-600">Headers</label>
                <pre className="text-xs mt-1 p-3 bg-gray-50 rounded overflow-x-auto">
                  {formatJSON(recording.response.headers)}
                </pre>
              </div>

              {recording.response.body && (
                <div>
                  <label className="text-sm font-medium text-gray-600">Body</label>
                  <pre className="text-xs mt-1 p-3 bg-gray-50 rounded overflow-x-auto max-h-96 overflow-y-auto whitespace-pre-wrap">
                    {formatBody(recording.response.body)}
                  </pre>
                </div>
              )}
            </div>
          </div>
        </div>

        {/* Footer */}
        <div className="flex items-center justify-end gap-2 p-6 border-t bg-gray-50">
          <Button variant="outline" onClick={onClose}>
            Close
          </Button>
        </div>
      </div>
    </div>
  )
}
