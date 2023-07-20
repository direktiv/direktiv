import { Command, CommandGroup, CommandList } from "~/design/Command";

import { BaseSyntheticEvent } from "react";
import { FiltersObj } from "~/api/instances/query/get";
import Input from "~/design/Input";
import moment from "moment";
import { useTranslation } from "react-i18next";

const RefineTime = ({
  field,
  date,
  setFilter,
}: {
  field: "AFTER" | "BEFORE";
  date: Date;
  setFilter: (filter: FiltersObj) => void;
}) => {
  const { t } = useTranslation();

  const setTimeOnDate = (event: BaseSyntheticEvent) => {
    const [hr, min, sec] = event.target.value.split(":");
    date.setHours(hr);
    date.setMinutes(min);
    date.setSeconds(sec);
    setFilter({
      [field]: { type: "MATCH", value: date },
    });
  };

  return (
    <Command>
      <CommandList className="max-h-[460px]">
        <CommandGroup
          heading={t("pages.instances.list.filter.menuHeading.time")}
        >
          <Input
            type="time"
            step={1}
            value={moment(date).format("HH:mm:ss")}
            onChange={(event) => {
              setTimeOnDate(event);
            }}
          />
        </CommandGroup>
      </CommandList>
    </Command>
  );
};

export default RefineTime;
