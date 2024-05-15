import {
  HoverCard,
  HoverCardContent,
  HoverCardTrigger,
} from "~/design/HoverCard";
import { Table, TableBody, TableCell, TableRow } from "~/design/Table";

import { ConditionalWrapper } from "~/util/helpers";
import { EventContextFiltersSchemaType } from "~/api/eventListeners/schema";
import { useTranslation } from "react-i18next";

const ContextFilters = ({
  filters,
}: {
  filters: EventContextFiltersSchemaType;
}) => {
  const { t } = useTranslation();

  const MappedFilters = filters.map((filter) => (
    <>
      <TableRow>
        <TableCell>{filter.type}</TableCell>
      </TableRow>
      {Object.entries(filter.context).map(([key, value]) => (
        <TableRow key={key}>
          <TableCell className="pl-10 text-gray-10">
            {key}: {value}
          </TableCell>
        </TableRow>
      ))}
    </>
  ));

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
              <TableBody>{MappedFilters}</TableBody>
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
