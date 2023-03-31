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
      "flex border-t border-gray-2 bg-white px-4 py-3 sm:px-6 ",
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
>(({ children, key, onClick, active, icon = false }, ref) => {
  const inactiveClass =
    "relative inline-flex items-center px-4 py-2 text-sm font-semibold text-gray-900 ring-1 ring-inset ring-gray-3 hover:bg-gray-1 focus:z-20 focus:outline-offset-0 cursor-pointer dark:bg-gray-12 dark:text-gray-1";
  const activeClass =
    "relative z-10 inline-flex items-center bg-gray-12 px-4 py-2 text-sm font-semibold text-gray-1 focus:z-20 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-600 cursor-pointer dark:bg-gray-dark-12 dark:text-gray-dark-1";

  return icon ? (
    <button
      ref={ref}
      key={key}
      onClick={onClick}
      className="relative inline-flex cursor-pointer items-center p-2 text-gray-4 ring-1 ring-inset ring-gray-3 hover:bg-gray-1 focus:z-20 focus:outline-offset-0 dark:bg-gray-12 dark:text-gray-1"
    >
      {children}
    </button>
  ) : (
    <button
      ref={ref}
      key={key}
      onClick={onClick}
      aria-current="page"
      className={active ? activeClass : inactiveClass}
    >
      {children}
    </button>
  );
});
PaginationLink.displayName = "PaginationLink";
