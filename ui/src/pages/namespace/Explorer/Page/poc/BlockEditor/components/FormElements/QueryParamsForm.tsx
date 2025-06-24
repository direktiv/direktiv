import { ArrayForm } from "~/components/Form/Array";
import { ControllerRenderProps } from "react-hook-form";
import { Fieldset } from "~/components/Form/Fieldset";
import Input from "~/design/Input";
import { QueryType } from "../../../schema/procedures/query";

type KeyValueInputProps = {
  label: string;
  field: ControllerRenderProps<QueryType, "queryParams">;
};

export const KeyValueInput = ({ field, label }: KeyValueInputProps) => (
  <Fieldset label={label}>
    <ArrayForm
      defaultValue={field.value || []}
      onChange={field.onChange}
      emptyItem={{ key: "", value: "" }}
      itemIsValid={(item) =>
        !!item && Object.values(item).every((v) => v !== "")
      }
      renderItem={({ value: objectValue, setValue, handleKeyDown }) => (
        <>
          {Object.entries(objectValue).map(([key, value]) => (
            <Input
              key={key}
              placeholder={key}
              value={value}
              onKeyDown={handleKeyDown}
              onChange={(e) => {
                const newObject = {
                  ...objectValue,
                  [key]: e.target.value,
                };
                setValue(newObject);
              }}
            />
          ))}
        </>
      )}
    />
  </Fieldset>
);
