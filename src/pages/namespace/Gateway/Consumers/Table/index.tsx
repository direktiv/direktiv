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
import { useConsumers } from "~/api/gateway/query/getConsumers";
import { useTranslation } from "react-i18next";

const ConsumerTable = () => {
  const { t } = useTranslation();
  const {
    data: consumerList,
    isSuccess,
    isAllowed,
    noPermissionMessage,
  } = useConsumers();

  const noResults = isSuccess && consumerList.data.length === 0;

  return (
    <Table className="border-t border-gray-5 dark:border-gray-dark-5">
      <TableHead>
        <TableRow className="hover:bg-inherit dark:hover:bg-inherit">
          <TableHeaderCell>
            {t("pages.gateway.consumer.columns.username")}
          </TableHeaderCell>
          <TableHeaderCell className="w-52">
            {t("pages.gateway.consumer.columns.password")}
          </TableHeaderCell>
          <TableHeaderCell className="w-52">
            {t("pages.gateway.consumer.columns.apikey")}
          </TableHeaderCell>
          <TableHeaderCell className="w-[200px]">
            {t("pages.gateway.consumer.columns.groups")}
          </TableHeaderCell>
          <TableHeaderCell className="w-[200px]">
            {t("pages.gateway.consumer.columns.tags")}
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
                    {t("pages.gateway.consumer.empty")}
                  </NoResult>
                </TableCell>
              </TableRow>
            ) : (
              consumerList?.data?.map((consumer) => (
                <Row key={consumer.username} consumer={consumer} />
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
export default ConsumerTable;
