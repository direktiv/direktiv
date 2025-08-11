import { Popover, PopoverContent, PopoverTrigger } from "~/design/Popover";

import Button from "~/design/Button";
import { CalendarIcon } from "lucide-react";
import { Datepicker } from "~/design/Datepicker";
import { Fieldset } from "../utils/FieldSet";
import { FormDateInputType } from "../../../../schema/blocks/form/dateInput";
import moment from "moment";
import { parseStringToDate } from "./utils";
import { useState } from "react";
import { useTranslation } from "react-i18next";

type FormDateInputProps = {
  blockProps: FormDateInputType;
};

export const FormDateInput = ({ blockProps }: FormDateInputProps) => {
  const { t } = useTranslation();
  const { id, label, description, defaultValue, optional } = blockProps;
  const htmlID = `form-input-${id}`;

  const [date, setDate] = useState<Date | undefined>(
    parseStringToDate(defaultValue)
  );

  return (
    <Fieldset
      label={label}
      description={description}
      htmlFor={htmlID}
      optional={optional}
    >
      <Popover>
        <PopoverTrigger asChild>
          <Button variant="outline" className="w-[300px]">
            <CalendarIcon />{" "}
            {date
              ? moment(date).format("LL")
              : t("direktivPage.page.blocks.form.datePickerPlaceholder")}
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-auto">
          <Datepicker
            id={htmlID}
            mode="single"
            selected={date}
            onSelect={setDate}
          />
        </PopoverContent>
      </Popover>
    </Fieldset>
  );
};
