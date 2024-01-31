import { FC, PropsWithChildren } from "react";
import {
  TableCell,
  TableHead,
  TableHeaderCell,
  TableRow,
} from "~/design/Table";

import { ContextMenu } from "~/pages/namespace/Explorer/Endpoint/EndpointEditor/Form/plugins/components/PluginsTable";
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
  onDelete: () => void;
  onMoveUp?: () => void;
  onMoveDown?: () => void;
};

export const PatchRow: FC<PatchRowProps> = ({
  patch,
  onClick,
  onDelete,
  onMoveUp,
  onMoveDown,
}) => (
  <TableRow onClick={onClick}>
    <TableCell>{patch.op}</TableCell>
    <TableCell>{patch.path}</TableCell>
    <TableCell className="text-right">
      {/* TODO: Context menu should be generic (imported from plugin context) */}
      <ContextMenu
        onDelete={onDelete}
        onMoveDown={onMoveDown}
        onMoveUp={onMoveUp}
      />
    </TableCell>
  </TableRow>
);
