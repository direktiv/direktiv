import {
  NoResult,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeaderCell,
  TableRow,
} from "~/design/Table";

import { Card } from "~/design/Card";
import Row from "./Row";
import { Users } from "lucide-react";
import { useGroups } from "~/api/enterprise/groups/query/get";
import { useTranslation } from "react-i18next";

const GroupsPage = () => {
  const { t } = useTranslation();
  const { data, isFetched } = useGroups();
  const noResults = isFetched && data?.groups.length === 0;

  return (
    <Card className="m-5">
      <Table>
        <TableHead>
          <TableRow className="hover:bg-inherit dark:hover:bg-inherit">
            <TableHeaderCell className="w-32">
              {t("pages.permissions.groups.tableHeader.name")}
            </TableHeaderCell>
            <TableHeaderCell>
              {t("pages.permissions.groups.tableHeader.description")}
            </TableHeaderCell>
            <TableHeaderCell className="w-36">
              {t("pages.permissions.groups.tableHeader.permissions")}
            </TableHeaderCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {noResults ? (
            <TableRow className="hover:bg-inherit dark:hover:bg-inherit">
              <TableCell colSpan={3}>
                <NoResult icon={Users}>
                  {t("pages.permissions.groups.noGroups")}
                </NoResult>
              </TableCell>
            </TableRow>
          ) : (
            data?.groups.map((group) => <Row key={group.id} group={group} />)
          )}
        </TableBody>
      </Table>
    </Card>
  );
};

export default GroupsPage;
