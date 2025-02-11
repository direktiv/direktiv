import { Table, TableBody } from "~/design/Table";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "~/design/Tooltip";

import Badge from "~/design/Badge";
import { useTranslation } from "react-i18next";

type OidcGroupsInfoProps = {
  groups: string[];
};

const OidcGroupsInfo = ({ groups }: OidcGroupsInfoProps) => {
  const { t } = useTranslation();
  return (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger>
          <Badge className="cursor-pointer" variant="outline">
            {t("pages.permissions.OidcGroupsInfo.title", {
              count: groups.length,
            })}
          </Badge>
        </TooltipTrigger>
        <TooltipContent className="flex flex-col max-w-xl flex-wrap gap-3 text-inherit">
          <Table>
            <TableBody>
              {groups.map((group) => (
                <Badge key={group} className="cursor-pointer" variant="outline">
                  {group}
                </Badge>
              ))}
            </TableBody>
          </Table>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
};

export default OidcGroupsInfo;
