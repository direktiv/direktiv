import { Command, CommandGroup, CommandList } from "~/design/Command";

import { ArrowRight } from "lucide-react";
import Button from "~/design/Button";
import { FiltersObj } from "~/api/instances/query/get";
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

  const setTimeOnDate = () => {
    const [hr, min, sec] = time.split(":").map((item) => Number(item));

    if (hr === undefined || min == undefined || sec === undefined) {
      console.error("Invalid time string in setTimeOnDate");
      return;
    }

    date.setHours(hr, min, sec);
    setFilter({
      [field]: { type: "MATCH", value: date },
    });
  };

  const handleKeyDown = (event: { key: string }) => {
    event.key === "Enter" && setTimeOnDate();
  };

  return (
    <Command>
      <CommandList className="max-h-[460px]">
        <CommandGroup
          heading={t("pages.instances.list.filter.menuHeading.time")}
        >
          <InputWithButton>
            <Input
              type="time"
              step={1}
              value={time}
              onChange={(event) => setTime(event.target.value)}
              onKeyDown={handleKeyDown}
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
