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
  useVariablesContext,
} from "../../primitives/Variable/VariableContext";

import { ActionsCell } from "./ActionsCell";
import { Card } from "~/design/Card";
import { PackageOpen } from "lucide-react";
import { TableCell } from "./TableCell";
import { TableType } from "../../../schema/blocks/table";
import { VariableError } from "../../primitives/Variable/Error";
import { useTranslation } from "react-i18next";
import { useVariableArrayResolver } from "../../primitives/Variable/utils/useVariableArrayResolver";

type TableProps = {
  blockProps: TableType;
};

export const Table = ({ blockProps }: TableProps) => {
  const { columns, actions, data: loop } = blockProps;
  const { t } = useTranslation();
  const resolveVariableArray = useVariableArrayResolver();
  const parentVariables = useVariablesContext();

  const variableArray = resolveVariableArray(loop.data);

  if (parentVariables.loop?.[loop.id]) {
    throw new Error(t("direktivPage.error.duplicateId", { id: loop.id }));
  }

  if (!variableArray.success) {
    return (
      <VariableError value={loop.data} errorCode={variableArray.error}>
        {t(`direktivPage.error.templateString.${variableArray.error}`, {
          variable: loop.data,
        })}{" "}
        ({variableArray.error})
      </VariableError>
    );
  }

  const hasActionsColumn = actions.length > 0;
  const numberOfColumns = columns.length + (hasActionsColumn ? 1 : 0);
  const hasRows = variableArray.data.length > 0;

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
          {hasRows ? (
            variableArray.data.map((item, index) => (
              <VariableContextProvider
                key={index}
                variables={{
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
                  {t("direktivPage.page.blocks.table.noResult")}
                </NoResult>
              </TableCellDesignComponent>
            </TableRow>
          )}
        </TableBody>
      </TableDesignComponent>
    </Card>
  );
};
