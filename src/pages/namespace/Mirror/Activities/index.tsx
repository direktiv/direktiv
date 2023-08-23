import {
  Table,
  TableBody,
  TableHead,
  TableHeaderCell,
  TableRow,
} from "~/design/Table";

import { Card } from "~/design/Card";
import { GitCompare } from "lucide-react";
import Header from "./Header";
import Row from "./Row";
import { treeKeys } from "~/api/tree";
import { useApiKey } from "~/util/store/apiKey";
import { useMirrorInfo } from "~/api/tree/query/mirrorInfo";
import { useQueryClient } from "@tanstack/react-query";
import { useTranslation } from "react-i18next";

const Activities = () => {
  const { data } = useMirrorInfo();
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const apiKey = useApiKey();

  const activities = data?.activities.results;

  if (!activities) return null;

  const refreshActivities = () => {
    queryClient.invalidateQueries(
      treeKeys.mirrorInfo(data.namespace, {
        apiKey: apiKey ?? undefined,
      })
    );
  };

  const pendingActivities = activities.filter(
    (activity) => activity.status === "executing"
  );

  if (pendingActivities.length) {
    setTimeout(() => refreshActivities(), 1000);
  }

  return (
    <>
      <Header mirror={data} />

      <div className="flex grow flex-col gap-y-4 p-5">
        <h3 className="flex items-center gap-x-2 font-bold">
          <GitCompare className="h-5" />
          {t("pages.mirror.activities.list.title")}
        </h3>
        <Card>
          <Table className="border-gray-5 dark:border-gray-dark-5">
            <TableHead>
              <TableRow className="hover:bg-inherit dark:hover:bg-inherit">
                <TableHeaderCell>
                  {t("pages.mirror.activities.tableHeader.id")}
                </TableHeaderCell>
                <TableHeaderCell>
                  {t("pages.mirror.activities.tableHeader.type")}
                </TableHeaderCell>
                <TableHeaderCell>
                  {t("pages.mirror.activities.tableHeader.status")}
                </TableHeaderCell>
                <TableHeaderCell>
                  {t("pages.mirror.activities.tableHeader.createdAt")}
                </TableHeaderCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {activities.map((activity) => (
                <Row
                  namespace={data.namespace}
                  key={activity.id}
                  item={activity}
                />
              ))}
            </TableBody>
          </Table>
        </Card>
      </div>
    </>
  );
};

export default Activities;
