import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { ChevronLeft, ChevronRight, MoreHorizontal } from "lucide-react";

interface PaginationProps {
  currentPage: number;
  totalPages: number;
  onPageChange: (page: number) => void;
  disabled?: boolean;
}

export function Pagination({
  currentPage,
  totalPages,
  onPageChange,
  disabled = false,
}: PaginationProps) {
  const effectiveTotalPages = Math.max(1, totalPages);

  // Common button styles
  const baseButtonStyles = "h-8";
  const iconButtonStyles = cn(baseButtonStyles, "w-8 min-w-[2.5rem]");
  const pageButtonStyles = cn(baseButtonStyles, "min-w-[2.5rem] px-1");

  const getPageNumbers = () => {
    const maxVisiblePages = 5;
    const pageNumbers = [];

    if (effectiveTotalPages <= maxVisiblePages) {
      for (let i = 1; i <= effectiveTotalPages; i++) {
        pageNumbers.push(i);
      }
    } else {
      const halfVisible = Math.floor(maxVisiblePages / 2);
      let startPage = Math.max(1, currentPage - halfVisible);
      const endPage = Math.min(
        effectiveTotalPages,
        startPage + maxVisiblePages - 1
      );

      if (endPage === effectiveTotalPages) {
        startPage = Math.max(1, endPage - maxVisiblePages + 1);
      }

      // First page and ellipsis
      if (startPage > 1) {
        pageNumbers.push(1);
        if (startPage > 2) {
          pageNumbers.push("ellipsis-start");
        }
      }

      // Middle pages
      for (let i = startPage; i <= endPage; i++) {
        pageNumbers.push(i);
      }

      // Last page and ellipsis
      if (endPage < effectiveTotalPages) {
        if (endPage < effectiveTotalPages - 1) {
          pageNumbers.push("ellipsis-end");
        }
        pageNumbers.push(effectiveTotalPages);
      }
    }

    return pageNumbers;
  };

  const handlePageChange = (page: number) => {
    if (
      page >= 1 &&
      page <= effectiveTotalPages &&
      page !== currentPage &&
      !disabled
    ) {
      onPageChange(page);
    }
  };

  return (
    <nav
      className="flex items-center justify-center space-x-2"
      aria-label="Pagination"
    >
      <Button
        variant="outline"
        size="icon"
        className={iconButtonStyles}
        onClick={() => handlePageChange(currentPage - 1)}
        disabled={disabled || currentPage <= 1}
        aria-label="Previous page"
      >
        <ChevronLeft className="h-4 w-4" />
        <span className="sr-only">Previous page</span>
      </Button>

      {getPageNumbers().map((page, index) => {
        if (typeof page === "string" && page.startsWith("ellipsis")) {
          return (
            <Button
              key={page}
              variant="outline"
              size="icon"
              className={iconButtonStyles}
              disabled={true}
              aria-hidden="true"
            >
              <MoreHorizontal className="h-4 w-4" />
              <span className="sr-only">More pages</span>
            </Button>
          );
        }

        return (
          <Button
            key={`page-${page}-${index}`}
            variant={currentPage === page ? "default" : "outline"}
            className={pageButtonStyles}
            onClick={() => handlePageChange(page as number)}
            disabled={disabled}
            aria-current={currentPage === page ? "page" : undefined}
          >
            {page}
          </Button>
        );
      })}

      <Button
        variant="outline"
        size="icon"
        className={iconButtonStyles}
        onClick={() => handlePageChange(currentPage + 1)}
        disabled={disabled || currentPage >= effectiveTotalPages}
        aria-label="Next page"
      >
        <ChevronRight className="h-4 w-4" />
        <span className="sr-only">Next page</span>
      </Button>
    </nav>
  );
}
