import { Command, CommandGroup, CommandList } from "~/design/Command";

import { Datepicker } from "~/design/Datepicker";
import { FiltersObj } from "~/api/events/query/get";
import { useTranslation } from "react-i18next";

const DatePicker = ({
  field,
  date,
  setFilter,
}: {
  field: "AFTER" | "BEFORE";
  date?: Date;
  setFilter: (filter: FiltersObj) => void;
}) => {
  const setDate = (type: "AFTER" | "BEFORE", value: Date) => {
    setFilter({
      [type]: { type, value },
    });
  };

  const { t } = useTranslation();
  return (
    <Command>
      <CommandList className="max-h-[460px]">
        <CommandGroup heading={t(`components.filters.menuHeading.${field}`)}>
          <Datepicker
            mode="single"
            selected={date}
            onSelect={(value) => value && setDate(field, value)}
          />
        </CommandGroup>
      </CommandList>
    </Command>
  );
};

export default DatePicker;
