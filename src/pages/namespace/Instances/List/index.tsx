import {
  Table,
  TableBody,
  TableHead,
  TableHeaderCell,
  TableRow,
} from "~/design/Table";

import { Boxes } from "lucide-react";
import { Card } from "~/design/Card";
import NoResult from "./NoResult";
import Row from "./Row";
import { useInstances } from "~/api/instances/query/get";
import { useTranslation } from "react-i18next";

const InstancesListPage = () => {
  const { data, isFetched } = useInstances({ limit: 10, offset: 0 });
  const { t } = useTranslation();

  if (!isFetched) return null;
  const noResults = data?.instances.results.length === 0 && isFetched;

  return (
    <div className="flex grow flex-col gap-y-4 p-5">
      <h3 className="flex items-center gap-x-2 font-bold text-gray-10 dark:text-gray-dark-10">
        <Boxes className="h-5" />
        {t("pages.instances.list.title")}
      </h3>
      <Card>
        {noResults ? (
          <NoResult />
        ) : (
          <Table>
            <TableHead>
              <TableRow>
                <TableHeaderCell>
                  {t("pages.instances.list.tableHeader.name")}
                </TableHeaderCell>
                <TableHeaderCell>
                  {t("pages.instances.list.tableHeader.id")}
                </TableHeaderCell>
                <TableHeaderCell>
                  {t("pages.instances.list.tableHeader.revisionId")}
                </TableHeaderCell>
                <TableHeaderCell>
                  {t("pages.instances.list.tableHeader.invoker")}
                </TableHeaderCell>
                <TableHeaderCell>
                  {t("pages.instances.list.tableHeader.state")}
                </TableHeaderCell>
                <TableHeaderCell>
                  {t("pages.instances.list.tableHeader.startedAt")}
                </TableHeaderCell>
                <TableHeaderCell>
                  {t("pages.instances.list.tableHeader.updatedAt")}
                </TableHeaderCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {data?.instances.results.map((instance) => (
                <Row
                  instance={instance}
                  key={instance.id}
                  namespace={data.namespace}
                />
              ))}
            </TableBody>
          </Table>
        )}
      </Card>
    </div>
  );
};

export default InstancesListPage;
