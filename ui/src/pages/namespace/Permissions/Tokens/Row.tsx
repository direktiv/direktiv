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
  onDeleteClicked: (tokenName: string) => void;
}) => {
  const { t } = useTranslation();
  const createdAt = useUpdatedAt(token.createdAt);
  const expiresAt = useUpdatedAt(token.expiredAt);
  return (
    <TooltipProvider>
      <TableRow className="hover:bg-inherit dark:hover:bg-inherit">
        <TableCell>{token.name}</TableCell>
        <TableCell>{token.description}</TableCell>
        <TableCell>
          <Tooltip>
            <TooltipTrigger>
              {t("pages.permissions.tokens.created", {
                relativeTime: createdAt,
              })}
            </TooltipTrigger>
            <TooltipContent>{token.createdAt}</TooltipContent>
          </Tooltip>
        </TableCell>
        <TableCell>
          <Badge variant={token.isExpired ? "destructive" : "outline"}>
            <Tooltip>
              <TooltipTrigger>
                {token.isExpired
                  ? t("pages.permissions.tokens.expiredAgo", {
                      relativeTime: expiresAt,
                    })
                  : t("pages.permissions.tokens.expiresIn", {
                      relativeTime: expiresAt,
                    })}
              </TooltipTrigger>
              <TooltipContent>{token.isExpired}</TooltipContent>
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
                  onDeleteClicked(token.name);
                }}
              >
                <DropdownMenuItem>
                  <Trash className="mr-2 size-4" />
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
