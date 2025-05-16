import { MoreVertical, PackageOpen } from "lucide-react";
import {
  NoResult,
  TableBody,
  TableCell as TableCellDesignComopnent,
  Table as TableDesignComponent,
  TableHead,
  TableHeaderCell,
  TableRow,
} from "~/design/Table";
import {
  VariableContextProvider,
  useVariables,
} from "../../primitives/Variable/VariableContext";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import { TableCell } from "./TableCell";
import { TableType } from "../../../schema/blocks/table";
import { VariableError } from "../../primitives/Variable/Error";
import { useResolveVariableArray } from "../../primitives/Variable/utils/useResolveVariableArray";
import { useTranslation } from "react-i18next";

type TableProps = {
  blockProps: TableType;
};

export const Table = ({ blockProps }: TableProps) => {
  const { columns, actions, data: loop } = blockProps;
  const { t } = useTranslation();
  const arrayVariable = useResolveVariableArray(loop.data);

  const parentVariables = useVariables();

  if (parentVariables.loop[loop.id]) {
    throw new Error(t("direktivPage.error.dublicateId", { id: loop.id }));
  }

  if (!arrayVariable.success) {
    return (
      <VariableError value={loop.data} errorCode={arrayVariable.error}>
        {t(`direktivPage.error.templateString.${arrayVariable.error}`)} (
        {arrayVariable.error})
      </VariableError>
    );
  }

  const hasActionsColumn = actions.length > 0;
  const numberOfColumns = columns.length + (hasActionsColumn ? 1 : 0);
  const noResult = arrayVariable.data.length === 0;

  return (
    <Card>
      <TableDesignComponent>
        <TableHead>
          <TableRow>
            {columns.map((column, index) => (
              <TableHeaderCell key={index}>{column.label}</TableHeaderCell>
            ))}
            {hasActionsColumn && <TableHeaderCell className="w-0" />}
          </TableRow>
        </TableHead>
        <TableBody>
          {noResult && (
            <TableRow>
              <TableCellDesignComopnent colSpan={numberOfColumns}>
                <NoResult icon={PackageOpen}>
                  {arrayVariable.success && arrayVariable.data.length}
                  {t("direktivPage.error.blocks.table.noResult")}
                </NoResult>
              </TableCellDesignComopnent>
            </TableRow>
          )}
          {arrayVariable.data.map((item, index) => (
            <VariableContextProvider
              key={index}
              value={{
                ...parentVariables,
                loop: {
                  ...parentVariables.loop,
                  [loop.id]: item,
                },
              }}
            >
              <TableRow key={index}>
                {columns.map((column, columnIndex) => (
                  <TableCell key={columnIndex} blockProps={column} />
                ))}
                {hasActionsColumn && (
                  <TableHeaderCell>
                    <Button variant="ghost" size="sm" icon>
                      <MoreVertical />
                    </Button>
                  </TableHeaderCell>
                )}
              </TableRow>
            </VariableContextProvider>
          ))}
        </TableBody>
      </TableDesignComponent>
    </Card>
  );
};
