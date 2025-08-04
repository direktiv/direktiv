import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "~/design/Select";

import { Fieldset } from "./utils/FieldSet";
import { FormSelectType } from "../../../schema/blocks/form/select";

type FormSelectProps = {
  blockProps: FormSelectType;
};

export const FormSelect = ({ blockProps }: FormSelectProps) => {
  const { id, label, description, defaultValue, values } = blockProps;
  return (
    <Fieldset label={label} description={description} htmlFor={id}>
      <Select defaultValue={defaultValue}>
        <SelectTrigger variant="outline">
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          {values.map((value) => (
            <SelectItem key={value} value={value}>
              {value}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    </Fieldset>
  );
};
