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
import { Row } from "./Row";
import { useRoutes } from "~/api/gateway/query/getRoutes";
import { useTranslation } from "react-i18next";

const RoutesTable = () => {
  const { t } = useTranslation();
  const {
    data: gatewayList,
    isSuccess,
    isAllowed,
    noPermissionMessage,
  } = useRoutes();

  const noResults = isSuccess && gatewayList.data.length === 0;

  return (
    <Table className="border-t border-gray-5 dark:border-gray-dark-5">
      <TableHead>
        <TableRow className="hover:bg-inherit dark:hover:bg-inherit">
          <TableHeaderCell>
            {t("pages.gateway.routes.columns.filePath")}
          </TableHeaderCell>
          <TableHeaderCell className="w-32">
            {t("pages.gateway.routes.columns.methods")}
          </TableHeaderCell>
          <TableHeaderCell className="w-52">
            {t("pages.gateway.routes.columns.path")}
          </TableHeaderCell>
          <TableHeaderCell className="w-32">
            {t("pages.gateway.routes.columns.plugins")}
          </TableHeaderCell>
          <TableHeaderCell className="w-40">
            {t("pages.gateway.routes.columns.anonymous")}
          </TableHeaderCell>
        </TableRow>
      </TableHead>
      <TableBody>
        {isAllowed ? (
          <>
            {noResults ? (
              <TableRow className="hover:bg-inherit dark:hover:bg-inherit">
                <TableCell colSpan={5}>
                  <NoResult icon={Network}>
                    {t("pages.gateway.routes.empty")}
                  </NoResult>
                </TableCell>
              </TableRow>
            ) : (
              gatewayList?.data?.map((gateway) => (
                <Row key={gateway.file_path} gateway={gateway} />
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
