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
  const [open, setOpen] = useState(false);
  const [date, setDate] = useState<Date | undefined>(
    parseStringToDate(defaultValue)
  );

  return (
    <Popover open={open} onOpenChange={(open) => setOpen(open)}>
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
        <StopPropagation>
          <DatepickerDesignComponent
            id={id}
            mode="single"
            selected={date}
            onSelect={(date) => {
              setOpen(false);
              setDate(date);
            }}
          />
        </StopPropagation>
      </PopoverContent>
    </Popover>
  );
};
