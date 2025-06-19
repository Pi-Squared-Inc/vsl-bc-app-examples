import { ScrollArea, ScrollBar } from "@/components/ui/scroll-area";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { ValidationRecord } from "@/types";
import { useMemo, useState } from "react";
import { Dialog } from "./Dialog";
import { Pagination } from "./Pagination";
import EthereumMirroringRecordRow from "./EthereumMirroringRecordRow";
import { ValidationRecordsTableHeaderCell } from "./ValidationRecordsTableHeaderCell";

const SKELETON_ROWS = 25;
const Cols = 5;

function useDialog() {
  const [isOpen, setIsOpen] = useState(false);
  const [title, setTitle] = useState("");
  const [content, setContent] = useState("");
  const [isJson, setIsJson] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [byteSize, setByteSize] = useState<number | null>(null);

  const open = () => setIsOpen(true);
  const close = () => setIsOpen(false);

  return {
    state: { isOpen, title, content, isJson, isLoading, byteSize },
    setTitle,
    setContent,
    setIsJson,
    setIsLoading,
    setByteSize,
    open,
    close,
  };
}

interface EthereumMirroringTableProps {
  claims: ValidationRecord[];
  isLoading: boolean;
  isRefreshing?: boolean;
  isPaging?: boolean;
  currentPage: number;
  totalPages: number;
  onPageChange: (page: number) => void;
}

export default function EthereumMirroringTable({
  claims,
  isLoading,
  isRefreshing = false,
  isPaging = false,
  currentPage,
  totalPages,
  onPageChange,
}: EthereumMirroringTableProps) {
  const dialog = useDialog();

  const effectiveTotalPages = useMemo(
    () => (totalPages > 0 ? totalPages : 1),
    [totalPages]
  );

  const skeletonRows = useMemo(() => {
    return Array.from({ length: SKELETON_ROWS }).map((_, index) => (
      <TableRow
        key={`skeleton-${index}`}
        className="animate-pulse h-[53px] [&_td]:border-r"
        style={{ animationDelay: `${index * 50}ms` }}
      >
        <TableCell className="px-4 align-middle">
          <Skeleton className="h-6 w-24" />
        </TableCell>
        {[...Array(Cols - 1)].map((_, i) => (
          <TableCell key={i} className="px-4 align-middle">
            <Skeleton className="h-6 w-24" />
          </TableCell>
        ))}
      </TableRow>
    ));
  }, []);

  const handleViewError = (title: string, content: string) => {
    dialog.setTitle(title);
    dialog.setContent(content);
    dialog.setIsJson(false);
    dialog.setByteSize(null);
    dialog.setIsLoading(false);
    dialog.open();
  };

  return (
    <section aria-label="Claims data" className="claims-table-container">
      <Dialog
        isOpen={dialog.state.isOpen}
        onClose={dialog.close}
        title={dialog.state.title}
        content={dialog.state.content}
        isJson={dialog.state.isJson}
        isLoading={dialog.state.isLoading}
        byteSize={dialog.state.byteSize}
      />

      <ScrollArea className="w-full">
        <Table className="table-fixed border w-full min-w-[920px]">
          <TableHeader className="border">
            <TableRow className="hover:bg-transparent [&_th]:border-r">
              <TableHead
                rowSpan={2}
                className="px-4 text-pi2-accent-white font-semibold text-base whitespace-nowrap first:border-r-2"
              >
                Block Number
              </TableHead>
              <TableHead
                colSpan={2}
                className="px-4 text-pi2-accent-white font-semibold text-base whitespace-nowrap text-center"
              >
                Reth
              </TableHead>
              <TableHead
                colSpan={2}
                className="px-4 text-pi2-accent-white font-semibold text-base whitespace-nowrap text-center"
              >
                Geth
              </TableHead>
            </TableRow>
            <TableRow className="hover:bg-transparent [&_th]:border-r [&_.claim]:border-r-0">
              <ValidationRecordsTableHeaderCell className="claim">{`Claim`}</ValidationRecordsTableHeaderCell>
              <ValidationRecordsTableHeaderCell className="time">{`Verification Time`}</ValidationRecordsTableHeaderCell>
              <ValidationRecordsTableHeaderCell className="claim">{`Claim`}</ValidationRecordsTableHeaderCell>
              <ValidationRecordsTableHeaderCell className="time">{`Verification Time`}</ValidationRecordsTableHeaderCell>
            </TableRow>
          </TableHeader>
          <TableBody
            className={
              isRefreshing || isPaging
                ? "opacity-60 transition-opacity duration-200"
                : ""
            }
          >
            {isLoading && !isPaging ? (
              <>{skeletonRows}</>
            ) : claims?.length > 0 ? (
              claims?.map((claim) => (
                <EthereumMirroringRecordRow
                  key={claim.block_number}
                  claim={claim}
                  onViewError={handleViewError}
                />
              ))
            ) : (
              <TableRow>
                <TableCell colSpan={Cols} className="text-center h-24 px-4">
                  No claims data available
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
        <ScrollBar orientation="horizontal" />
      </ScrollArea>
      <div className="mt-6">
        <Pagination
          currentPage={currentPage}
          totalPages={effectiveTotalPages}
          onPageChange={onPageChange}
          disabled={isLoading || isRefreshing || isPaging}
        />
      </div>
    </section>
  );
}
