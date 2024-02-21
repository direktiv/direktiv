import { Command, CommandGroup, CommandList } from "~/design/Command";
import TimePicker, { getTimeString } from "~/design/Timepicker";

import { ArrowRight } from "lucide-react";
import Button from "~/design/Button";
import { FiltersObj } from "~/api/events/query/get";
import Input from "~/design/Input";
import { InputWithButton } from "~/design/InputWithButton";
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
  const [time, setTime] = useState<string>(moment(date).format("HH:mm:ss"));

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
    console.log("upperKEY");
    event.key === "Enter" && setTimeOnDate();
  };

  return (
    <Command>
      <CommandList className="max-h-[460px]">
        <CommandGroup
          heading={t("pages.events.history.filter.menuHeading.time")}
        >
          <TimePicker
            onTimeChange={(time) => {
              setTime(time);
              console.log("here");
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
              console.log("There");
            }}
          />

          <Button icon variant="ghost" onClick={() => setTimeOnDate()}>
            <ArrowRight />
          </Button>
          <InputWithButton>
            <Input
              type="time"
              step={1}
              value={time}
              onChange={(event) => setTime(event.target.value)}
              onKeyDown={() => {
                handleKeyDown;
                console.log("XThere");
              }}
            />
            <Button icon variant="ghost" onClick={() => setTimeOnDate()}>
              <ArrowRight />
            </Button>
          </InputWithButton>
        </CommandGroup>
      </CommandList>
    </Command>
  );
};

export default RefineTime;
