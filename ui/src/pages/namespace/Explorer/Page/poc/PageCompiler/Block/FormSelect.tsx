import { FormSelectType } from "../../schema/blocks/form/select";
import { TemplateString } from "../primitives/TemplateString";

type FormSelectProps = {
  blockProps: FormSelectType;
};

export const FormSelect = ({ blockProps }: FormSelectProps) => (
  <div>
    <label>
      <TemplateString value={blockProps.label} />
    </label>
    <select defaultValue={blockProps.defaultValue}>
      {blockProps.values.map((value) => (
        <option key={value} value={value}>
          {value}
        </option>
      ))}
    </select>
    <p>
      <TemplateString value={blockProps.description} />
    </p>
  </div>
);
