import { TableCell, TableRow } from "~/design/Table";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "~/design/Tooltip";

import Badge from "~/design/Badge";
import PermissionsInfo from "../components/PermissionsInfo";
import { TokenSchemaType } from "~/api/enterprise/tokens/schema";
import { useTranslation } from "react-i18next";
import useUpdatedAt from "~/hooksNext/useUpdatedAt";

const Row = ({ token }: { token: TokenSchemaType }) => {
  const { t } = useTranslation();
  const createdAt = useUpdatedAt(token.created);
  const expiresAt = useUpdatedAt(token.expires);
  return (
    <TooltipProvider>
      <TableRow className="hover:bg-inherit dark:hover:bg-inherit">
        <TableCell>{token.description}</TableCell>
        <TableCell>
          <Tooltip>
            <TooltipTrigger>
              {t("pages.permissions.tokens.created", {
                relativeTime: createdAt,
              })}
            </TooltipTrigger>
            <TooltipContent>{token.created}</TooltipContent>
          </Tooltip>
        </TableCell>
        <TableCell>
          <Badge variant={token.expired ? "destructive" : "outline"}>
            <Tooltip>
              <TooltipTrigger>
                {token.expired
                  ? t("pages.permissions.tokens.expiredAgo", {
                      relativeTime: expiresAt,
                    })
                  : t("pages.permissions.tokens.expiresIn", {
                      relativeTime: expiresAt,
                    })}
              </TooltipTrigger>
              <TooltipContent>{token.expires}</TooltipContent>
            </Tooltip>
          </Badge>
        </TableCell>
        <TableCell>
          <PermissionsInfo permissions={token.permissions} />
        </TableCell>
      </TableRow>
    </TooltipProvider>
  );
};

export default Row;
