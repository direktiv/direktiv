import { Popover, PopoverContent, PopoverTrigger } from "~/design/Popover";

import Button from "~/design/Button";
import { CalendarIcon } from "lucide-react";
import { Datepicker } from "~/design/Datepicker";
import { Fieldset } from "../utils/FieldSet";
import { FormDateInputType } from "../../../../schema/blocks/form/dateInput";
import moment from "moment";
import { parseStringToDate } from "./utils";
import { useState } from "react";
import { useTemplateStringResolver } from "../../../primitives/Variable/utils/useTemplateStringResolver";
import { useTranslation } from "react-i18next";

type FormDateInputProps = {
  blockProps: FormDateInputType;
};

// this formatter reflects the format of a native date input
const formatDateInput = (date: Date | undefined) =>
  date ? moment(date).format("YYYY-MM-DD") : "";

export const FormDateInput = ({ blockProps }: FormDateInputProps) => {
  const { t } = useTranslation();
  const { id, label, description, defaultValue, optional } = blockProps;
  const htmlID = `form-input-${id}`;

  const templateStringResolver = useTemplateStringResolver();
  const value = templateStringResolver(defaultValue);

  const [date, setDate] = useState<Date | undefined>(parseStringToDate(value));

  return (
    <Fieldset
      label={label}
      description={description}
      htmlFor={htmlID}
      optional={optional}
    >
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
          <Datepicker
            id={htmlID}
            mode="single"
            selected={date}
            onSelect={setDate}
          />
        </PopoverContent>
      </Popover>
      <input type="hidden" name={id} value={formatDateInput(date)} />
    </Fieldset>
  );
};
