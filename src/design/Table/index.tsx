import React from "react";
import clsx from "clsx";

export const Table = React.forwardRef<
  HTMLTableElement,
  React.TableHTMLAttributes<HTMLTableElement>
>(({ children, className, ...props }, ref) => (
  <table
    ref={ref}
    className={clsx(
      "min-w-full divide-y",
      "divide-gray-3",
      "dark:divide-gray-dark-3",
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
  React.HTMLProps<HTMLTableSectionElement>
>(({ children, className, ...props }, ref) => (
  <thead ref={ref} {...props} className={clsx(className)}>
    {children}
  </thead>
));
TableHead.displayName = "TableHead";

export const TableBody = React.forwardRef<
  HTMLTableSectionElement,
  React.HTMLProps<HTMLTableSectionElement>
>(({ children, className, ...props }, ref) => (
  <tbody
    ref={ref}
    className={clsx(
      "divide-y",
      "divide-gray-3",
      "dark:divide-gray-dark-3",
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
  React.HTMLProps<HTMLTableCellElement>
>(({ children, className, ...props }, ref) => (
  <td
    ref={ref}
    {...props}
    className={clsx(
      "whitespace-nowrap px-3 py-2 text-sm",
      "text-gray-9",
      "dark:text-gray-dark-9",
      className
    )}
  >
    {children}
  </td>
));
TableCell.displayName = "TableCell";

export interface TableHeaderCellProps {
  sticky?: boolean;
}

export const TableHeaderCell = React.forwardRef<
  HTMLTableCellElement,
  React.HTMLProps<HTMLTableCellElement> & TableHeaderCellProps
>(({ children, className, sticky, ...props }, ref) => (
  <th
    ref={ref}
    {...props}
    className={clsx(
      "px-3 py-3.5 text-left text-sm font-semibold",
      "text-gray-12",
      "dark:text-gray-dark-12",
      sticky && "sticky top-0 z-10 border-b backdrop-blur",
      sticky && " border-gray-3 bg-white/75",
      sticky && " dark:border-gray-dark-3 dark:bg-black/75",
      className
    )}
  >
    {children}
  </th>
));
TableHeaderCell.displayName = "TableHeaderCell";

export const TableRow = React.forwardRef<
  HTMLTableRowElement,
  React.HTMLProps<HTMLTableRowElement>
>(({ children, className, ...props }, ref) => (
  <tr
    ref={ref}
    {...props}
    className={clsx(className, "hover:bg-gray-2 dark:hover:bg-gray-dark-2")}
  >
    {children}
  </tr>
));
TableRow.displayName = "TableRow";
