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
import { Pagination, PaginationLink } from "~/design/Pagination";

import { Card } from "~/design/Card";
import { GitCompare } from "lucide-react";
import Header from "./Header";
import PaginationProvider from "~/components/PaginationProvider";
import Row from "./Row";
import { treeKeys } from "~/api/tree";
import { useApiKey } from "~/util/store/apiKey";
import { useMirrorInfo } from "~/api/tree/query/mirrorInfo";
import { useQueryClient } from "@tanstack/react-query";
import { useTranslation } from "react-i18next";

const pageSize = 10;

const Activities = () => {
  const { data, isAllowed, noPermissionMessage, isFetched } = useMirrorInfo();
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const apiKey = useApiKey();

  const activities = data?.activities.results;
  const noResults = isFetched && data?.activities.results.length === 0;

  if (!isAllowed)
    return (
      <Card className="m-5 flex grow flex-col p-4">
        <NoPermissions>{noPermissionMessage}</NoPermissions>
      </Card>
    );

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
      <Header mirror={data} loading={!!pendingActivities.length} />
      <PaginationProvider items={activities} pageSize={pageSize}>
        {({
          currentItems,
          goToPage,
          goToNextPage,
          goToPreviousPage,
          currentPage,
          pagesList,
          totalPages,
        }) => (
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
                  {noResults && (
                    <TableRow className="hover:bg-inherit dark:hover:bg-inherit">
                      <TableCell colSpan={4}>
                        <NoResult icon={GitCompare}>
                          {t("pages.mirror.activities.list.noResults")}
                        </NoResult>
                      </TableCell>
                    </TableRow>
                  )}
                  {currentItems.map((activity) => (
                    <Row
                      namespace={data.namespace}
                      key={activity.id}
                      item={activity}
                    />
                  ))}
                </TableBody>
              </Table>
            </Card>
            {totalPages > 1 && (
              <Pagination>
                <PaginationLink
                  icon="left"
                  onClick={() => goToPreviousPage()}
                />
                {pagesList.map((page) => (
                  <PaginationLink
                    active={currentPage === page}
                    key={`${page}`}
                    onClick={() => goToPage(page)}
                  >
                    {page}
                  </PaginationLink>
                ))}
                <PaginationLink icon="right" onClick={() => goToNextPage()} />
              </Pagination>
            )}
          </div>
        )}
      </PaginationProvider>
    </>
  );
};

export default Activities;
