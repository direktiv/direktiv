import { Table, TableHead, TableHeaderCell, TableRow } from "~/design/Table";

import { Card } from "~/design/Card";
import { GitCompare } from "lucide-react";
import Row from "./Row";
import { useMirrorInfo } from "~/api/tree/query/mirrorInfo";
import { useTranslation } from "react-i18next";

const Activities = () => {
  const { data } = useMirrorInfo();
  const { t } = useTranslation();

  const activities = data?.activities.results;

  if (!activities) return null;

  return (
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
                {t("pages.mirror.activities.tableHeader.status")}
              </TableHeaderCell>
              <TableHeaderCell>
                {t("pages.mirror.activities.tableHeader.type")}
              </TableHeaderCell>
              <TableHeaderCell>
                {t("pages.mirror.activities.tableHeader.id")}
              </TableHeaderCell>
              <TableHeaderCell>
                {t("pages.mirror.activities.tableHeader.createdAt")}
              </TableHeaderCell>
            </TableRow>
          </TableHead>
          {activities.map((activity) => (
            <Row key={activity.id} item={activity} />
          ))}
        </Table>
      </Card>
    </div>
  );
};

export default Activities;
