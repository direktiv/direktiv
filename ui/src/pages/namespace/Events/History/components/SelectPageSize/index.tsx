import {
  EventsPageSizeValueSchema,
  eventsPageSizeValue,
  useEventsPageSize,
  useEventsPageSizeActions,
} from "~/util/store/events";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "~/design/Select";

import { useTranslation } from "react-i18next";

export const SelectPageSize = ({
  goToPage,
}: {
  goToPage: (page: number) => void;
}) => {
  const { t } = useTranslation();
  const { setEventsPageSize } = useEventsPageSizeActions();
  const pageSize = useEventsPageSize();

  return (
    <Select
      defaultValue={String(pageSize)}
      onValueChange={(value) => {
        const parseValue = EventsPageSizeValueSchema.safeParse(value);
        if (parseValue.success) {
          setEventsPageSize(parseValue.data);
          goToPage(1);
        }
      }}
    >
      <SelectTrigger variant="outline">
        <SelectValue />
      </SelectTrigger>
      <SelectContent>
        {eventsPageSizeValue.map((size) => (
          <SelectItem key={size} value={size}>
            {t("pages.events.history.selectPageSize.selectItem", {
              pagesize: size,
            })}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  );
};
