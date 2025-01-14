import {
  EventListenersPageSizeValueSchema,
  eventListenersPageSizeValue,
  useEventListenersPageSize,
  useEventListenersPageSizeActions,
} from "~/util/store/pagesizes/eventListeners";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "~/design/Select";

import { useTranslation } from "react-i18next";

export const SelectPageSize = ({ onChange }: { onChange: () => void }) => {
  const { t } = useTranslation();
  const { setEventListenersPageSize } = useEventListenersPageSizeActions();
  const pageSize = useEventListenersPageSize();

  return (
    <Select
      defaultValue={String(pageSize)}
      onValueChange={(value) => {
        const parseValue = EventListenersPageSizeValueSchema.safeParse(value);
        if (parseValue.success) {
          setEventListenersPageSize(parseValue.data);
          onChange();
        }
      }}
    >
      <SelectTrigger variant="outline">
        <SelectValue />
      </SelectTrigger>
      <SelectContent>
        {eventListenersPageSizeValue.map((size) => (
          <SelectItem key={size} value={size}>
            {t("pages.events.history.selectPageSize.selectItem", {
              count: Number(size),
            })}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  );
};
