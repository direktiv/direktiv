import { ArrayInput } from "~/components/Form/ArrayInput";
import { ControllerRenderProps } from "react-hook-form";
import { HeaderManipulationFormSchemaType } from "../../../../schema/plugins/inbound/headerManipulation";
import Input from "~/design/Input";
import { useTranslation } from "react-i18next";

type HeaderArrayInputProps = {
  field:
    | ControllerRenderProps<
        HeaderManipulationFormSchemaType,
        "configuration.headers_to_add"
      >
    | ControllerRenderProps<
        HeaderManipulationFormSchemaType,
        "configuration.headers_to_modify"
      >
    | ControllerRenderProps<
        HeaderManipulationFormSchemaType,
        "configuration.headers_to_remove"
      >;
};

export const HeaderArrayInput = ({ field }: HeaderArrayInputProps) => {
  const isHeadersToRemove = field.name === "configuration.headers_to_remove";
  const { t } = useTranslation();
  return (
    <div className="grid gap-5">
      <ArrayInput
        defaultValue={field.value || []}
        onChange={(changedValue) => {
          field.onChange(changedValue);
        }}
        emptyItem={
          isHeadersToRemove
            ? { name: "" }
            : {
                name: "",
                value: "",
              }
        }
        itemIsValid={(item) => !!(item?.name && item?.value)}
        renderItem={({
          value: objectValue,
          setValue,
          onChange,
          handleKeyDown,
        }) => (
          <>
            {Object.entries(objectValue).map(([key, value]) => {
              const typedKey = key as keyof typeof objectValue;
              return (
                <Input
                  key={key}
                  data-testid={`env-${typedKey}`}
                  placeholder={t(
                    `pages.explorer.endpoint.editor.form.plugins.inbound.headerManipulation.${typedKey}Placeholder`
                  )}
                  value={value}
                  onKeyDown={handleKeyDown}
                  onChange={(e) => {
                    // TODO: better naming
                    const newObject = { [key]: e.target.value };
                    const newObjectValue = {
                      ...objectValue,
                      ...newObject,
                    };
                    setValue(newObjectValue);
                    onChange(newObjectValue);
                  }}
                />
              );
            })}
          </>
        )}
      />
    </div>
  );
};
