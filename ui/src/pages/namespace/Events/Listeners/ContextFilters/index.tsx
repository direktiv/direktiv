import {
  HoverCard,
  HoverCardContent,
  HoverCardTrigger,
} from "~/design/HoverCard";
import { Table, TableBody } from "~/design/Table";

import { ConditionalWrapper } from "~/util/helpers";
import { EventContextFilterSchemaType } from "~/api/eventListeners/schema";
import FilterEntry from "./FilterEntry";
import { useTranslation } from "react-i18next";

const ContextFilters = ({
  filters,
}: {
  filters: EventContextFilterSchemaType[];
}) => {
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
            <Table>
              <TableBody>
                {filters.map((filter, index) => (
                  <FilterEntry key={index} filter={filter} />
                ))}
              </TableBody>
            </Table>
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
