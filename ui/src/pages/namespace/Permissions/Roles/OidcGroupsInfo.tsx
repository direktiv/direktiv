import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "~/design/Tooltip";

import Badge from "~/design/Badge";
import { ConditionalWrapper } from "~/util/helpers";
import { useTranslation } from "react-i18next";

type OidcGroupsInfoProps = {
  groups: string[];
};

const OidcGroupsInfo = ({ groups }: OidcGroupsInfoProps) => {
  const { t } = useTranslation();
  return (
    <ConditionalWrapper
      condition={groups.length > 0}
      wrapper={(children) => (
        <TooltipProvider>
          <Tooltip>
            <TooltipTrigger>{children}</TooltipTrigger>
            <TooltipContent className="flex max-w-xl flex-row flex-wrap gap-3 text-inherit">
              {groups.map((group) => (
                <Badge key={group} className="cursor-pointer" variant="outline">
                  {group}
                </Badge>
              ))}
            </TooltipContent>
          </Tooltip>
        </TooltipProvider>
      )}
    >
      <Badge variant="outline">
        {t("pages.permissions.OidcGroupsInfo.title", {
          count: groups.length,
        })}
      </Badge>
    </ConditionalWrapper>
  );
};

export default OidcGroupsInfo;
