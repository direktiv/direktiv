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

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import CreateRole from "./Create";
import Delete from "./Delete";
import EditRole from "./Edit";
import { RoleSchemaType } from "~/api/enterprise/roles/schema";
import Row from "./Row";
import { useRoles } from "~/api/enterprise/roles/query/get";
import { useState } from "react";
import { useTranslation } from "react-i18next";

const RolesPage = () => {
  const { t } = useTranslation();
  const { data, isFetched, isAllowed, noPermissionMessage } = useRoles();
  const noResults = isFetched && data?.data.length === 0;
  const [dialogOpen, setDialogOpen] = useState(false);
  const [createRole, setCreateRole] = useState(false);
  const [deleteRole, setDeleteRole] = useState<RoleSchemaType>();
  const [editRole, setEditRole] = useState<RoleSchemaType>();

  const allAvailableNames = data?.data.map((role) => role.name) ?? [];

  const createNewButton = (
    <DialogTrigger asChild>
      <Button onClick={() => setCreateRole(true)} variant="outline">
        <PlusCircle />
        {t("pages.permissions.roles.createBtn")}
      </Button>
    </DialogTrigger>
  );

  const onOpenChange = (openState: boolean) => {
    if (openState === false) {
      setCreateRole(false);
      setDeleteRole(undefined);
      setEditRole(undefined);
    }
    setDialogOpen(openState);
  };

  return (
    <Card className="m-5">
      <Dialog open={dialogOpen} onOpenChange={onOpenChange}>
        <div className="flex justify-end gap-5 p-2">{createNewButton}</div>
        <Table className="border-t border-gray-5 dark:border-gray-dark-5">
          <TableHead>
            <TableRow className="hover:bg-inherit dark:hover:bg-inherit">
              <TableHeaderCell className="w-32">
                {t("pages.permissions.roles.tableHeader.name")}
              </TableHeaderCell>
              <TableHeaderCell>
                {t("pages.permissions.roles.tableHeader.description")}
              </TableHeaderCell>
              <TableHeaderCell>
                {t("pages.permissions.roles.tableHeader.oidcGroups")}
              </TableHeaderCell>
              <TableHeaderCell className="w-36">
                {t("pages.permissions.roles.tableHeader.permissions")}
              </TableHeaderCell>
              <TableHeaderCell className="w-36">
                {t("pages.permissions.roles.tableHeader.createdAt")}
              </TableHeaderCell>
              <TableHeaderCell className="w-16" />
            </TableRow>
          </TableHead>
          <TableBody>
            {isAllowed ? (
              <>
                {noResults ? (
                  <TableRow className="hover:bg-inherit dark:hover:bg-inherit">
                    <TableCell colSpan={5}>
                      <NoResult icon={Users} button={createNewButton}>
                        {t("pages.permissions.roles.noRoles")}
                      </NoResult>
                    </TableCell>
                  </TableRow>
                ) : (
                  data?.data.map((role) => (
                    <Row
                      key={role.name}
                      role={role}
                      onDeleteClicked={setDeleteRole}
                      onEditClicked={setEditRole}
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
        <DialogContent className="sm:max-w-2xl">
          {deleteRole && (
            <Delete group={deleteRole} close={() => onOpenChange(false)} />
          )}
          {editRole && (
            <EditRole
              group={editRole}
              close={() => onOpenChange(false)}
              unallowedNames={allAvailableNames.filter(
                (name) => name !== editRole.name
              )}
            />
          )}
          {createRole && (
            <CreateRole
              close={() => onOpenChange(false)}
              unallowedNames={allAvailableNames}
            />
          )}
        </DialogContent>
      </Dialog>
    </Card>
  );
};

export default RolesPage;
