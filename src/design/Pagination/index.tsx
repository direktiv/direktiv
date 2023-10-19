import { RxChevronLeft, RxChevronRight } from "react-icons/rx";

import React from "react";
import { twMergeClsx } from "~/util/helpers";

interface PaginationProps {
  children: React.ReactNode;
  align?: "center" | "left";
}

export const Pagination: React.FC<PaginationProps> = ({ children, align }) => (
  <div
    className={twMergeClsx(
      "flex",
      align === "center" && "justify-center",
      align === "left" && "justify-start",
      !align && "justify-end"
    )}
    data-testid="pagination-wrapper"
  >
    <nav
      className="isolate inline-flex -space-x-px rounded-md shadow-sm"
      aria-label="Pagination"
    >
      {children}
    </nav>
  </div>
);
Pagination.displayName = "Pagination";
export interface PaginationLinkProps {
  onClick?: () => void;
  active?: boolean;
  icon?: "left" | "right";
}
export const PaginationLink = React.forwardRef<
  HTMLButtonElement,
  React.ButtonHTMLAttributes<HTMLButtonElement> & PaginationLinkProps
>(({ children, active, icon = false, ...props }, ref) =>
  icon ? (
    <button
      ref={ref}
      {...props}
      className={twMergeClsx(
        "relative inline-flex cursor-pointer items-center ring-1 ring-inset focus:z-20 focus:outline-offset-0",
        "disabled:pointer-events-none disabled:opacity-50",
        "p-2 text-gray-9 ring-gray-7 hover:bg-gray-2 focus-visible:outline-gray-9",
        "dark:text-gray-dark-9 dark:ring-gray-dark-7 dark:hover:bg-gray-dark-2 dark:focus-visible:outline-gray-dark-9",
        icon === "left" && "rounded-l-md",
        icon === "right" && "rounded-r-md"
      )}
    >
      {icon === "left" ? (
        <RxChevronLeft className="h-5 w-5" aria-hidden="true" />
      ) : (
        <RxChevronRight className="h-5 w-5" aria-hidden="true" />
      )}
    </button>
  ) : (
    <button
      ref={ref}
      {...props}
      aria-current={active ? "page" : undefined}
      className={twMergeClsx(
        "relative inline-flex cursor-pointer items-center px-4 py-2 text-sm font-semibold ring-1 ring-inset focus:z-20 ",
        "disabled:pointer-events-none disabled:opacity-50",
        "text-gray-12 ring-gray-7 focus:outline-offset-0 focus-visible:outline-gray-9",
        "dark:text-gray-dark-12 dark:ring-gray-dark-7 dark:focus-visible:outline-gray-dark-9",
        active && "z-10 bg-gray-3 dark:bg-gray-dark-3 ",
        !active && "hover:bg-gray-2 dark:hover:bg-gray-dark-2"
      )}
    >
      {children}
    </button>
  )
);
PaginationLink.displayName = "PaginationLink";
