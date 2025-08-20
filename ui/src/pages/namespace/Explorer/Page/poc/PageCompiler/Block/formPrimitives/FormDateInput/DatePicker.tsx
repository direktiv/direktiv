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
  fieldName: string;
};

// this formatter reflects the format of a native date input
const formatDateInput = (date: Date | undefined) =>
  date ? moment(date).format("YYYY-MM-DD") : "";

export const DatePicker = ({ defaultValue, fieldName }: DatePickerProps) => {
  const { t } = useTranslation();
  const [open, setOpen] = useState(false);
  const [date, setDate] = useState<Date | undefined>(
    parseStringToDate(defaultValue)
  );

  return (
    <>
      <Popover open={open} onOpenChange={setOpen}>
        <StopPropagation>
          <PopoverTrigger asChild>
            <Button variant="outline">
              <CalendarIcon />{" "}
              {date
                ? moment(date).format("LL")
                : t("direktivPage.page.blocks.form.datePickerPlaceholder")}
            </Button>
          </PopoverTrigger>
        </StopPropagation>
        <StopPropagation>
          <PopoverContent className="w-auto">
            <DatepickerDesignComponent
              id={fieldName}
              mode="single"
              selected={date}
              onSelect={(date) => {
                setOpen(false);
                setDate(date);
              }}
            />
          </PopoverContent>
        </StopPropagation>
      </Popover>
      <input type="hidden" name={fieldName} value={formatDateInput(date)} />
    </>
  );
};
