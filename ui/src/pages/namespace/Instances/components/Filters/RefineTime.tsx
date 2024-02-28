import { Command, CommandGroup, CommandList } from "~/design/Command";
import TimePicker, { getTimeString } from "~/design/Timepicker";

import { ArrowRight } from "lucide-react";
import Button from "~/design/Button";
import { FiltersObj } from "~/api/events/query/get";
import { useState } from "react";
import { useTranslation } from "react-i18next";

const RefineTime = ({
  field,
  date: givenDate,
  setFilter,
}: {
  field: "AFTER" | "BEFORE";
  date: Date;
  setFilter: (filter: FiltersObj) => void;
}) => {
  const { t } = useTranslation();

  const [date, setDate] = useState<Date>(givenDate ?? new Date());

  const time = getTimeString(date);

  const setTimeOnDate = () => {
    const [hr, min, sec] = time.split(":").map((item) => Number(item));

    if (hr === undefined || min == undefined || sec === undefined) {
      console.error("Invalid time string in setTimeOnDate");
      return;
    }

    givenDate.setHours(hr, min, sec);
    setFilter({
      [field]: { type: field, value: givenDate },
    });
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    e.key === "Enter" && setTimeOnDate();
  };

  return (
    <Command>
      <CommandList className="max-h-[460px]">
        <CommandGroup
          heading={t("pages.events.history.filter.menuHeading.time")}
        >
          <div className="flex items-center">
            <TimePicker
              onKeyDown={(e) => {
                handleKeyDown(e);
              }}
              date={date}
              setDate={setDate}
              hours={t("pages.events.history.filter.menuLabels.time.hours")}
              minutes={t("pages.events.history.filter.menuLabels.time.minutes")}
              seconds={t("pages.events.history.filter.menuLabels.time.seconds")}
            />

            <Button icon variant="ghost" onClick={() => setTimeOnDate()}>
              <ArrowRight />
            </Button>
          </div>
        </CommandGroup>
      </CommandList>
    </Command>
  );
};

export default RefineTime;
