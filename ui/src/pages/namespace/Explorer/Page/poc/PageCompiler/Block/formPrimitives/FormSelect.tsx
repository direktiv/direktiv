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

export const FormSelect = ({ blockProps }: FormSelectProps) => (
  <Fieldset label={blockProps.label} description={blockProps.description}>
    <Select defaultValue={blockProps.defaultValue}>
      <SelectTrigger variant="outline">
        <SelectValue />
      </SelectTrigger>
      <SelectContent>
        {blockProps.values.map((value) => (
          <SelectItem key={value} value={value}>
            {value}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  </Fieldset>
);
