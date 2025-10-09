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

import { BlockPathType } from "..";
import { Card } from "~/design/Card";
import { PackageOpen } from "lucide-react";
import { Pagination } from "~/components/Pagination";
import PaginationProvider from "~/components/PaginationProvider";
import { RowActions } from "./actions/RowActions";
import { StopPropagation } from "~/components/StopPropagation";
import { TableActions } from "./actions/TableActions";
import { TableCell } from "./TableCell";
import { TableType } from "../../../schema/blocks/table";
import { VariableError } from "../../primitives/Variable/Error";
import { useTranslation } from "react-i18next";
import { useVariableArrayResolver } from "../../primitives/Variable/utils/useVariableArrayResolver";

type TableProps = {
  blockProps: TableType;
  blockPath: BlockPathType;
};

export const Table = ({ blockProps, blockPath }: TableProps) => {
  const { columns, blocks, data: loop } = blockProps;
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

  const hasRowActions = blocks[1].blocks.length > 0;
  const hasTableActions = blocks[0].blocks.length > 0;
  const numberOfHeaderColumns = columns.length + (hasTableActions ? 1 : 0);
  const hasRows = variableArray.data.length > 0;

  return (
    <>
      <PaginationProvider items={variableArray.data} pageSize={loop.pageSize}>
        {({
          currentItems,
          goToPage,
          goToFirstPage,
          currentPage,
          totalPages,
        }) => {
          currentPage > totalPages && goToFirstPage();
          return (
            <>
              <Card>
                <TableDesignComponent>
                  <TableHead>
                    <TableRow>
                      {columns.map((column, index) => (
                        <TableHeaderCell key={index}>
                          {column.label}
                        </TableHeaderCell>
                      ))}
                      {hasTableActions && (
                        <TableHeaderCell className="w-0">
                          <TableActions
                            actions={blocks[0]}
                            blockPath={blockPath}
                          />
                        </TableHeaderCell>
                      )}
                    </TableRow>
                  </TableHead>
                  <TableBody>
                    {hasRows ? (
                      currentItems.map((item, index) => (
                        <VariableContextProvider
                          key={`${currentItems.length}-${index}`}
                          variables={{
                            ...parentVariables,
                            loop: {
                              ...parentVariables.loop,
                              [loop.id]: item,
                            },
                          }}
                        >
                          <TableRow>
                            {columns.map((column, columnIndex) => {
                              const isLastColumn =
                                columnIndex === columns.length - 1;
                              return (
                                <TableCell
                                  key={columnIndex}
                                  blockProps={column}
                                  colSpan={
                                    isLastColumn &&
                                    hasTableActions &&
                                    !hasRowActions
                                      ? 2
                                      : 1
                                  }
                                />
                              );
                            })}
                            {hasRowActions && (
                              <RowActions
                                actions={blocks[1]}
                                blockPath={blockPath}
                              />
                            )}
                          </TableRow>
                        </VariableContextProvider>
                      ))
                    ) : (
                      <TableRow>
                        <TableCellDesignComponent
                          colSpan={numberOfHeaderColumns}
                        >
                          <NoResult icon={PackageOpen}>
                            {t("direktivPage.page.blocks.table.noResult")}
                          </NoResult>
                        </TableCellDesignComponent>
                      </TableRow>
                    )}
                  </TableBody>
                </TableDesignComponent>
              </Card>
              <div className="flex items-center justify-end gap-2">
                <StopPropagation>
                  <div className="mt-2 pb-4">
                    <Pagination
                      totalPages={totalPages}
                      value={currentPage}
                      onChange={goToPage}
                    />
                  </div>
                </StopPropagation>
              </div>
            </>
          );
        }}
      </PaginationProvider>
    </>
  );
};
