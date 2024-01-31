import { FC, PropsWithChildren } from "react";
import { TableHead, TableHeaderCell, TableRow } from "~/design/Table";

import { PatchSchemaType } from "../../schema";

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

type PatchRowProps = {
  patch: PatchSchemaType;
  onClick: () => void;
};

export const PatchRow: FC<PatchRowProps> = ({ patch, onClick }) => (
  <TableRow onClick={onClick}>
    {patch.op}: {patch.path}
  </TableRow>
);
