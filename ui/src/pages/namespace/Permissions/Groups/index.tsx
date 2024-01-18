import { Dialog, DialogContent, DialogTrigger } from "~/design/Dialog";
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
import { PlusCircle, Users } from "lucide-react";
import { useEffect, useState } from "react";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import CreateGroup from "./Create";
import Delete from "./Delete";
import EditGroup from "./Edit";
import { GroupSchemaType } from "~/api/enterprise/groups/schema";
import Row from "./Row";
import { useGroups } from "~/api/enterprise/groups/query/get";
import { useTranslation } from "react-i18next";

const GroupsPage = () => {
  const { t } = useTranslation();
  const { data, isFetched, isAllowed, noPermissionMessage } = useGroups();
  const noResults = isFetched && data?.groups.length === 0;
  const [dialogOpen, setDialogOpen] = useState(false);
  const [createGroup, setCreateGroup] = useState(false);
  const [deleteGroup, setDeleteGroup] = useState<GroupSchemaType>();
  const [editGroup, setEditGroup] = useState<GroupSchemaType>();

  const allAvailableNames = data?.groups.map((group) => group.group) ?? [];

  useEffect(() => {
    if (dialogOpen === false) {
      setCreateGroup(false);
      setDeleteGroup(undefined);
      setEditGroup(undefined);
    }
  }, [dialogOpen]);

  const createNewButton = (
    <DialogTrigger asChild>
      <Button onClick={() => setCreateGroup(true)} variant="outline">
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
              <TableHeaderCell className="w-16" />
            </TableRow>
          </TableHead>
          <TableBody>
            {isAllowed ? (
              <>
                {noResults ? (
                  <TableRow className="hover:bg-inherit dark:hover:bg-inherit">
                    <TableCell colSpan={3}>
                      <NoResult icon={Users} button={createNewButton}>
                        {t("pages.permissions.groups.noGroups")}
                      </NoResult>
                    </TableCell>
                  </TableRow>
                ) : (
                  data?.groups.map((group) => (
                    <Row
                      key={group.id}
                      group={group}
                      onDeleteClicked={setDeleteGroup}
                      onEditClicked={setEditGroup}
                    />
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
        <DialogContent className="sm:max-w-2xl md:max-w-3xl">
          {deleteGroup && (
            <Delete group={deleteGroup} close={() => setDialogOpen(false)} />
          )}
          {editGroup && (
            <EditGroup
              group={editGroup}
              close={() => setDialogOpen(false)}
              unallowedNames={allAvailableNames.filter(
                (name) => name !== editGroup.group
              )}
            />
          )}
          {createGroup && (
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
