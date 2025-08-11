import { Popover, PopoverContent, PopoverTrigger } from "~/design/Popover";

import Button from "~/design/Button";
import { CalendarIcon } from "lucide-react";
import { Datepicker } from "~/design/Datepicker";
import { InputProps } from "./types";
import moment from "moment";
import { parseStringToDate } from "./utils";
import { useState } from "react";
import { useTranslation } from "react-i18next";

type DateInputProps = InputProps;

export const DateInput = ({ id, defaultValue }: DateInputProps) => {
  const { t } = useTranslation();
  const [date, setDate] = useState<Date | undefined>(
    parseStringToDate(defaultValue)
  );
  return (
    <Popover>
      <PopoverTrigger asChild>
        <Button variant="outline">
          <CalendarIcon />{" "}
          {date
            ? moment(date).format("LL")
            : t("direktivPage.page.blocks.form.datePickerPlaceholder")}
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-auto">
        <Datepicker id={id} mode="single" selected={date} onSelect={setDate} />
      </PopoverContent>
    </Popover>
  );
};
