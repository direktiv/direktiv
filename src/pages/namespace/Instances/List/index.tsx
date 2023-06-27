import { Pagination, PaginationLink } from "~/design/Pagination";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeaderCell,
  TableRow,
} from "~/design/Table";

import { Boxes } from "lucide-react";
import { Card } from "~/design/Card";
import NoResult from "./NoResult";
import Row from "./Row";
import { useInstances } from "~/api/instances/query/get";
import { useState } from "react";
import { useTranslation } from "react-i18next";

const instancesPerPage = 10;

const InstancesListPage = () => {
  const [offset, setOffset] = useState(0);
  const { data, isFetched } = useInstances({
    limit: instancesPerPage,
    offset,
  });

  const { total } = data?.instances?.pageInfo ?? {};

  const pages = Math.ceil((total ?? 0) / instancesPerPage);
  const currentPage = Math.ceil(offset / instancesPerPage) + 1;

  const { t } = useTranslation();

  const noResults = isFetched && data?.instances.results.length === 0;

  return (
    <div className="flex grow flex-col gap-y-4 p-5">
      <h3 className="flex items-center gap-x-2 font-bold text-gray-10 dark:text-gray-dark-10">
        <Boxes className="h-5" />
        {t("pages.instances.list.title")}
      </h3>
      <Card>
        <Table>
          <TableHead>
            <TableRow>
              <TableHeaderCell>
                {t("pages.instances.list.tableHeader.name")}
              </TableHeaderCell>
              <TableHeaderCell className="w-32">
                {t("pages.instances.list.tableHeader.id")}
              </TableHeaderCell>
              <TableHeaderCell className="w-28">
                {t("pages.instances.list.tableHeader.revisionId")}
              </TableHeaderCell>
              <TableHeaderCell className="w-28">
                {t("pages.instances.list.tableHeader.invoker")}
              </TableHeaderCell>
              <TableHeaderCell className="w-28">
                {t("pages.instances.list.tableHeader.state")}
              </TableHeaderCell>
              <TableHeaderCell className="w-40">
                {t("pages.instances.list.tableHeader.startedAt")}
              </TableHeaderCell>
              <TableHeaderCell className="w-40">
                {t("pages.instances.list.tableHeader.updatedAt")}
              </TableHeaderCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {noResults ? (
              <TableRow className="hover:bg-inherit dark:hover:bg-inherit">
                <TableCell colSpan={7}>
                  <NoResult />
                </TableCell>
              </TableRow>
            ) : (
              data?.instances.results.map((instance) => (
                <Row
                  instance={instance}
                  key={instance.id}
                  namespace={data.namespace}
                />
              ))
            )}
          </TableBody>
        </Table>
      </Card>
      {pages > 0 && (
        <Pagination>
          <PaginationLink icon="left" onClick={() => setOffset(0)} />
          {[...Array(pages)].map((_, i) => {
            const pageNumber = i + 1;
            const isActive = currentPage === pageNumber;
            return (
              <PaginationLink
                key={i}
                active={isActive}
                onClick={() => {
                  isActive
                    ? null
                    : setOffset((pageNumber - 1) * instancesPerPage);
                }}
              >
                {pageNumber}
              </PaginationLink>
            );
          })}
          <PaginationLink
            icon="right"
            onClick={() => setOffset((pages - 1) * instancesPerPage)}
          />
        </Pagination>
      )}
    </div>
  );
};

export default InstancesListPage;
