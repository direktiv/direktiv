import React from "react";
import clsx from "clsx";

interface PaginationProps {
  children: React.ReactNode;
  align?: "center" | "left" | "right";
}

export const Pagination: React.FC<PaginationProps> = ({
  children,
  align = "right",
}) => (
  <div
    className={clsx(
      "flex border-t border-gray-2 bg-gray-1 px-4 py-3 dark:border-gray-dark-2 dark:bg-gray-dark-1 sm:px-6 ",
      align === "center" && "justify-center",
      align === "right" && "justify-end",
      align === "left" && "justify-start"
    )}
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
  key?: string;
  onClick?: () => void;
  active?: boolean;
  icon?: boolean;
}
export const PaginationLink = React.forwardRef<
  HTMLButtonElement,
  React.ButtonHTMLAttributes<HTMLButtonElement> & PaginationLinkProps
>(({ children, key, onClick, active, icon = false }, ref) =>
  icon ? (
    <button
      ref={ref}
      key={key}
      onClick={onClick}
      className="relative inline-flex cursor-pointer items-center bg-gray-1 p-2  text-gray-4 ring-1 ring-inset ring-gray-3 hover:bg-gray-1 focus:z-20 focus:outline-offset-0 dark:bg-gray-dark-1 dark:text-gray-dark-4 dark:ring-gray-dark-3 dark:hover:bg-gray-dark-1"
    >
      {children}
    </button>
  ) : (
    <button
      ref={ref}
      key={key}
      onClick={onClick}
      aria-current="page"
      className={clsx(
        "relative inline-flex cursor-pointer items-center px-4 py-2 text-sm font-semibold focus:z-20",
        active &&
          "z-10 bg-gray-12 text-gray-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-gray-1 dark:bg-gray-dark-12 dark:text-gray-dark-1  dark:focus-visible:outline-gray-dark-1",
        !active &&
          "text-gray-12 ring-1 ring-inset ring-gray-3 hover:bg-gray-1 focus:outline-offset-0 dark:bg-gray-12 dark:text-gray-dark-12 dark:ring-gray-dark-3 dark:hover:bg-gray-dark-1"
      )}
    >
      {children}
    </button>
  )
);
PaginationLink.displayName = "PaginationLink";
