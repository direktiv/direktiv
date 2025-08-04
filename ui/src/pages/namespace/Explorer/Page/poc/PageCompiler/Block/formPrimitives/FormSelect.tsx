import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "~/design/Select";

import { FormSelectType } from "../../../schema/blocks/form/select";
import { TemplateString } from "../../primitives/TemplateString";

type FormSelectProps = {
  blockProps: FormSelectType;
};

export const FormSelect = ({ blockProps }: FormSelectProps) => (
  <div>
    <label>
      <TemplateString value={blockProps.label} />
    </label>
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
    <p>
      <TemplateString value={blockProps.description} />
    </p>
  </div>
);
