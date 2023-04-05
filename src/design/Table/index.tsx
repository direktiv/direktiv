import React from "react";
import clsx from "clsx";

export interface TableRowProps {
  stripe?: boolean;
}

export const Table = React.forwardRef<
  HTMLTableElement,
  React.TableHTMLAttributes<HTMLTableElement>
>(({ children, className, ...props }, ref) => (
  <table
    ref={ref}
    className={clsx(
      "min-w-full divide-y divide-gray-3 dark:bg-gray-10",
      className
    )}
    {...props}
  >
    {children}
  </table>
));
Table.displayName = "Table";

export const TableHead = React.forwardRef<
  HTMLTableSectionElement,
  React.HTMLAttributes<HTMLTableSectionElement>
>(({ children, className, ...props }) => (
  <thead {...props} className={clsx("dark:bg-gray-12", className)}>
    {children}
  </thead>
));
TableHead.displayName = "TableHead";

export const TableBody = React.forwardRef<
  HTMLTableSectionElement,
  React.HTMLAttributes<HTMLTableSectionElement>
>(({ children, className, ...props }) => (
  <tbody
    className={clsx(
      "divide-y divide-gray-2",
      "dark:divide-gray-dark-10 dark:bg-gray-12",
      className
    )}
    {...props}
  >
    {children}
  </tbody>
));
TableBody.displayName = "TableBody";

export const TableCell = React.forwardRef<
  HTMLTableCellElement,
  React.HTMLAttributes<HTMLTableCellElement>
>(({ children, className, ...props }) => (
  <td
    {...props}
    className={clsx(
      "whitespace-nowrap px-3 py-4 text-sm text-gray-9 dark:text-gray-dark-12",
      className
    )}
  >
    {children}
  </td>
));
TableCell.displayName = "TableCell";

export const TableHeaderCell = React.forwardRef<
  HTMLTableCellElement,
  React.HTMLAttributes<HTMLTableCellElement>
>(({ children, className, ...props }) => (
  <th
    {...props}
    className={clsx(
      "px-3 py-3.5 text-left text-sm font-semibold text-gray-12  dark:text-gray-dark-12",
      className
    )}
  >
    {children}
  </th>
));
TableHeaderCell.displayName = "TableHeaderCell";

export const TableRow = React.forwardRef<
  HTMLTableRowElement,
  React.HTMLAttributes<HTMLTableRowElement> & TableRowProps
>(({ children, className, stripe, ...props }) => (
  <tr
    {...props}
    className={clsx(className, stripe && "bg-gray-5 dark:bg-gray-dark-5")}
  >
    {children}
  </tr>
));
TableRow.displayName = "TableRow";
