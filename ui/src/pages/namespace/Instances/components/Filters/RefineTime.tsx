import { Command, CommandGroup, CommandList } from "~/design/Command";
import TimePicker, { getTimeString } from "~/design/Timepicker";

import { ArrowRight } from "lucide-react";
import Button from "~/design/Button";
import { FiltersObj } from "~/api/events/query/get";
import moment from "moment";
import { useState } from "react";
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
  const [time, setTime] = useState<string>(getTimeString(date));

  const [dateNew, setDate] = useState<Date>(date ?? new Date());

  const setTimeOnDate = () => {
    const [hr, min, sec] = time.split(":").map((item) => Number(item));

    if (hr === undefined || min == undefined || sec === undefined) {
      console.error("Invalid time string in setTimeOnDate");
      return;
    }

    date.setHours(hr, min, sec);
    setFilter({
      [field]: { type: field, value: date },
    });
  };

  const handleKeyDown = (event: { key: string }) => {
    event.key === "Enter" && setTimeOnDate();
  };

  return (
    <Command>
      <CommandList className="max-h-[460px]">
        <CommandGroup
          heading={t("pages.events.history.filter.menuHeading.time")}
        >
          <div className="flex items-end">
            <TimePicker
              onTimeChange={(time) => {
                setTime(time);
                handleKeyDown;
              }}
              time={time}
              setTime={(time) => setTime(time)}
              date={dateNew}
              setDate={setDate}
              hours="Hours"
              minutes="Minutes"
              seconds="Seconds"
              onKeyDown={() => {
                handleKeyDown;
              }}
            />

            <Button
              className="mb-2"
              icon
              variant="ghost"
              onClick={() => setTimeOnDate()}
            >
              <ArrowRight />
            </Button>
          </div>
        </CommandGroup>
      </CommandList>
    </Command>
  );
};

export default RefineTime;
