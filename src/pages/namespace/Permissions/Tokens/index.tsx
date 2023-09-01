import { KeyRound, Users } from "lucide-react";
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
import { useTokens } from "~/api/enterprise/tokens/query/get";
import { useTranslation } from "react-i18next";

const TokensPage = () => {
  const { t } = useTranslation();
  const { data, isFetched } = useTokens();
  const noResults = isFetched && data?.tokens.length === 0;

  return (
    <Card className="m-5">
      <Table>
        <TableHead>
          <TableRow className="hover:bg-inherit dark:hover:bg-inherit">
            <TableHeaderCell className="w-32">
              {t("pages.permissions.tokens.tableHeader.description")}
            </TableHeaderCell>
            <TableHeaderCell>
              {t("pages.permissions.tokens.tableHeader.permissions")}
            </TableHeaderCell>
            <TableHeaderCell className="w-32">
              {t("pages.permissions.tokens.tableHeader.created")}
            </TableHeaderCell>
            <TableHeaderCell className="w-32">
              {t("pages.permissions.tokens.tableHeader.expires")}
            </TableHeaderCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {noResults ? (
            <TableRow className="hover:bg-inherit dark:hover:bg-inherit">
              <TableCell colSpan={3}>
                <NoResult icon={KeyRound}>
                  {t("pages.permissions.tokens.noTokens")}
                </NoResult>
              </TableCell>
            </TableRow>
          ) : (
            data?.tokens.map((token) => <Row key={token.id} token={token} />)
          )}
        </TableBody>
      </Table>
    </Card>
  );
};

export default TokensPage;
