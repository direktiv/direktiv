import {
  HoverCard,
  HoverCardContent,
  HoverCardTrigger,
} from "~/design/HoverCard";

import { ConditionalWrapper } from "~/util/helpers";
import { useTranslation } from "react-i18next";

const ContextFilters = ({ filters }: { filters: object[] }) => {
  const { t } = useTranslation();

  return (
    <ConditionalWrapper
      condition={filters.length > 0}
      wrapper={(children) => (
        <HoverCard>
          <HoverCardTrigger>{children}</HoverCardTrigger>
          <HoverCardContent
            align="center"
            side="left"
            className="whitespace-pre p-1"
          >
            {JSON.stringify(filters, null, "\t")}
          </HoverCardContent>
        </HoverCard>
      )}
    >
      <div>
        {t("pages.events.listeners.tableRow.contextFilters", {
          count: filters.length,
        })}
      </div>
    </ConditionalWrapper>
  );
};

export default ContextFilters;
