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
  const { id, label, description, defaultValue, values, required } = blockProps;
  const htmlID = `id-${id}`;
  return (
    <Fieldset
      label={label}
      description={description}
      htmlFor={htmlID}
      required={required}
    >
      <Select defaultValue={defaultValue}>
        <SelectTrigger variant="outline" id={htmlID}>
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
