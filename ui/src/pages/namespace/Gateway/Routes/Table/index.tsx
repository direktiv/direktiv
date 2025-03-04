import {
  NoPermissions,
  NoResult,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeaderCell,
  TableRow,
} from "~/design/Table";

import { Network } from "lucide-react";
import { RouteSchemaType } from "~/api/gateway/schema";
import { Row } from "./Row";
import { useRoutes } from "~/api/gateway/query/getRoutes";
import { useTranslation } from "react-i18next";

const RoutesTable = ({
  search,
  filteredRoutes,
}: {
  search: string;
  filteredRoutes: RouteSchemaType[];
}) => {
  const { t } = useTranslation();
  const {
    data: routes,
    isSuccess,
    isAllowed,
    noPermissionMessage,
  } = useRoutes();

  const isSearch = search.length > 0;

  const noResults =
    (isSuccess && routes.data.length === 0) ||
    (isSuccess && filteredRoutes.length === 0);

  return (
    <Table className="border-t border-gray-5 dark:border-gray-dark-5">
      <TableHead className="sticky top-0 bg-white dark:bg-gray-dark-1 z-10 border-y border-gray-5 dark:border-gray-dark-5">
        <TableRow className="hover:bg-inherit dark:hover:bg-inherit">
          <TableHeaderCell className="w-60 grow">
            {t("pages.gateway.routes.columns.filePath")}
          </TableHeaderCell>

          <TableHeaderCell className="w-24">
            {t("pages.gateway.routes.columns.plugins")}
          </TableHeaderCell>
          <TableHeaderCell className="grow">
            {t("pages.gateway.routes.columns.path")}
          </TableHeaderCell>
        </TableRow>
      </TableHead>
      <TableBody data-testid="route-table">
        {isAllowed ? (
          <>
            {noResults ? (
              <TableRow className="hover:bg-inherit dark:hover:bg-inherit">
                <TableCell colSpan={5}>
                  <NoResult icon={Network}>
                    {t(
                      isSearch
                        ? "pages.gateway.routes.emptySearch"
                        : "pages.gateway.routes.empty"
                    )}
                  </NoResult>
                </TableCell>
              </TableRow>
            ) : (
              filteredRoutes.map((route) => (
                <Row key={route.file_path} route={route} />
              ))
            )}
          </>
        ) : (
          <TableRow className="hover:bg-inherit dark:hover:bg-inherit">
            <TableCell colSpan={5}>
              <NoPermissions>{noPermissionMessage}</NoPermissions>
            </TableCell>
          </TableRow>
        )}
      </TableBody>
    </Table>
  );
};
export default RoutesTable;
