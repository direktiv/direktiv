import { Dialog, DialogContent, DialogTrigger } from "~/design/Dialog";
import {
  NoResult,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeaderCell,
  TableRow,
} from "~/design/Table";
import { PlusCircle, Users } from "lucide-react";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import CreateGroup from "./Create";
import Row from "./Row";
import { useGroups } from "~/api/enterprise/groups/query/get";
import { useState } from "react";
import { useTranslation } from "react-i18next";

const GroupsPage = () => {
  const { t } = useTranslation();
  const { data, isFetched } = useGroups();
  const noResults = true; //isFetched && data?.groups.length === 0;
  const [dialogOpen, setDialogOpen] = useState(false);
  const [createToken, setCreateToken] = useState(false);

  const allAvailableNames = data?.groups.map((group) => group.group) ?? [];

  const createNewButton = (
    <DialogTrigger asChild>
      <Button onClick={() => setCreateToken(true)} variant="outline">
        <PlusCircle />
        {t("pages.permissions.groups.createBtn")}
      </Button>
    </DialogTrigger>
  );

  return (
    <Card className="m-5">
      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <div className="flex justify-end gap-5 p-2">{createNewButton}</div>
        <Table className="border-t border-gray-5 dark:border-gray-dark-5">
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
                  <NoResult icon={Users} button={createNewButton}>
                    {t("pages.permissions.groups.noGroups")}
                  </NoResult>
                </TableCell>
              </TableRow>
            ) : (
              data?.groups.map((group) => <Row key={group.id} group={group} />)
            )}
          </TableBody>
        </Table>
        <DialogContent>
          {createToken && (
            <CreateGroup
              close={() => setDialogOpen(false)}
              unallowedNames={allAvailableNames}
            />
          )}
        </DialogContent>
      </Dialog>
    </Card>
  );
};

export default GroupsPage;
