import { FC, PropsWithChildren } from "react";
import { TableHead, TableHeaderCell, TableRow } from "~/design/Table";

type TableHeaderProps = PropsWithChildren & {
  title: string;
};

export const TableHeader: FC<TableHeaderProps> = ({ title, children }) => (
  <TableHead>
    <TableRow className="hover:bg-inherit dark:hover:bg-inherit">
      <TableHeaderCell>{title}</TableHeaderCell>
      <TableHeaderCell className="w-60 text-right">{children}</TableHeaderCell>
    </TableRow>
  </TableHead>
);
