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
import RefreshButton from "~/design/RefreshButton";
import { Row } from "./Row";
import { useGatewayList } from "~/api/gateway/query/get";
import { useTranslation } from "react-i18next";

const GatewayPage = () => {
  const { t } = useTranslation();
  const {
    data: gatewayList,
    isFetching,
    refetch,
    isSuccess,
    isAllowed,
    noPermissionMessage,
  } = useGatewayList();

  const noResults = isSuccess && gatewayList.data.length === 0;

  return (
    <div className="flex grow flex-col gap-y-4 p-5">
      <div className="flex">
        <h3 className="flex grow items-center gap-x-2 font-bold">
          <Network className="h-5" />
          {t("pages.gateway.title")}
        </h3>
        <RefreshButton
          icon
          variant="outline"
          disabled={isFetching}
          onClick={() => {
            refetch();
          }}
        />
      </div>
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
    </div>
  );
};

export default GatewayPage;
