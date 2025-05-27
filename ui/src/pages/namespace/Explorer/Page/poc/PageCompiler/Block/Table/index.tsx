import {
  NoResult,
  TableBody,
  TableCell as TableCellDesignComponent,
  Table as TableDesignComponent,
  TableHead,
  TableHeaderCell,
  TableRow,
} from "~/design/Table";
import {
  VariableContextProvider,
  useVariables,
} from "../../primitives/Variable/VariableContext";

import { ActionsCell } from "./ActionsCell";
import { Card } from "~/design/Card";
import { PackageOpen } from "lucide-react";
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
    throw new Error(t("direktivPage.error.duplicateId", { id: loop.id }));
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
  const hasResults = arrayVariable.data.length > 0;

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
          {hasResults ? (
            arrayVariable.data.map((item, index) => (
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
                <TableRow>
                  {columns.map((column, columnIndex) => (
                    <TableCell key={columnIndex} blockProps={column} />
                  ))}
                  {hasActionsColumn && <ActionsCell actions={actions} />}
                </TableRow>
              </VariableContextProvider>
            ))
          ) : (
            <TableRow>
              <TableCellDesignComponent colSpan={numberOfColumns}>
                <NoResult icon={PackageOpen}>
                  {t("direktivPage.error.blocks.table.noResult")}
                </NoResult>
              </TableCellDesignComponent>
            </TableRow>
          )}
        </TableBody>
      </TableDesignComponent>
    </Card>
  );
};
