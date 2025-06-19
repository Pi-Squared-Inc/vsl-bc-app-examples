import { ScrollArea, ScrollBar } from "@/components/ui/scroll-area";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHeader, TableRow } from "@/components/ui/table";
import { cn } from "@/lib/utils";
import { useMemo, useState } from "react";
import { Dialog } from "./Dialog";
import { Pagination } from "./Pagination";
import { ValidationRecordRow } from "./ValidationRecordRow";
import { ValidationRecordsTableHeaderCell } from "./ValidationRecordsTableHeaderCell";
import { VerificationRecord } from "@/types";

const SKELETON_ROWS = 25;

function useDialog() {
  const [isOpen, setIsOpen] = useState(false);
  const [title, setTitle] = useState("");
  const [content, setContent] = useState("");
  const [isJson, setIsJson] = useState(false);
  const [isLoading, setIsLoading] = useState(false);

  const open = () => setIsOpen(true);
  const close = () => setIsOpen(false);

  return {
    state: { isOpen, title, content, isJson, isLoading },
    setTitle,
    setContent,
    setIsJson,
    setIsLoading,
    open,
    close,
  };
}

// Define interfaces for the expected API responses
// interface ClaimResponse {
//   claim?: unknown; // Use unknown instead of any
//   size?: number;
// }

// interface VerificationContextResponse {
//   verification_context?: unknown; // Use unknown instead of any
//   size?: number;
// }

interface ValidationRecordsTableProps {
  records: VerificationRecord[];
  isLoading: boolean;
  isPaging?: boolean;
  currentPage: number;
  totalPages: number;
  onPageChange: (page: number) => void;
}

export function ValidationRecordsTable({ records, isLoading, isPaging = false, currentPage, totalPages, onPageChange }: ValidationRecordsTableProps) {
  const dialog = useDialog();

  const effectiveTotalPages = useMemo(() => (totalPages > 0 ? totalPages : 1), [totalPages]);

  const skeletonRows = useMemo(() => {
    const numCols = 4;
    return Array.from({ length: SKELETON_ROWS }).map((_, index) => (
      <TableRow key={`skeleton-${index}`} className="animate-pulse h-[53px]" style={{ animationDelay: `${index * 50}ms` }}>
        {[...Array(numCols)].map((_, i) => (
          <TableCell key={i} className="px-4 align-middle">
            <Skeleton className="h-6 w-24" />
          </TableCell>
        ))}
      </TableRow>
    ));
  }, []);

  // const handleViewData = async (title: string, recordId: number, contentType: "claim" | "proof") => {
  //   dialog.setTitle(title);
  //   dialog.setContent(`Loading ${contentType} details...`);
  //   dialog.setIsJson(false);
  //   dialog.setIsLoading(true);
  //   dialog.open();

  //   try {
  //     // Extract the relevant data based on contentType and type assertion
  //     const contentData = "hello world"; // Placeholder for actual data fetching logic

  //     if (contentData) {
  //       let finalContent = "";
  //       let isFinalContentJson = false;
  //       // Check if contentData is a string (likely a JSON string)
  //       if (typeof contentData === "string") {
  //         try {
  //           // Attempt to parse the JSON string
  //           const parsedData = JSON.parse(contentData);
  //           // Re-stringify the parsed object for pretty printing
  //           finalContent = JSON.stringify(parsedData, null, 2);
  //           isFinalContentJson = true;
  //         } catch (parseError) {
  //           // If parsing fails, it might not be JSON, display as plain text
  //           console.error("Failed to parse contentData as JSON:", parseError);
  //           finalContent = contentData;
  //           isFinalContentJson = false;
  //         }
  //       } else if (typeof contentData === "object" && contentData !== null) {
  //         // If it's already an object (less likely based on user info, but handle defensively)
  //         finalContent = JSON.stringify(contentData, null, 2);
  //         isFinalContentJson = true;
  //       } else {
  //         // Handle other unexpected types (numbers, booleans, null, etc.)
  //         finalContent = String(contentData);
  //         isFinalContentJson = false;
  //       }

  //       dialog.setContent(finalContent);
  //       dialog.setIsJson(isFinalContentJson);
  //     } else {
  //       // Handle cases where the response was received but didn't contain the expected data
  //       dialog.setContent(`No ${contentType} data is available for Record ${recordId}`);
  //       dialog.setIsJson(false);
  //     }
  //   } catch (error) {
  //     dialog.setContent(`Error loading ${contentType} data for Record ${recordId}: ` + (error instanceof Error ? error.message : String(error)));
  //     dialog.setIsJson(false);
  //   } finally {
  //     dialog.setIsLoading(false);
  //   }
  // };

  const handleViewError = (title: string, content: string) => {
    dialog.setTitle(title);
    dialog.setContent(content);
    dialog.setIsJson(false);
    dialog.setIsLoading(false);
    dialog.open();
  };

  return (
    <section aria-label="Validation Records data" className="records-table-container">
      <Dialog
        isOpen={dialog.state.isOpen}
        onClose={dialog.close}
        title={dialog.state.title}
        content={dialog.state.content}
        isJson={dialog.state.isJson}
        isLoading={dialog.state.isLoading}
      />

      <ScrollArea className="w-full">
        <Table className="table-fixed w-full min-w-[1040px]">
          <TableHeader>
            <TableRow className="hover:bg-transparent">
              <ValidationRecordsTableHeaderCell>{`Created At`}</ValidationRecordsTableHeaderCell>
              <ValidationRecordsTableHeaderCell>{`Type`}</ValidationRecordsTableHeaderCell>
              <ValidationRecordsTableHeaderCell>{`Status`}</ValidationRecordsTableHeaderCell>
              <ValidationRecordsTableHeaderCell>{`Claim`}</ValidationRecordsTableHeaderCell>
              <ValidationRecordsTableHeaderCell>{`Result`}</ValidationRecordsTableHeaderCell>
            </TableRow>
          </TableHeader>
          <TableBody className={cn("transition-opacity duration-200", isPaging && "opacity-60")}>
            {isLoading && !isPaging ? (
              <>{skeletonRows}</>
            ) : records?.length > 0 ? (
              records?.map((record) => (
                <ValidationRecordRow
                  key={record.id}
                  record={record}
                  //  onViewData={handleViewData}
                  onViewError={handleViewError}
                />
              ))
            ) : (
              <TableRow>
                <TableCell colSpan={5} className="text-center h-24 px-4">
                  No verification records available
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
        <ScrollBar orientation="horizontal" />
      </ScrollArea>
      <div className="mt-6">
        <Pagination currentPage={currentPage} totalPages={effectiveTotalPages} onPageChange={onPageChange} disabled={isLoading || isPaging} />
      </div>
    </section>
  );
}
