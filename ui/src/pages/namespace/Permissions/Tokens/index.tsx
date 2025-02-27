import { Dialog, DialogContent, DialogTrigger } from "~/design/Dialog";
import { KeyRound, PlusCircle } from "lucide-react";
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

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import CreateToken from "./Create";
import Delete from "./Delete";
import Row from "./Row";
import { useState } from "react";
import { useTokens } from "~/api/enterprise/tokens/query/get";
import { useTranslation } from "react-i18next";

const TokensPage = () => {
  const { t } = useTranslation();
  const { data, isFetched, isAllowed, noPermissionMessage } = useTokens();
  const noResults = isFetched && data?.data.length === 0;
  const [dialogOpen, setDialogOpen] = useState(false);
  const [createToken, setCreateToken] = useState(false);
  const [deleteToken, setDeleteToken] = useState<string>();

  const onOpenChange = (openState: boolean) => {
    if (openState === false) {
      setCreateToken(false);
      setDeleteToken(undefined);
    }
    setDialogOpen(openState);
  };

  const createNewButton = (
    <DialogTrigger asChild>
      <Button onClick={() => setCreateToken(true)} variant="outline">
        <PlusCircle />
        {t("pages.permissions.tokens.createBtn")}
      </Button>
    </DialogTrigger>
  );

  return (
    <Card className="m-5">
      <Dialog open={dialogOpen} onOpenChange={onOpenChange}>
        <div className="flex justify-end gap-5 p-2">{createNewButton}</div>
        <Table className="border-t border-gray-5 dark:border-gray-dark-5">
          <TableHead>
            <TableRow className="hover:bg-inherit dark:hover:bg-inherit">
              <TableHeaderCell>
                {t("pages.permissions.tokens.tableHeader.name")}
              </TableHeaderCell>
              <TableHeaderCell>
                {t("pages.permissions.tokens.tableHeader.description")}
              </TableHeaderCell>
              <TableHeaderCell>
                {t("pages.permissions.tokens.tableHeader.prefix")}
              </TableHeaderCell>
              <TableHeaderCell className="w-40">
                {t("pages.permissions.tokens.tableHeader.expires")}
              </TableHeaderCell>
              <TableHeaderCell className="w-36">
                {t("pages.permissions.tokens.tableHeader.permissions")}
              </TableHeaderCell>
              <TableHeaderCell className="w-40">
                {t("pages.permissions.tokens.tableHeader.created")}
              </TableHeaderCell>
              <TableHeaderCell className="w-16" />
            </TableRow>
          </TableHead>
          <TableBody>
            {isAllowed ? (
              <>
                {noResults ? (
                  <TableRow className="hover:bg-inherit dark:hover:bg-inherit">
                    <TableCell colSpan={6}>
                      <NoResult icon={KeyRound} button={createNewButton}>
                        {t("pages.permissions.tokens.noTokens")}
                      </NoResult>
                    </TableCell>
                  </TableRow>
                ) : (
                  data?.data.map((token) => (
                    <Row
                      key={token.name}
                      token={token}
                      onDeleteClicked={setDeleteToken}
                    />
                  ))
                )}
              </>
            ) : (
              <TableRow className="hover:bg-inherit dark:hover:bg-inherit">
                <TableCell colSpan={6}>
                  <NoPermissions>{noPermissionMessage}</NoPermissions>
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
        <DialogContent className="sm:max-w-2xl">
          {deleteToken && (
            <Delete tokenName={deleteToken} close={() => onOpenChange(false)} />
          )}
          {createToken && (
            <CreateToken
              close={() => onOpenChange(false)}
              unallowedNames={data?.data.map((token) => token.name)}
            />
          )}
        </DialogContent>
      </Dialog>
    </Card>
  );
};

export default TokensPage;
