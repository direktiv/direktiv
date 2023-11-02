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

import { Card } from "~/design/Card";
import { Network } from "lucide-react";
import { Row } from "./Row";
import { useGatewayList } from "~/api/gateway/query/get";
import { useTranslation } from "react-i18next";

const GatewayTable = () => {
  const { t } = useTranslation();
  const {
    data: gatewayList,
    isSuccess,
    isAllowed,
    noPermissionMessage,
  } = useGatewayList();

  const noResults = isSuccess && gatewayList.data.length === 0;
  return (
    <Card>
      <Table>
        <TableHead>
          <TableRow className="hover:bg-inherit dark:hover:bg-inherit">
            <TableHeaderCell>
              {t("pages.gateway.columns.filePath")}
            </TableHeaderCell>
            <TableHeaderCell className="w-32">
              {t("pages.gateway.columns.method")}
            </TableHeaderCell>
            <TableHeaderCell className="w-32">
              {t("pages.gateway.columns.plugins")}
            </TableHeaderCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {isAllowed ? (
            <>
              {noResults ? (
                <TableRow className="hover:bg-inherit dark:hover:bg-inherit">
                  <TableCell colSpan={3}>
                    <NoResult icon={Network}>
                      {t("pages.gateway.empty")}
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
              <TableCell colSpan={3}>
                <NoPermissions>{noPermissionMessage}</NoPermissions>
              </TableCell>
            </TableRow>
          )}
        </TableBody>
      </Table>
    </Card>
  );
};
export default GatewayTable;
