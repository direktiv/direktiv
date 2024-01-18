import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "~/design/Dropdown";
import { MoreVertical, Trash } from "lucide-react";
import { TableCell, TableRow } from "~/design/Table";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "~/design/Tooltip";

import Badge from "~/design/Badge";
import Button from "~/design/Button";
import { DialogTrigger } from "~/design/Dialog";
import PermissionsInfo from "../components/PermissionsInfo";
import { TokenSchemaType } from "~/api/enterprise/tokens/schema";
import { useTranslation } from "react-i18next";
import useUpdatedAt from "~/hooks/useUpdatedAt";

const Row = ({
  token,
  onDeleteClicked,
}: {
  token: TokenSchemaType;
  onDeleteClicked: (group: TokenSchemaType) => void;
}) => {
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
        <TableCell>
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button
                variant="ghost"
                size="sm"
                onClick={(e) => e.preventDefault()}
                icon
              >
                <MoreVertical />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent className="w-40">
              <DialogTrigger
                className="w-full"
                onClick={(e) => {
                  e.stopPropagation();
                  onDeleteClicked(token);
                }}
              >
                <DropdownMenuItem>
                  <Trash className="mr-2 h-4 w-4" />
                  {t("pages.permissions.tokens.contextMenu.delete")}
                </DropdownMenuItem>
              </DialogTrigger>
            </DropdownMenuContent>
          </DropdownMenu>
        </TableCell>
      </TableRow>
    </TooltipProvider>
  );
};

export default Row;
