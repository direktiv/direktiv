import { Popover, PopoverContent, PopoverTrigger } from "~/design/Popover";

import Button from "~/design/Button";
import { CalendarIcon } from "lucide-react";
import { Datepicker as DatepickerDesignComponent } from "~/design/Datepicker";
import { StopPropagation } from "~/components/StopPropagation";
import moment from "moment";
import { parseStringToDate } from "./utils";
import { useState } from "react";
import { useTranslation } from "react-i18next";

type DatePickerProps = {
  defaultValue: string;
  id: string;
};

export const DatePicker = ({ defaultValue, id }: DatePickerProps) => {
  const { t } = useTranslation();
  const [date, setDate] = useState<Date | undefined>(
    parseStringToDate(defaultValue)
  );

  return (
    <Popover>
      <StopPropagation asChild>
        <PopoverTrigger asChild>
          <Button variant="outline">
            <CalendarIcon />{" "}
            {date
              ? moment(date).format("LL")
              : t("direktivPage.page.blocks.form.datePickerPlaceholder")}
          </Button>
        </PopoverTrigger>
      </StopPropagation>
      <PopoverContent className="w-auto">
        <DatepickerDesignComponent
          id={id}
          mode="single"
          selected={date}
          onSelect={setDate}
        />
      </PopoverContent>
    </Popover>
  );
};
