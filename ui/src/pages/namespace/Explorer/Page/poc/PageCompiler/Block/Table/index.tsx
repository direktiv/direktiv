import {
  NoResult,
  TableBody,
  TableCell,
  Table as TableDesignComponent,
  TableHead,
  TableHeaderCell,
  TableRow,
} from "~/design/Table";

import { Card } from "~/design/Card";
import { PackageOpen } from "lucide-react";
import { TableType } from "../../../schema/blocks/table";
import { useTranslation } from "react-i18next";

type TableProps = {
  blockProps: TableType;
};

export const Table = ({ blockProps }: TableProps) => {
  const { t } = useTranslation();
  const { columns, actions } = blockProps;

  const hasActionsColumn = actions.length > 0;
  const numberOfColumns = columns.length + (hasActionsColumn ? 1 : 0);

  return (
    <Card>
      <TableDesignComponent>
        <TableHead>
          <TableRow>
            {columns.map((column, index) => (
              <TableHeaderCell key={index}>{column.label}</TableHeaderCell>
            ))}
            {hasActionsColumn && <TableHeaderCell />}
          </TableRow>
        </TableHead>
        <TableBody>
          <TableCell colSpan={numberOfColumns}>
            <NoResult icon={PackageOpen}>
              {t("direktivPage.error.blocks.table.noResult")}
            </NoResult>
          </TableCell>
        </TableBody>
      </TableDesignComponent>
    </Card>
  );
};
