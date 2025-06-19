import { Button } from "@/components/ui/button";
import { Loader2 } from "lucide-react";
import dynamic from "next/dynamic";

// Dynamically import ReactJson to avoid SSR issues
const ReactJson = dynamic(() => import("react-json-view"), { ssr: false });

interface DialogProps {
  isOpen: boolean;
  onClose: () => void;
  title: string;
  content: string;
  isJson?: boolean;
  isLoading?: boolean;
}

export function Dialog({
  isOpen,
  onClose,
  title,
  content,
  isJson = false,
  isLoading = false,
}: DialogProps) {
  if (!isOpen) return null;

  // Display loading indicator when in loading state
  if (isLoading) {
    return (
      <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
        <div className="bg-background border border-border text-foreground rounded-lg shadow-lg w-11/12 max-w-3xl">
          <div className="p-4 border-b border-border flex justify-between items-center">
            <div className="flex items-center gap-4">
              <h3 className="text-lg font-semibold">{title}</h3>
            </div>
            <button
              onClick={onClose}
              className="text-muted-foreground hover:text-foreground"
            >
              ✕
            </button>
          </div>
          <div className="p-8 flex flex-col items-center justify-center">
            <Loader2 className="h-8 w-8 animate-spin text-primary" />
            <p className="mt-2 text-sm text-muted-foreground">
              Loading data...
            </p>
          </div>
          <div className="p-4 border-t border-border flex justify-end">
            <Button onClick={onClose} variant="outline">
              Cancel
            </Button>
          </div>
        </div>
      </div>
    );
  }

  // Format JSON content or display plain text
  let contentDisplay;
  if (isJson) {
    try {
      // Parse JSON string to object
      const jsonData = JSON.parse(content);

      // Use react-json-view to display JSON data
      contentDisplay = (
        <div className="p-4 rounded overflow-auto max-h-[80vh]">
          <ReactJson
            src={jsonData}
            theme="monokai"
            name={null}
            displayDataTypes={false}
            displayObjectSize={true}
            enableClipboard={true}
            collapsed={1}
            collapseStringsAfterLength={80}
            style={{
              fontFamily: "monospace",
              fontSize: "0.9rem",
              overflowX: "auto",
            }}
          />
        </div>
      );
    } catch (error) {
      // Display error message if JSON parsing fails
      contentDisplay = (
        <pre className="bg-muted text-wrap whitespace-pre-line p-4 rounded overflow-auto max-h-[80vh] text-sm text-destructive">
          Failed to parse JSON:{" "}
          {error instanceof Error ? error.message : String(error)}
        </pre>
      );
    }
  } else {
    // For non-JSON content, display as plain text
    contentDisplay = (
      <pre className="bg-muted text-wrap whitespace-pre-line p-4 rounded overflow-auto max-h-[80vh] text-sm text-foreground">
        {content}
      </pre>
    );
  }

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div className="bg-background border border-border text-foreground rounded-lg shadow-lg w-11/12 max-w-3xl">
        <div className="p-4 border-b border-border flex justify-between items-center">
          <div className="flex items-center gap-4">
            <h3 className="text-lg font-semibold">{title}</h3>
          </div>
          <button
            onClick={onClose}
            className="text-muted-foreground hover:text-foreground"
          >
            ✕
          </button>
        </div>
        <div className="p-4">{contentDisplay}</div>
        <div className="p-4 border-t border-border flex justify-end">
          <Button onClick={onClose} variant="outline">
            Close
          </Button>
        </div>
      </div>
    </div>
  );
}

// Helper function to format bytes to a readable format
// Removed formatBytes function as it's now imported from utils
